package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/sshirox/isaac/internal/storage"
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
	gaugeStore := make(map[string]string)
	counterStore := make(map[string]string)
	ms, err := storage.NewMemStorage(gaugeStore, counterStore)

	if err != nil {
		panic(err)
	}

	r.Post("/update/{metric_type}/{metric_name}/{metric_value}", UpdateMetricsHandler(ms))

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
	gaugeStore := map[string]string{"myMetrc": "789.00"}
	counterStore := make(map[string]string)
	ms, err := storage.NewMemStorage(gaugeStore, counterStore)

	if err != nil {
		panic(err)
	}

	r.Get("/value/{metric_type}/{metric_name}/", GetMetricHandler(ms))

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
