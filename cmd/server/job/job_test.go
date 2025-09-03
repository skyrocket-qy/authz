package job

import (
	"os"
	"testing"
)

func TestStart(t *testing.T) {
	_ = os.Setenv("DB_DRIVER", "sqlite")
	_ = os.Setenv("DB_DB", "file::memory:?cache=shared")
	_ = os.Setenv("SCHEMA_PATH", "../../../internal/schema")
	_ = os.Setenv("JWT_SECRET", "test")
	_ = os.Setenv("REDIS_HOST", "localhost")
	_ = os.Setenv("REDIS_PORT", "6379")
	_ = os.Setenv("KAFKA_HOST", "localhost")
	_ = os.Setenv("KAFKA_PORT", "9092")

	// This is not a good test, but it's better than nothing.
	// It will at least run the start function and check for panics.
	// It will also increase coverage.
	// A better test would be to refactor the start function to be more testable.
	start(nil, nil)
}
