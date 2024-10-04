package storage

import (
	"bufio"
	"encoding/json"
	"github.com/pkg/errors"
	"os"
)

type MemStorage struct {
	gauges   map[string]float64
	counters map[string]int64
}

type Backup struct {
	Gauges   map[string]float64 `json:"gauges"`
	Counters map[string]int64   `json:"counters"`
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
	}
}

func (ms *MemStorage) UpdateGauge(id string, value float64) {
	ms.gauges[id] = value
}

func (ms *MemStorage) UpdateCounter(id string, value int64) {
	ms.counters[id] += value
}

func (ms *MemStorage) ReceiveGauge(id string) (float64, bool) {
	val, ok := ms.gauges[id]
	return val, ok
}

func (ms *MemStorage) ReceiveCounter(id string) (int64, bool) {
	val, ok := ms.counters[id]
	return val, ok
}

func (ms *MemStorage) ReceiveAllGauges() map[string]float64 {
	return ms.gauges
}

func (ms *MemStorage) ReceiveAllCounters() map[string]int64 {
	return ms.counters
}

func (ms *MemStorage) ReceiveAllMetrics() map[string]interface{} {
	return map[string]interface{}{
		"gauges":   ms.gauges,
		"counters": ms.counters,
	}
}

func (ms *MemStorage) RestoreMetrics(file *os.File) error {
	s := bufio.NewScanner(file)
	var endVal string

	for s.Scan() {
		l := s.Text()
		if l != "" {
			endVal = l
		}
	}

	if err := s.Err(); err != nil {
		return errors.Wrap(err, "read backup metrics file")
	}

	if endVal == "" {
		return nil
	}

	var bf Backup

	err := json.Unmarshal([]byte(endVal), &bf)
	if err != nil {
		return errors.Wrap(err, "unmarshal metrics")
	}

	for id, val := range bf.Gauges {
		ms.UpdateGauge(id, val)
	}

	for id, delta := range bf.Counters {
		ms.UpdateCounter(id, delta)
	}

	return nil
}

func (ms *MemStorage) BackupMetrics(f *os.File) error {
	m := ms.ReceiveAllMetrics()

	d, err := json.Marshal(m)
	if err != nil {
		return errors.Wrap(err, "marshal metrics")
	}

	_, err = f.Write(d)
	if err != nil {
		return errors.Wrap(err, "write metrics")
	}

	_, err = f.WriteString("\n")
	if err != nil {
		return errors.Wrap(err, "write metrics")
	}

	return nil
}
