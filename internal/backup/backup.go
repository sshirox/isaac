package backup

import (
	"encoding/json"
	"os"

	"github.com/pkg/errors"

	"github.com/sshirox/isaac/internal/storage"
)

// SaveBackup save metrics from storage to file
func SaveBackup(ms *storage.MemStorage, f *os.File) error {
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
