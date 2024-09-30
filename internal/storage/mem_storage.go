package storage

import (
	"errors"
	"github.com/sshirox/isaac/internal/model"
	"slices"
	"strconv"
)

const (
	GaugeMetricType   = "gauge"
	CounterMetricType = "counter"
)

var (
	metricTypes = []string{GaugeMetricType, CounterMetricType}
)

type MemStorage struct {
	Gauges   map[string]float64
	Counters map[string]int64
	Store    map[string]map[string]string
}

func NewMemStorage(
	gaugeStore,
	counterStore map[string]string,
	gauges map[string]float64,
	counters map[string]int64,
) (*MemStorage, error) {
	store := map[string]map[string]string{
		GaugeMetricType:   gaugeStore,
		CounterMetricType: counterStore,
	}

	return &MemStorage{
		Store:    store,
		Gauges:   gauges,
		Counters: counters,
	}, nil
}

func (m *MemStorage) Upsert(metricType, name, value string) error {
	switch metricType {
	case GaugeMetricType:
		m.Store[metricType][name] = value
	case CounterMetricType:
		m.upsertCounterMetric(metricType, name, value)
	}

	return nil
}

func (m *MemStorage) upsertCounterMetric(metricType, name, value string) error {
	if v, ok := m.Store[metricType][name]; ok {
		currVal, err := strconv.Atoi(v)
		if err != nil {
			panic(err)
		}
		newVal, err := strconv.Atoi(value)
		if err != nil {
			panic(err)
		}
		updatedVal := currVal + newVal
		res := strconv.Itoa(updatedVal)
		m.Store[metricType][name] = res
	} else {
		m.Store[metricType][name] = value
	}

	return nil
}

func (m *MemStorage) Get(metricType, name string) (string, error) {
	if !slices.Contains(metricTypes, metricType) {
		return "", errors.New("invalid metric type")
	}

	if val, found := m.Store[metricType][name]; found {
		return val, nil
	} else {
		return "", errors.New("not found metric")
	}
}

func (m *MemStorage) Update(MType, id string, value *float64, delta *int64) model.Metric {
	var res model.Metric
	switch MType {
	case GaugeMetricType:
		res = m.updateGaugeMetric(id, value)
	case CounterMetricType:
		res = m.updateCounterMetric(id, delta)
	}

	return res
}

func (m *MemStorage) updateGaugeMetric(id string, value *float64) model.Metric {
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

func (m *MemStorage) updateCounterMetric(id string, delta *int64) model.Metric {
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

func (m *MemStorage) Value(MType, id string) (model.Metric, error) {
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
