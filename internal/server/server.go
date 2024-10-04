package server

import (
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/sshirox/isaac/internal/handler"
	"github.com/sshirox/isaac/internal/logger"
	"github.com/sshirox/isaac/internal/middleware"
	"github.com/sshirox/isaac/internal/storage"
	"go.uber.org/zap"
	"net/http"
	"os"
)

func Run() error {
	parseFlags()
	err := initConf()

	if err != nil {
		return err
	}

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

	if err := logger.Initialize(flagLogLevel); err != nil {
		return err
	}

	return nil
}
