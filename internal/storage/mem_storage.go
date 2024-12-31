package storage

type MemStorage struct {
	gauges   map[string]float64
	counters map[string]int64
}

// NewMemStorage creates new instance of metrics storage
func NewMemStorage() *MemStorage {
	return &MemStorage{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
	}
}

// UpdateGauge updates metric by value
func (ms *MemStorage) UpdateGauge(id string, value float64) {
	ms.gauges[id] = value
}

// UpdateCounter updates metric by value
func (ms *MemStorage) UpdateCounter(id string, value int64) {
	ms.counters[id] += value
}

// ReceiveGauge get metric by id
func (ms *MemStorage) ReceiveGauge(id string) (float64, bool) {
	val, ok := ms.gauges[id]
	return val, ok
}

// ReceiveCounter get metric by id
func (ms *MemStorage) ReceiveCounter(id string) (int64, bool) {
	val, ok := ms.counters[id]
	return val, ok
}

// ReceiveAllGauges get all gauge metrics
func (ms *MemStorage) ReceiveAllGauges() map[string]float64 {
	return ms.gauges
}

// ReceiveAllCounters get all counter metrics
func (ms *MemStorage) ReceiveAllCounters() map[string]int64 {
	return ms.counters
}

// ReceiveAllMetrics get all gauge and counter metrics
func (ms *MemStorage) ReceiveAllMetrics() map[string]interface{} {
	return map[string]interface{}{
		"gauges":   ms.gauges,
		"counters": ms.counters,
	}
}
