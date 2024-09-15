package server

import (
	"flag"
	"github.com/sshirox/isaac/internal/usecase"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/sshirox/isaac/internal/handler"
	"github.com/sshirox/isaac/internal/storage"
)

var flagRunAddr string

func parseFlags() {
	flag.StringVar(&flagRunAddr, "a", ":8080", "address and port to run server")
	flag.Parse()
}

func Run() error {
	parseFlags()

	r := chi.NewRouter()
	gaugeStore := make(map[string]string)
	counterStore := make(map[string]string)
	ms, err := storage.NewMemStorage(gaugeStore, counterStore)
	uc := usecase.New(ms)
	runAddr := os.Getenv("ADDRESS")

	if runAddr == "" {
		runAddr = flagRunAddr
	}

	if err != nil {
		return err
	}

	r.Post("/update/{metric_type}/{metric_name}/{metric_value}", handler.UpdateMetricsHandler(uc))
	r.Get("/value/{metric_type}/{metric_name}", handler.GetMetricHandler(uc))
	r.Get("/", handler.IndexHandler(uc))

	err = http.ListenAndServe(runAddr, r)

	if err != nil {
		return err
	}

	return nil
}
