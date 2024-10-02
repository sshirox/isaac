package storage

type MemStorage struct {
	gauges   map[string]float64
	counters map[string]int64
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
	}
}

func (m *MemStorage) UpdateGauge(id string, value float64) {
	m.gauges[id] = value
}

func (m *MemStorage) UpdateCounter(id string, value int64) {
	m.counters[id] += value
}

func (m *MemStorage) ReceiveGauge(id string) (float64, bool) {
	val, ok := m.gauges[id]
	return val, ok
}

func (m *MemStorage) ReceiveCounter(id string) (int64, bool) {
	val, ok := m.counters[id]
	return val, ok
}

func (m *MemStorage) ReceiveAllGauges() map[string]float64 {
	return m.gauges
}

func (m *MemStorage) ReceiveAllCounters() map[string]int64 {
	return m.counters
}
