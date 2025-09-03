package tuplememory

import (
	"testing"
)

func TestCalculateMemoryUsage(t *testing.T) {
	// Call calculateMemoryUsage with smaller parameters to avoid out-of-memory errors.
	calculateMemoryUsage(100, 5)
}
