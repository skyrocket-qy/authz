package util

import (
	"bytes"
	"errors"
	"testing"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
)

func TestLogE(t *testing.T) {
	// --- Setup ---
	// Redirect zerolog output to a buffer for inspection
	var logBuf bytes.Buffer

	originalLogger := log.Logger
	log.Logger = zerolog.New(&logBuf)

	defer func() {
		// Restore original logger
		log.Logger = originalLogger
	}()
	// --- End Setup ---

	t.Run("should do nothing for standard error", func(t *testing.T) {
		logBuf.Reset() // Clear buffer before test

		// Arrange
		err := errors.New("a plain error")

		// Act
		LogE(err)

		// Assert
		assert.Empty(t, logBuf.String())
	})

	t.Run("should do nothing for a nil error", func(t *testing.T) {
		logBuf.Reset()

		// Act
		LogE(nil)

		// Assert
		assert.Empty(t, logBuf.String())
	})

	// NOTE: Testing the erx.CtxErr path of LogE is not currently possible
	// in this package's unit tests for the same reasons outlined in
	// connect_test.go: it would require creating a circular dependency
	// to import valid errors, and the `erx` library does not seem to
	// provide a stable public API for creating test-only errors.
}
