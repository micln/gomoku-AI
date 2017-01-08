package gomoku_AI

import (
	"testing"

	"docker/pkg/testutil/assert"
)

func TestPoint_String(t *testing.T) {
	p := NewPoint(4, 5)
	assert.Equal(t, p.String(), `(4,5)`)
}
