package ratelimit

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLimiter_Release(t *testing.T) {
	lim := NewLimiter(1)
	lim.Acquire()
	lim.Release()
}

func TestNewLimiter(t *testing.T) {
	lim := NewLimiter(10)

	assert.Equal(t, "*ratelimit.Limiter", fmt.Sprintf("%T", lim))
}
