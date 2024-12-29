package ratelimit

type Limiter struct {
	ch chan struct{}
}

// NewLimiter creates new instance of Limiter
func NewLimiter(limit int64) *Limiter {
	return &Limiter{
		ch: make(chan struct{}, limit),
	}
}

// Acquire blocks limiter buffer
func (l *Limiter) Acquire() {
	l.ch <- struct{}{}
}

// Release frees limiter buffer
func (l *Limiter) Release() {
	<-l.ch
}
