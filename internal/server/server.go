package server

import (
	"flag"
	"github.com/sshirox/isaac/internal/logger"
	"github.com/sshirox/isaac/internal/usecase"
	"go.uber.org/zap"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/sshirox/isaac/internal/handler"
	"github.com/sshirox/isaac/internal/storage"
)

var (
	flagRunAddr  string
	flagLogLevel string
)

func parseFlags() {
	flag.StringVar(&flagRunAddr, "a", ":8080", "address and port to run server")
	flag.StringVar(&flagLogLevel, "l", "info", "log level")
	flag.Parse()
}

func Run() error {
	parseFlags()

	r := chi.NewRouter()
	gaugeStore := make(map[string]string)
	counterStore := make(map[string]string)
	ms, err := storage.NewMemStorage(gaugeStore, counterStore)

	if err != nil {
		return err
	}

	uc := usecase.New(ms)

	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		flagRunAddr = envRunAddr
	}

	if envLogLevel := os.Getenv("LOG_LEVEL"); envLogLevel != "" {
		flagLogLevel = envLogLevel
	}

	if err = logger.Initialize(flagLogLevel); err != nil {
		return err
	}

	r.Post("/update/{metric_type}/{metric_name}/{metric_value}", logger.WithLogging(handler.UpdateMetricsHandler(uc)))
	r.Get("/value/{metric_type}/{metric_name}", logger.WithLogging(handler.GetMetricHandler(uc)))
	r.Get("/", logger.WithLogging(handler.IndexHandler(uc)))

	logger.Log.Info("Running server", zap.String("address", flagRunAddr))

	err = http.ListenAndServe(flagRunAddr, r)

	if err != nil {
		return err
	}

	return nil
}
