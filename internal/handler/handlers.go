package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/sshirox/isaac/internal/metric"
)

type Repository interface {
	UpdateGauge(string, float64)
	UpdateCounter(string, int64)
	ReceiveGauge(string) (float64, bool)
	ReceiveCounter(string) (int64, bool)
	ReceiveAllGauges() map[string]float64
	ReceiveAllCounters() map[string]int64
}

func UpdateMetricsHandler(repo Repository) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		updateMetrics(repo, rw, r)
	}
}

func updateMetrics(repo Repository, rw http.ResponseWriter, r *http.Request) {
	mType := chi.URLParam(r, "type")
	name := chi.URLParam(r, "name")
	value := chi.URLParam(r, "value")

	rw.Header().Set("Content-Type", "text/plain; charset=utf-8")

	if len(name) == 0 {
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte("empty metric name"))
		return
	} else if len(value) == 0 {
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte("empty metric value"))
		return
	}

	switch mType {
	case metric.GaugeMetricType:
		val, err := strconv.ParseFloat(value, 64)
		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte("metric value is not a float"))
			return
		}
		repo.UpdateGauge(name, val)
		rw.WriteHeader(http.StatusOK)
		rw.Write([]byte("gauge successfully updated"))
	case metric.CounterMetricType:
		val, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte("metric value is not a integer"))
			return
		}
		repo.UpdateCounter(name, val)
		rw.WriteHeader(http.StatusOK)
		rw.Write([]byte("counter successfully updated"))
	default:
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte("invalid metric type"))
		return
	}
}

func ValueMetricHandler(repo Repository) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		valueMetric(repo, rw, r)
	}
}

func valueMetric(repo Repository, rw http.ResponseWriter, r *http.Request) {
	mType := chi.URLParam(r, "type")
	name := chi.URLParam(r, "name")

	rw.Header().Set("Content-Type", "text/plain; charset=utf-8")

	switch mType {
	case metric.GaugeMetricType:
		val, ok := repo.ReceiveGauge(name)
		if !ok {
			rw.WriteHeader(http.StatusNotFound)
			rw.Write([]byte("metric not found"))
			return
		}
		rw.WriteHeader(http.StatusOK)
		rw.Write([]byte(strconv.FormatFloat(val, 'g', -1, 64)))
	case metric.CounterMetricType:
		val, ok := repo.ReceiveCounter(name)
		if !ok {
			rw.WriteHeader(http.StatusNotFound)
			rw.Write([]byte("metric not found"))
			return
		}
		rw.WriteHeader(http.StatusOK)
		rw.Write([]byte(fmt.Sprintf("%d", val)))
	default:
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte("invalid metric type"))
	}
}

func UpdateByContentTypeHandler(repo Repository) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		contentType := r.Header.Get("Content-type")
		if contentType == "application/json" {
			rw.Header().Set("Content-Type", "application/json; charset=utf-8")

			body, err := io.ReadAll(r.Body)
			if err != nil {
				rw.WriteHeader(http.StatusBadRequest)
				rw.Write([]byte("invalid body"))
				return
			}
			var m metric.Metrics
			err = json.Unmarshal(body, &m)
			if err != nil {
				rw.WriteHeader(http.StatusBadRequest)
				rw.Write([]byte("invalid json body"))
				return
			}

			switch m.MType {
			case metric.GaugeMetricType:
				id, value := m.ID, m.Value
				if value == nil {
					rw.WriteHeader(http.StatusBadRequest)
					rw.Write([]byte("empty value"))
					return
				}
				repo.UpdateGauge(id, *value)
				newVal, _ := repo.ReceiveGauge(id)
				m.Value = &newVal

				rw.WriteHeader(http.StatusOK)
				json.NewEncoder(rw).Encode(m)
			case metric.CounterMetricType:
				id, delta := m.ID, m.Delta
				if delta == nil {
					rw.WriteHeader(http.StatusBadRequest)
					rw.Write([]byte("empty delta"))
					return
				}
				repo.UpdateCounter(id, *delta)
				newDelta, _ := repo.ReceiveCounter(id)
				m.Delta = &newDelta

				rw.WriteHeader(http.StatusOK)
				json.NewEncoder(rw).Encode(m)
			default:
				rw.WriteHeader(http.StatusBadRequest)
				rw.Write([]byte("invalid metric type"))
			}
		} else {
			updateMetrics(repo, rw, r)
		}
	}
}

