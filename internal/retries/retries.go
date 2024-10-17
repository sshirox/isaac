package retries

import (
	"errors"
	errs "github.com/sshirox/isaac/internal/errors"
	"time"
)

func Retry(op func() error) error {
	var err error
	for _, interval := range intervals {
		err = op()
		if err == nil {
			return nil
		}
		if errors.Is(err, errs.NonRetryErr) {
			return err
		}
		time.Sleep(interval)
	}

	return err
}
