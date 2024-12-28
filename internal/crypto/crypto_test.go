package crypto

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEncoder_Encode(t *testing.T) {
	enc := NewEncoder("key")

	expected := "97d15beaba060d0738ec759ea31865178ab8bb781b2d2107644ba881f399d8d6"
	got := enc.Encode([]byte("string"))

	assert.Equal(t, expected, got)
}

func TestEncoder_IsEnabled(t *testing.T) {
	enc := NewEncoder("key")

	assert.True(t, enc.IsEnabled())
}