func BulkUpdateHandler(repo Repository) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "application/json; charset=utf-8")

		body, err := io.ReadAll(r.Body)
		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte("invalid body"))
			return
		}
		var metrics []metric.Metrics
		err = json.Unmarshal(body, &metrics)
		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte("invalid json body"))
			return
		}

		for _, m := range metrics {
			switch m.MType {
			case metric.GaugeMetricType:
				id, value := m.ID, m.Value
				if value == nil {
					rw.WriteHeader(http.StatusBadRequest)
					rw.Write([]byte("empty value"))
					return
				}
				repo.UpdateGauge(id, *value)
			case metric.CounterMetricType:
				id, delta := m.ID, m.Delta
				if delta == nil {
					rw.WriteHeader(http.StatusBadRequest)
					rw.Write([]byte("empty delta"))
					return
				}
				repo.UpdateCounter(id, *delta)
			default:
				rw.WriteHeader(http.StatusBadRequest)
				rw.Write([]byte("invalid metric type"))
				return
			}
		}

		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(metrics)
	}
}

func ValueByContentTypeHandler(repo Repository) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		contentType := r.Header.Get("Content-type")
		if contentType == "application/json" {
			rw.Header().Set("Content-Type", "application/json; charset=utf-8")

			body, err := io.ReadAll(r.Body)
			if err != nil {
				rw.WriteHeader(http.StatusBadRequest)
				rw.Write([]byte("invalid body"))
				return
			}
			var m metric.Metrics
			err = json.Unmarshal(body, &m)
			if err != nil {
				rw.WriteHeader(http.StatusBadRequest)
				rw.Write([]byte("invalid json body"))
				return
			}

			id := m.ID
			switch m.MType {
			case metric.GaugeMetricType:
				value, ok := repo.ReceiveGauge(id)
				if !ok {
					rw.WriteHeader(http.StatusNotFound)
					rw.Write([]byte("metric not found"))
					return
				}
				m.Value = &value

				rw.WriteHeader(http.StatusOK)
				json.NewEncoder(rw).Encode(m)
			case metric.CounterMetricType:
				delta, ok := repo.ReceiveCounter(id)
				if !ok {
					rw.WriteHeader(http.StatusNotFound)
					rw.Write([]byte("metric not found"))
					return
				}
				m.Delta = &delta

				rw.WriteHeader(http.StatusOK)
				json.NewEncoder(rw).Encode(m)
			default:
				rw.WriteHeader(http.StatusBadRequest)
				rw.Write([]byte("invalid metric type"))
			}
		} else {
			valueMetric(repo, rw, r)
		}
	}
}

func IndexHandler(repo Repository) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		var tpl = template.Must(template.ParseFiles("templates/index.html"))
		type metrics struct {
			Gauges   map[string]float64
			Counters map[string]int64
		}
		m := &metrics{
			Gauges:   repo.ReceiveAllGauges(),
			Counters: repo.ReceiveAllCounters(),
		}

		rw.Header().Set("Content-Type", "text/html; charset=utf-8")
		rw.WriteHeader(http.StatusOK)
		err := tpl.Execute(rw, m)
		if err != nil {
			panic(err)
		}
	}
}

func PingDBHandler(db *sql.DB) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "text/html; charset=utf-8")

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		if err := db.PingContext(ctx); err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Write([]byte(err.Error()))

			cancel()

			return
		}

		rw.WriteHeader(http.StatusOK)
		rw.Write([]byte("success ping"))
	}
}
