package metric

const (
	GaugeMetricType   = "gauge"
	CounterMetricType = "counter"
)

var (
	ValidMetricTypes = []string{GaugeMetricType, CounterMetricType}
)
