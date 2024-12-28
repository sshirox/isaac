package ratelimit

type Limiter struct {
	ch chan struct{}
}

func NewLimiter(limit int64) *Limiter {
	return &Limiter{
		ch: make(chan struct{}, limit),
	}
}

func (l *Limiter) Acquire() {
	l.ch <- struct{}{}
}

func (l *Limiter) Release() {
	<-l.ch
}
