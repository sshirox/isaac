package main

import (
	"flag"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sshirox/isaac/internal/handler"
	"github.com/sshirox/isaac/internal/storage"
)

var flagRunAddr string

func main() {
	parseFlags()

	if err := run(); err != nil {
		panic(err)
	}
}

func parseFlags() {
	flag.StringVar(&flagRunAddr, "a", ":8080", "address and port to run server")
	flag.Parse()
}

func run() error {
	r := chi.NewRouter()
	gaugeStore := make(map[string]string)
	counterStore := make(map[string]string)
	ms, err := storage.NewMemStorage(gaugeStore, counterStore)

	if err != nil {
		return err
	}

	r.Post("/update/{metric_type}/{metric_name}/{metric_value}", handler.UpdateMetricsHandler(ms))
	r.Get("/value/{metric_type}/{metric_name}", handler.GetMetricHandler(ms))
	r.Get("/", handler.IndexHandler(ms))

	err = http.ListenAndServe(flagRunAddr, r)

	if err != nil {
		return err
	}

	return nil
}
