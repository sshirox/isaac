package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sshirox/isaac/internal/handler"
	"github.com/sshirox/isaac/internal/storage"
)

func main() {
	r := chi.NewRouter()
	gaugeStore := make(map[string]string)
	counterStore := make(map[string]string)
	ms, err := storage.NewMemStorage(gaugeStore, counterStore)

	if err != nil {
		panic(err)
	}

	r.Post("/update/{metric_type}/{metric_name}/{metric_value}", handler.UpdateMetricsHandler(ms))
	r.Get("/value/{metric_type}/{metric_name}", handler.GetMetricHandler(ms))
	r.Get("/", handler.IndexHandler(ms))

	err = http.ListenAndServe(":8080", r)

	if err != nil {
		panic(err)
	}
}
