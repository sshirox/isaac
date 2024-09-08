package storage

const (
	GaugeMetricType   = "gauge"
	CounterMetricType = "counter"
)

type MetricValue interface {
	~float64 | ~int64
}

type MetricArg struct {
	Type  string
	Name  string
	Value string
}

type Repository interface {
	Upsert(MetricArg) error
}

type MemStorage[T MetricValue] struct {
	Store map[string]map[string]T
}
