package util

import (
	"authz/internal/config"
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

// setupAndTeardown resets the global config variables for a clean test run
// and returns a teardown function to restore the original state.
func setupAndTeardown(t *testing.T) func() {
	t.Helper()

	// Store original global state
	originalConf := config.Conf
	originalEnv := config.Env

	// Reset global state for the test
	config.Conf = config.Config{}
	config.Env = ""

	// Return a teardown function to restore original state
	return func() {
		config.Conf = originalConf
		config.Env = originalEnv
	}
}

// Use a mutex to serialize tests that modify global environment state,
// preventing race conditions since subtests can run in parallel.
var configTestMutex sync.Mutex

func TestNewConfig(t *testing.T) {
	t.Run("should load config from .env file in local environment", func(t *testing.T) {
		configTestMutex.Lock()
		defer configTestMutex.Unlock()

		teardown := setupAndTeardown(t)
		defer teardown()

		// Arrange
		config.Env = config.EnvLocal
		content := []byte("MAX_CHECK_NODES=100\nPORT=8080\nDB_USER=testuser")

		// Defer unsetting of environment variables that godotenv will set.
		// This is critical to prevent polluting the environment for other tests.
		defer os.Unsetenv("MAX_CHECK_NODES")
		defer os.Unsetenv("PORT")
		defer os.Unsetenv("DB_USER")

		err := os.WriteFile(".env", content, 0644)
		assert.NoError(t, err)
		defer os.Remove(".env")

		// Act
		err = NewConfig()

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, 100, config.Conf.MaxCheckNodes)
		assert.Equal(t, 8080, config.Conf.Port)
		assert.Equal(t, "testuser", config.Conf.Db.User)
	})

	t.Run("should not return error if .env file is missing in local environment", func(t *testing.T) {
		configTestMutex.Lock()
		defer configTestMutex.Unlock()

		teardown := setupAndTeardown(t)
		defer teardown()

		// Arrange
		config.Env = config.EnvLocal
		os.Remove(".env") // Make sure it's gone

		// Act
		err := NewConfig()

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, config.Config{}, config.Conf)
	})

	t.Run("should load config from environment variables in non-local environment", func(t *testing.T) {
		configTestMutex.Lock()
		defer configTestMutex.Unlock()

		teardown := setupAndTeardown(t)
		defer teardown()

		// Arrange
		config.Env = config.EnvDev
		t.Setenv("MAX_CHECK_NODES", "200")
		t.Setenv("PORT", "9090")
		t.Setenv("JWT_SECRET", "supersecret")

		// Act
		err := NewConfig()

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, 200, config.Conf.MaxCheckNodes)
		assert.Equal(t, 9090, config.Conf.Port)
		assert.Equal(t, "supersecret", config.Conf.Jwt.Secret)
	})

	t.Run("should return error on parse error", func(t *testing.T) {
		configTestMutex.Lock()
		defer configTestMutex.Unlock()

		teardown := setupAndTeardown(t)
		defer teardown()

		// Arrange
		config.Env = config.EnvDev
		t.Setenv("PORT", "not-an-int") // Invalid value

		// Act
		err := NewConfig()

		// Assert
		assert.Error(t, err)
	})
}
