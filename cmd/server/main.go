package main

import (
	"net/http"

	"github.com/sshirox/isaac/internal/handler"
)

func main() {
	m := http.NewServeMux()

	m.Handle("POST /update/{metric_type}/{metric_name}/{metric_value}", handler.MetricsHandler())

	err := http.ListenAndServe(":8080", m)

	if err != nil {
		panic(err)
	}
}
