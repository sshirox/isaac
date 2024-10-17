package retries

import "time"

var (
	intervals = []time.Duration{
		1 * time.Second,
		3 * time.Second,
		5 * time.Second,
	}
)
