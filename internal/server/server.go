package server

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sshirox/isaac/internal/compress"
	"github.com/sshirox/isaac/internal/handler"
	"github.com/sshirox/isaac/internal/logger"
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
	r.Use(middleware.Recoverer)
	s := storage.NewMemStorage()

	r.Get("/", logger.WithLogging(handler.IndexHandler(s)))
	r.Route("/update", func(r chi.Router) {
		r.Post("/", compress.GzipMiddleware(logger.WithLogging(handler.UpdateByContentTypeHandler(s))))
		r.Post("/{type}/{name}/{value}", compress.GzipMiddleware(logger.WithLogging(handler.UpdateMetricsHandler(s))))
	})
	r.Route("/value", func(r chi.Router) {
		r.Post("/", compress.GzipMiddleware(logger.WithLogging(handler.ValueByContentTypeHandler(s))))
		r.Get("/{type}/{name}", compress.GzipMiddleware(logger.WithLogging(handler.ValueByContentTypeHandler(s))))
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
