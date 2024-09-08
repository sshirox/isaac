package main

import (
	"net/http"
	"slices"
	"strconv"
)

const (
	GaugeMetricType   = "gauge"
	CounterMetricType = "counter"
)

var (
	metricTypes = []string{GaugeMetricType, CounterMetricType}
)

func metricsHandler() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		metricType := r.PathValue("metric_type")
		metricName := r.PathValue("metric_name")
		metricValue := r.PathValue("metric_value")

		rw.Header().Set("Content-Type", "text/plain; charset=utf-8")

		if !slices.Contains(metricTypes, metricType) {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte(`""`))
			return
		}

		if len(metricName) == 0 {
			rw.WriteHeader(http.StatusNotFound)
			rw.Write([]byte(`""`))
			return
		}

		switch metricType {
		case GaugeMetricType:
			if _, err := strconv.ParseFloat(metricValue, 64); err != nil {
				rw.WriteHeader(http.StatusBadRequest)
				rw.Write([]byte(`""`))
				return
			}
		case CounterMetricType:
			if _, err := strconv.ParseInt(metricValue, 10, 64); err != nil {
				rw.WriteHeader(http.StatusBadRequest)
				rw.Write([]byte(`""`))
				return
			}
		}

		rw.WriteHeader(http.StatusOK)
		rw.Write([]byte(`""`))
	}
}

func main() {
	m := http.NewServeMux()

	m.Handle("POST /update/{metric_type}/{metric_name}/{metric_value}", metricsHandler())

	err := http.ListenAndServe(":8081", m)

	if err != nil {
		panic(err)
	}
}
