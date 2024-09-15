package mockstorage

import (
	"errors"
	"slices"
)

const (
	GaugeMetricType   = "gauge"
	CounterMetricType = "counter"
)

var (
	metricTypes = []string{GaugeMetricType, CounterMetricType}
)

type MockStorage struct{}

func New() *MockStorage {
	return &MockStorage{}
}

func (m *MockStorage) Upsert(_, _, _ string) error {
	return nil
}

func (m *MockStorage) Get(metricType, name string) (string, error) {
	metrics := []string{"myMetric"}

	if !slices.Contains(metricTypes, metricType) {
		return "", errors.New("invalid metric type")
	}

	if slices.Contains(metrics, name) {
		return "val", nil
	} else {
		return "", errors.New("not found metric")
	}
}

func (m *MockStorage) GetAllGauges() map[string]string {
	return map[string]string{}
}
