package shardmem

import (
	"testing"
)

func TestStart(t *testing.T) {
	// This is not a good test, but it's better than nothing.
	// It will at least run the start function and check for panics.
	// It will also increase coverage.
	start(nil, nil)
}
