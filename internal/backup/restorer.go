package backup

import (
	"bufio"
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/sshirox/isaac/internal/storage"
	"log/slog"
	"os"
	"path"
)

type backupFile struct {
	Gauges   map[string]float64 `json:"gauges"`
	Counters map[string]int64   `json:"counters"`
}

func RestoreMetrics(
	ms *storage.MemStorage,
	storagePath string,
	enable bool,
) (
	*storage.MemStorage,
	*os.File,
	error,
) {
	var err error
	var f *os.File

	if err = os.MkdirAll(storagePath, 0755); err != nil {
		slog.Error("create backup dir", "error", err)
		return nil, nil, err
	}

	fp := path.Join(storagePath, "metrics.bk")
	f, err = os.OpenFile(fp, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		slog.Error("open file", "error", err)
		return nil, nil, err
	}

	if enable {
		err = restore(ms, f)
		if err != nil {
			slog.Error("restore metrics", "error", err)
			return nil, nil, err
		}
	}
	return ms, f, nil
}

func restore(ms *storage.MemStorage, file *os.File) error {
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

	var bf backupFile

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
