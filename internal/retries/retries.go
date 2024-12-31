package retries

import (
	"errors"
	"time"

	errs "github.com/sshirox/isaac/internal/errors"
)

// Retry repeats a function with an error at preset intervals
func Retry(op func() error) error {
	var err error
	for _, interval := range intervals {
		err = op()
		if err == nil {
			return nil
		}
		if errors.Is(err, errs.ErrNonRetry) {
			return err
		}
		time.Sleep(interval)
	}

	return err
}
