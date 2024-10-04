package handler

import (
	"bytes"
	"encoding/json"
	"github.com/sshirox/isaac/internal/metric"
	"github.com/sshirox/isaac/internal/storage"
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
		body        string
	}
	testCases := []struct {
		name    string
		request string
		want    want
	}{
		{
			name: "Valid gauge metric",
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  200,
				body:        "gauge successfully updated",
			},
			request: "/update/gauge/Alloc/78910987.77",
		},
		{
			name: "Valid counter metric",
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  200,
				body:        "counter successfully updated",
			},
			request: "/update/counter/PollCount/10",
		},
		{
			name: "Empty metric name",
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  400,
				body:        "empty metric name",
			},
			request: "/update/gauge//78910987.77",
		},
		{
			name: "Empty metric value",
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  404,
				body:        "404 page not found\n",
			},
			request: "/update/gauge/Alloc/",
		},
		{
			name: "Invalid metric value",
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  400,
				body:        "metric value is not a float",
			},
			request: "/update/gauge/Alloc/invalid",
		},
		{
			name: "Invalid metric type",
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  400,
				body:        "metric value is not a float",
			},
			request: "/update/gauge/Alloc/invalid",
		},
	}

	r := chi.NewRouter()
	s := storage.NewMemStorage()

	r.Post("/update/{type}/{name}/{value}", UpdateMetricsHandler(s))

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, tc.request, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, request)

			result := w.Result()

			defer result.Body.Close()

			assert.Equal(t, tc.want.statusCode, result.StatusCode)
			assert.Equal(t, tc.want.contentType, result.Header.Get("Content-Type"))
			assert.Equal(t, tc.want.body, w.Body.String())
		})
	}
}

func TestValueMetricHandler(t *testing.T) {
	type want struct {
		contentType string
		statusCode  int
		body        string
	}
	testCases := []struct {
		name    string
		request string
		want    want
	}{
		{
			name: "Success gauge metric",
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  200,
				body:        "789765.77",
			},
			request: "/value/gauge/Alloc/",
		},
		{
			name: "Success counter metric",
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  200,
				body:        "10",
			},
			request: "/value/counter/PollCount/",
		},
		{
			name: "Not found gauge metric",
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  404,
				body:        "metric not found",
			},
			request: "/value/gauge/Alloc1/",
		},
		{
			name: "Not found counter metric",
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  404,
				body:        "metric not found",
			},
			request: "/value/counter/PollCount1/",
		},
		{
			name: "Invalid metric type",
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  400,
				body:        "invalid metric type",
			},
			request: "/value/invalid/PollCount/",
		},
	}

	r := chi.NewRouter()
	s := storage.NewMemStorage()
	s.UpdateGauge("Alloc", 789765.77)
	s.UpdateCounter("PollCount", 10)

	r.Get("/value/{type}/{name}/", ValueMetricHandler(s))

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, tc.request, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, request)

			result := w.Result()

			defer result.Body.Close()

			assert.Equal(t, tc.want.statusCode, result.StatusCode)
			assert.Equal(t, tc.want.contentType, result.Header.Get("Content-Type"))
			assert.Equal(t, tc.want.body, w.Body.String())
		})
	}
}

func TestUpdateByContentTypeHandler(t *testing.T) {
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
				contentType: "application/json; charset=utf-8",
				statusCode:  200,
			},
		},
		{
			name:        "Invalid metric type",
			metricID:    "TotalAlloc",
			metricType:  "TotalAlloc1",
			metricValue: 199840.0,
			want: want{
				contentType: "application/json; charset=utf-8",
				statusCode:  400,
			},
		},
	}

	r := chi.NewRouter()
	s := storage.NewMemStorage()

	r.Post("/update", UpdateByContentTypeHandler(s))

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m := metric.Metrics{
				ID:    tc.metricID,
				MType: tc.metricType,
				Value: &tc.metricValue,
			}
			b, _ := json.Marshal(m)

			request := httptest.NewRequest(http.MethodPost, "/update", bytes.NewReader(b))
			request.Header.Set("Content-Type", "application/json")
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
			metricID:   "Alloc",
			metricType: "gauge",
			want: want{
				contentType: "application/json; charset=utf-8",
				statusCode:  200,
			},
		},
		{
			name:       "Invalid metric type",
			metricID:   "Alloc",
			metricType: "invalid",
			want: want{
				contentType: "application/json; charset=utf-8",
				statusCode:  400,
			},
		},
		{
			name:       "Not exist metric",
			metricID:   "Alloc1",
			metricType: "gauge",
			want: want{
				contentType: "application/json; charset=utf-8",
				statusCode:  404,
			},
		},
	}

	r := chi.NewRouter()
	s := storage.NewMemStorage()
	s.UpdateGauge("Alloc", 789765.77)

	r.Post("/value", ValueByContentTypeHandler(s))

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m := metric.Metrics{
				ID:    tc.metricID,
				MType: tc.metricType,
			}
			b, _ := json.Marshal(m)

			request := httptest.NewRequest(http.MethodPost, "/value", bytes.NewReader(b))
			request.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			r.ServeHTTP(w, request)

			result := w.Result()

			defer result.Body.Close()

			assert.Equal(t, tc.want.statusCode, result.StatusCode)
			assert.Equal(t, tc.want.contentType, result.Header.Get("Content-Type"))
		})
	}
}
