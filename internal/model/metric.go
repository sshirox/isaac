package model

type Metric struct {
	ID    string   `json:"id"`              // metric name
	MType string   `json:"type"`            // parameter that takes the value gauge or counter
	Delta *int64   `json:"delta,omitempty"` // metric value in case of counter transfer
	Value *float64 `json:"value,omitempty"` // metric value in case of gauge transfer
}
