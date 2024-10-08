package server

import (
	"database/sql"
	_ "database/sql"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/sshirox/isaac/internal/backup"
	"github.com/sshirox/isaac/internal/handler"
	"github.com/sshirox/isaac/internal/logger"
	"github.com/sshirox/isaac/internal/middleware"
	"github.com/sshirox/isaac/internal/storage"
	"go.uber.org/zap"
	"log/slog"
	"net/http"
	"os"
	"strconv"
)

func Run() error {
	parseFlags()
	err := initConf()

	if err != nil {
		return err
	}

	db, err := sql.Open("pgx", flagDatabaseDSN)
	if err != nil {
		slog.Error("open database", "error", err)
	}
	defer db.Close()

	r := chi.NewRouter()
	r.Use(chimiddleware.Recoverer)
	r.Use(logger.WithLogging)
	r.Use(middleware.GZipMiddleware)
	s := storage.NewMemStorage()

	r.Get("/", handler.IndexHandler(s))
	r.Route("/update", func(r chi.Router) {
		r.Post("/", handler.UpdateByContentTypeHandler(s))
		r.Post("/{type}/{name}/{value}", handler.UpdateMetricsHandler(s))
	})
	r.Route("/value", func(r chi.Router) {
		r.Post("/", handler.ValueByContentTypeHandler(s))
		r.Get("/{type}/{name}", handler.ValueByContentTypeHandler(s))
	})
	r.Get("/ping", handler.PingDBHandler(db))

	s, f, err := backup.RestoreMetrics(s, flagFileStoragePath, flagRestore)
	if err != nil {
		return err
	}
	defer f.Close()

	go backup.RunWorker(s, flagStoreInterval, f, make(chan struct{}))

	logger.Log.Info("Running server", zap.String("address", flagRunAddr))

	err = http.ListenAndServe(flagRunAddr, r)

	if err != nil {
		return err
	}

	return nil
}

func initConf() error {
	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		flagRunAddr = envRunAddr
	}

	if envLogLevel := os.Getenv("LOG_LEVEL"); envLogLevel != "" {
		flagLogLevel = envLogLevel
	}

	envStoreInterval := os.Getenv("STORE_INTERVAL")
	if envStoreInterval != "" {
		storeInterval, err := strconv.ParseInt(envStoreInterval, 10, 64)
		if err != nil {
			slog.Error("parse store interval", "error", err)
		}
		flagStoreInterval = storeInterval
	}

	if envFileStoragePath := os.Getenv("FILE_STORAGE_PATH"); envFileStoragePath != "" {
		flagFileStoragePath = envFileStoragePath
	}

	if envRestore := os.Getenv("RESTORE"); envRestore != "" {
		flagRestoreStr = envRestore
	}

	var err error
	flagRestore, err = strconv.ParseBool(flagRestoreStr)
	if err != nil {
		slog.Error("parse restore", "error", err)
		os.Exit(1)
	}

	if envDatabaseDSN := os.Getenv("DATABASE_DSN"); envDatabaseDSN != "" {
		flagDatabaseDSN = envDatabaseDSN
	}

	if err = logger.Initialize(flagLogLevel); err != nil {
		return err
	}

	return nil
}
