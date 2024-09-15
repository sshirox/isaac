package storage

import (
	"errors"
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
	Store map[string]map[string]string
}

func NewMemStorage(gaugeStore, counterStore map[string]string) (*MemStorage, error) {
	store := map[string]map[string]string{
		GaugeMetricType:   gaugeStore,
		CounterMetricType: counterStore,
	}

	return &MemStorage{Store: store}, nil
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

func (m *MemStorage) GetAllGauges() map[string]string {
	return m.Store["gauge"]
}
