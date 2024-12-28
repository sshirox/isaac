package server

import (
	"context"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/sshirox/isaac/internal/backup"
	"github.com/sshirox/isaac/internal/crypto"
	"github.com/sshirox/isaac/internal/handler"
	"github.com/sshirox/isaac/internal/logger"
	"github.com/sshirox/isaac/internal/middleware"
	"github.com/sshirox/isaac/internal/storage"
	"github.com/sshirox/isaac/internal/storage/pg"
	"log"
	"log/slog"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strconv"
)

const (
	dbStorageSource     = "database"
	fileStorageSource   = "file"
	memoryStorageSource = "memory"
	dbDriver            = "postgres"
)

var (
	storageSource string
)

func Run() error {
	parseFlags()
	err := initConf()

	if err != nil {
		return err
	}

	db, err := pg.Open(dbDriver, flagDatabaseDSN)
	if err != nil {
		slog.Error("open database", "err", err)
	}
	defer db.Close()

	err = pg.Ping(db)
	if err != nil {
		slog.Error("ping database", "err", err)
	} else {
		slog.Info("open database", "addr", flagDatabaseDSN)
	}
	encoder := crypto.NewEncoder(flagEncryptionKey)
	signValidator := middleware.NewSignValidator(encoder).Validate

	r := chi.NewRouter()
	r.Use(chimiddleware.Recoverer)
	r.Use(logger.WithLogging)
	r.Use(middleware.GZipMiddleware)
	s := storage.NewMemStorage()

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	r.Get("/", handler.IndexHandler(s))
	r.Route("/update", func(r chi.Router) {
		r.Post("/", handler.UpdateByContentTypeHandler(s))
		r.Post("/{type}/{name}/{value}", handler.UpdateMetricsHandler(s))
	})
	r.Route("/updates", func(r chi.Router) {
		r.With(signValidator).Post("/", handler.BulkUpdateHandler(s))
	})
	r.Route("/value", func(r chi.Router) {
		r.Post("/", handler.ValueByContentTypeHandler(s))
		r.Get("/{type}/{name}", handler.ValueByContentTypeHandler(s))
	})
	r.Get("/ping", handler.PingDBHandler(db))

	switch storageSource {
	case fileStorageSource:
		ms, f, err := backup.RestoreMetrics(s, flagFileStoragePath, flagRestore)
		if err != nil {
			return err
		}
		defer f.Close()

		slog.Info("File is used as storage")

		go backup.RunWorker(ms, flagStoreInterval, f, make(chan struct{}))
	case dbStorageSource:
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		err = pg.Bootstrap(db, ctx)
		if err != nil {
			slog.Error("bootstrap database", "err", err)
			return err
		}

		err = pg.ListMetrics(db, s)
		if err != nil {
			return err
		}
		slog.Info("Database is used as a storage")

		go pg.RunSaver(db, s, flagStoreInterval, make(chan struct{}))
	}

	slog.Info("Running server", "address", flagRunAddr)

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
			slog.Error("parse store interval", "err", err)
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
		slog.Error("parse restore", "err", err)
		os.Exit(1)
	}

	if envDatabaseDSN := os.Getenv("DATABASE_DSN"); envDatabaseDSN != "" {
		flagDatabaseDSN = envDatabaseDSN
	}

	if envEncryptionKey := os.Getenv("KEY"); envEncryptionKey != "" {
		flagEncryptionKey = envEncryptionKey
	}

	if flagDatabaseDSN != "" {
		storageSource = dbStorageSource
	} else if flagFileStoragePath != "" {
		storageSource = fileStorageSource
	} else {
		storageSource = memoryStorageSource
	}

	if err = logger.Initialize(flagLogLevel); err != nil {
		return err
	}

	return nil
}
