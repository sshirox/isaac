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

		updatedMetric := uc.UpsertMetric(m.MType, m.ID, m.Value, m.Delta)

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

func GetMetricHandler(uc *usecase.UseCase) http.HandlerFunc {
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

		res, err := uc.GetMetric(m.MType, m.ID)

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
