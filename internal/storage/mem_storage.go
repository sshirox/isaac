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

type Repository interface {
	Upsert(string, string, string) error
	Get(string, string) (string, error)
	GetAllGauges() map[string]string
}

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

func (m *MemStorage) Upsert(kind, name, value string) error {
	switch kind {
	case GaugeMetricType:
		m.Store[kind][name] = value
	case CounterMetricType:
		m.upsertCounterMetric(kind, name, value)
	}

	return nil
}

func (m *MemStorage) upsertCounterMetric(kind, name, value string) error {
	if v, ok := m.Store[kind][name]; ok {
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
		m.Store[kind][name] = res
	} else {
		m.Store[kind][name] = value
	}

	return nil
}

func (m *MemStorage) Get(kind, name string) (string, error) {
	if !slices.Contains(metricTypes, kind) {
		return "", errors.New("invalid metric type")
	}

	if val, found := m.Store[kind][name]; found {
		return val, nil
	} else {
		return "", errors.New("not found metric")
	}
}

func (m *MemStorage) GetAllGauges() map[string]string {
	return m.Store["gauge"]
}
