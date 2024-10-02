package handler

import (
	"bytes"
	"encoding/json"
	"github.com/sshirox/isaac/internal/metric"
	"github.com/sshirox/isaac/internal/tests/mocks/mockstorage"
	"github.com/sshirox/isaac/internal/usecase"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestUpdateMetricsHandler(t *testing.T) {
	type want struct {
		contentType string
		statusCode  int
	}
	testCases := []struct {
		name    string
		request string
		want    want
	}{
		{
			name: "Valid metric",
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  200,
			},
			request: "/update/gauge/someMetric/789",
		},
		{
			name: "Invalid metric type",
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  400,
			},
			request: "/update/metric/someMetric/789",
		},
		{
			name: "Empty metric name",
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  404,
			},
			request: "/update/gauge/789",
		},
		{
			name: "Invalid metric value",
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  400,
			},
			request: "/update/gauge/someMetric/abc",
		},
	}

	r := chi.NewRouter()
	gauges := make(map[string]float64)
	counters := make(map[string]int64)
	ms := mockstorage.New(gauges, counters)
	uc := usecase.New(ms)

	r.Post("/update/{metric_type}/{metric_name}/{metric_value}", UpdateMetricsHandler(uc))

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, tc.request, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, request)

			result := w.Result()

			defer result.Body.Close()

			assert.Equal(t, tc.want.statusCode, result.StatusCode)
			assert.Equal(t, tc.want.contentType, result.Header.Get("Content-Type"))
		})
	}
}

func TestGetMetricHandler(t *testing.T) {
	type want struct {
		contentType string
		statusCode  int
	}
	testCases := []struct {
		name    string
		request string
		want    want
	}{
		{
			name: "Valid metric",
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  404,
			},
			request: "/value/gauge/myMetric",
		},
		{
			name: "Invalid metric type",
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  404,
			},
			request: "/value/metric/myMetric",
		},
		{
			name: "Not exist metric",
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  404,
			},
			request: "/value/gauge/someMetric",
		},
	}

	r := chi.NewRouter()
	gauges := make(map[string]float64)
	counters := make(map[string]int64)
	ms := mockstorage.New(gauges, counters)
	uc := usecase.New(ms)

	r.Get("/value/{metric_type}/{metric_name}/", GetMetricHandler(uc))

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, tc.request, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, request)

			result := w.Result()

			defer result.Body.Close()

			assert.Equal(t, tc.want.statusCode, result.StatusCode)
			assert.Equal(t, tc.want.contentType, result.Header.Get("Content-Type"))
		})
	}
}

func TestUpdateMetricsJSONHandler(t *testing.T) {
	type want struct {
		contentType string
		statusCode  int
	}
	testCases := []struct {
		name        string
		request     string
		metricID    string
		metricType  string
		metricValue float64
		want        want
	}{
		{
			name:        "Valid metric",
			metricID:    "TotalAlloc",
			metricType:  "gauge",
			metricValue: 199840.0,
			want: want{
				contentType: "application/json",
				statusCode:  200,
			},
		},
		{
			name:        "Invalid metric type",
			metricID:    "TotalAlloc",
			metricType:  "invalid",
			metricValue: 199840.0,
			want: want{
				contentType: "application/json",
				statusCode:  400,
			},
		},
		{
			name:        "Empty metric name",
			metricID:    "",
			metricType:  "gauge",
			metricValue: 199840.0,
			want: want{
				contentType: "application/json",
				statusCode:  404,
			},
		},
	}

	r := chi.NewRouter()
	gauges := make(map[string]float64)
	counters := make(map[string]int64)
	ms := mockstorage.New(gauges, counters)
	uc := usecase.New(ms)

	r.Post("/update", UpdateMetricsJSONHandler(uc))

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m := metric.Metric{
				ID:    tc.metricID,
				MType: tc.metricType,
				Value: &tc.metricValue,
			}
			b, _ := json.Marshal(m)

			request := httptest.NewRequest(http.MethodPost, "/update", bytes.NewReader(b))
			w := httptest.NewRecorder()

			r.ServeHTTP(w, request)

			result := w.Result()

			defer result.Body.Close()

			assert.Equal(t, tc.want.statusCode, result.StatusCode)
			assert.Equal(t, tc.want.contentType, result.Header.Get("Content-Type"))
		})
	}
}

func TestGetMetricJSONHandler(t *testing.T) {
	type want struct {
		contentType string
		statusCode  int
	}
	testCases := []struct {
		name       string
		metricID   string
		metricType string
		want       want
	}{
		{
			name:       "Valid metric",
			metricID:   "TotalAlloc",
			metricType: "gauge",
			want: want{
				contentType: "application/json",
				statusCode:  200,
			},
		},
		{
			name:       "Invalid metric type",
			metricID:   "TotalAlloc",
			metricType: "invalid",
			want: want{
				contentType: "application/json",
				statusCode:  404,
			},
		},
		{
			name:       "Not exist metric",
			metricID:   "NotExist",
			metricType: "gauge",
			want: want{
				contentType: "application/json",
				statusCode:  404,
			},
		},
	}

	r := chi.NewRouter()
	gauges := map[string]float64{"TotalAlloc": 199840}
	counters := map[string]int64{"PollCount": 10}
	ms := mockstorage.New(gauges, counters)
	uc := usecase.New(ms)

	r.Post("/value", GetMetricsJSONHandler(uc))

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m := metric.Metric{
				ID:    tc.metricID,
				MType: tc.metricType,
			}
			b, _ := json.Marshal(m)

			request := httptest.NewRequest(http.MethodPost, "/value", bytes.NewReader(b))
			w := httptest.NewRecorder()

			r.ServeHTTP(w, request)

			result := w.Result()

			defer result.Body.Close()

			assert.Equal(t, tc.want.statusCode, result.StatusCode)
			assert.Equal(t, tc.want.contentType, result.Header.Get("Content-Type"))
		})
	}
}
