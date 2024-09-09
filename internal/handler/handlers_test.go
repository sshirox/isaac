package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetricsHandler(t *testing.T) {
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

	mux := &http.ServeMux{}
	mux.HandleFunc("POST /update/{metric_type}/{metric_name}/{metric_value}", MetricsHandler())

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, tc.request, nil)
			w := httptest.NewRecorder()

			mux.ServeHTTP(w, request)

			result := w.Result()

			assert.Equal(t, tc.want.statusCode, result.StatusCode)
			assert.Equal(t, tc.want.contentType, result.Header.Get("Content-Type"))
		})
	}
}
