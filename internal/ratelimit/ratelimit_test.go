package ratelimit

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
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
