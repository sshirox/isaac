package storage

import (
	"errors"
	"github.com/sshirox/isaac/internal/model"
	"slices"
)

const (
	GaugeMetricType   = "gauge"
	CounterMetricType = "counter"
)

var metricTypes = []string{GaugeMetricType, CounterMetricType}

type MemStorage struct {
	Gauges   map[string]float64
	Counters map[string]int64
}

func NewMemStorage(gauges map[string]float64, counters map[string]int64) (*MemStorage, error) {
	return &MemStorage{
		Gauges:   gauges,
		Counters: counters,
	}, nil
}

func (m *MemStorage) Upsert(MType, id string, value *float64, delta *int64) model.Metric {
	var res model.Metric
	switch MType {
	case GaugeMetricType:
		res = m.upsertGaugeMetric(id, value)
	case CounterMetricType:
		res = m.upsertCounterMetric(id, delta)
	}

	return res
}

func (m *MemStorage) upsertGaugeMetric(id string, value *float64) model.Metric {
	newVal := *value
	m.Gauges[id] = newVal

	metric := model.Metric{
		ID:    id,
		MType: GaugeMetricType,
		Delta: nil,
		Value: &newVal,
	}

	return metric
}

func (m *MemStorage) upsertCounterMetric(id string, delta *int64) model.Metric {
	newDelta := *delta
	if v, ok := m.Counters[id]; ok {
		newDelta += v
	}

	metric := model.Metric{
		ID:    id,
		MType: CounterMetricType,
		Delta: &newDelta,
		Value: nil,
	}

	return metric
}

func (m *MemStorage) Get(MType, id string) (model.Metric, error) {
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

func (m *MemStorage) GetAllGauges() map[string]float64 {
	return m.Gauges
}
