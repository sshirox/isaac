package handler

import (
	"encoding/json"
	"github.com/sshirox/isaac/internal/model"
	"github.com/sshirox/isaac/internal/usecase"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"slices"
	"strconv"

	"github.com/go-chi/chi/v5"
)

const (
	GaugeMetricType   = "gauge"
	CounterMetricType = "counter"
)

var (
	metricTypes = []string{GaugeMetricType, CounterMetricType}
)

func UpdateMetricsHandler(uc *usecase.UseCase) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		metricType := chi.URLParam(r, "metric_type")
		metricName := chi.URLParam(r, "metric_name")
		metricValue := chi.URLParam(r, "metric_value")

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

		_ = uc.UpsertMetric(metricType, metricName, metricValue)

		rw.WriteHeader(http.StatusOK)
		rw.Write([]byte(`""`))
	}
}

func GetMetricHandler(uc *usecase.UseCase) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		metricType := chi.URLParam(r, "metric_type")
		metricName := chi.URLParam(r, "metric_name")

		rw.Header().Set("Content-Type", "text/plain; charset=utf-8")

		val, err := uc.GetMetric(metricType, metricName)

		if err != nil {
			rw.WriteHeader(http.StatusNotFound)
			rw.Write([]byte(`""`))
			return
		}

		rw.WriteHeader(http.StatusOK)
		rw.Write([]byte(val))
	}
}

func UpdateMetricsJSONHandler(uc *usecase.UseCase) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}
		var m model.Metric
		err = json.Unmarshal(body, &m)
		if err != nil {
			panic(err)
		}

		rw.Header().Set("Content-Type", "application/json")

		if !slices.Contains(metricTypes, m.MType) {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte(`""`))
			return
		}

		if len(m.ID) == 0 {
			rw.WriteHeader(http.StatusNotFound)
			rw.Write([]byte(`""`))
			return
		}

		updatedMetric := uc.UpdateMetric(m.MType, m.ID, m.Value, m.Delta)

		jsonResp, err := json.Marshal(updatedMetric)
		if err != nil {
			slog.Error("marshal response", "err", err)
		}

		rw.WriteHeader(http.StatusOK)
		_, err = rw.Write(jsonResp)
		if err != nil {
			slog.Error("write response", "err", err)
		}
	}
}

func GetMetricsJSONHandler(uc *usecase.UseCase) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}
		var m model.Metric
		err = json.Unmarshal(body, &m)
		if err != nil {
			panic(err)
		}

		res, err := uc.ReceiveMetric(m.MType, m.ID)

		rw.Header().Set("Content-Type", "application/json")

		if err != nil {
			rw.WriteHeader(http.StatusNotFound)
			rw.Write([]byte(`""`))
			return
		}

		jsonResp, err := json.Marshal(res)
		if err != nil {
			slog.Error("marshal response", "err", err)
		}

		rw.WriteHeader(http.StatusOK)
		_, err = rw.Write(jsonResp)
		if err != nil {
			slog.Error("write response", "err", err)
		}
	}
}

func IndexHandler(uc *usecase.UseCase) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		var tpl = template.Must(template.ParseFiles("templates/index.html"))

		rw.WriteHeader(http.StatusOK)
		tpl.Execute(rw, uc.GetAllMetrics())
	}
}
