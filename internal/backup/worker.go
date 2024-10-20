package backup

import (
	"github.com/sshirox/isaac/internal/storage"
	"log/slog"
	"os"
	"time"
)

func RunWorker(
	ms *storage.MemStorage,
	interval int64,
	f *os.File,
	sc chan struct{},
) {
	t := time.NewTicker(time.Duration(interval) * time.Second)
	for {
		select {
		case <-t.C:
			if err := SaveBackup(ms, f); err != nil {
				slog.Error("backup metrics", "err", err)
			}
		case <-sc:
			slog.Info("stop backup worker")
			return
		}
	}
}
