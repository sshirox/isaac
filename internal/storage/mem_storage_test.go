package storage

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMemStorage_ReceiveAllCounters(t *testing.T) {
	ms := NewMemStorage()
	expected := map[string]int64{"PollCount": 5}

	ms.UpdateCounter("PollCount", 5)

	t.Run("Receive all counter metrics", func(t *testing.T) {
		assert.Equal(t, expected, ms.ReceiveAllCounters())
	})
}

func TestMemStorage_ReceiveAllGauges(t *testing.T) {
	ms := NewMemStorage()
	expected := map[string]float64{"Alloc": 9765.77, "TotalAlloc": 199879.0}

	ms.UpdateGauge("Alloc", 9765.77)
	ms.UpdateGauge("TotalAlloc", 199879.0)

	t.Run("Receive all gauge metrics", func(t *testing.T) {
		assert.Equal(t, expected, ms.ReceiveAllGauges())
	})
}

func TestMemStorage_ReceiveAllMetrics(t *testing.T) {
	ms := NewMemStorage()
	expected := map[string]interface{}{
		"gauges":   map[string]float64{"Alloc": 9765.77, "TotalAlloc": 199879.0},
		"counters": map[string]int64{"PollCount": 5},
	}

	ms.UpdateCounter("PollCount", 5)
	ms.UpdateGauge("Alloc", 9765.77)
	ms.UpdateGauge("TotalAlloc", 199879.0)

	t.Run("Receive all metrics", func(t *testing.T) {
		assert.Equal(t, expected, ms.ReceiveAllMetrics())
	})
}

func TestMemStorage_ReceiveCounter(t *testing.T) {
	ms := NewMemStorage()
	ms.UpdateCounter("PollCount", 5)

	expected := int64(5)
	got, _ := ms.ReceiveCounter("PollCount")

	t.Run("Receive counter value", func(t *testing.T) {
		assert.Equal(t, expected, got)
	})
}

func TestMemStorage_ReceiveGauge(t *testing.T) {
	ms := NewMemStorage()
	ms.UpdateGauge("Alloc", 9765.77)

	expected := 9765.77
	got, _ := ms.ReceiveGauge("Alloc")

	t.Run("Receive gauge value", func(t *testing.T) {
		assert.Equal(t, expected, got)
	})
}

func TestNewMemStorage(t *testing.T) {
	ms := NewMemStorage()

	assert.Equal(t, "*storage.MemStorage", fmt.Sprintf("%T", ms))
}
