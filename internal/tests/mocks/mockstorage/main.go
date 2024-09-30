package mockstorage

import (
	"errors"
	"github.com/sshirox/isaac/internal/model"
	"log/slog"
	"slices"
)

const (
	GaugeMetricType   = "gauge"
	CounterMetricType = "counter"
)

var (
	metricTypes = []string{GaugeMetricType, CounterMetricType}
)

type MockStorage struct {
	Gauges   map[string]float64
	Counters map[string]int64
}

func New(gauges map[string]float64, counters map[string]int64) *MockStorage {
    return &MockStorage{Gauges: gauges, Counters: counters}
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

func (m *MockStorage) Update(MType, id string, value *float64, delta *int64) model.Metric {
	slog.Info("upsert metric", "MType", MType, "ID", id, "value", value, "delta", delta)

	return model.Metric{}
}

func (m *MockStorage) Value(MType, id string) (model.Metric, error) {
	if !slices.Contains(metricTypes, MType) {
		return model.Metric{}, errors.New("invalid metric type")
	}

	if value, found := m.Gauges[id]; found {
		metric := model.Metric{
			ID:    id,
			MType: GaugeMetricType,
			Delta: nil,
			Value: &value,
		}
		return metric, nil
	} else if delta, found := m.Counters[id]; found {
		metric := model.Metric{
			ID:    id,
			MType: CounterMetricType,
			Delta: &delta,
			Value: nil,
		}
		return metric, nil
	} else {
		return model.Metric{}, errors.New("not found metric")
	}
}

func (m *MockStorage) GetAllGauges() map[string]float64 {
	return map[string]float64{}
}
