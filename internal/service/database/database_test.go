package database_test

import (
	"context"
	"testing"

	"authz/internal/config"
	"authz/internal/service/database"
	"authz/internal/util"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	// Mock the config to use sqlite
	config.Conf = config.Config{
		Db: struct {
			Driver   string `env:"DB_DRIVER"`
			User     string `env:"DB_USER"`
			Password string `env:"DB_PASSWORD"`
			Host     string `env:"DB_HOST"`
			Port     int    `env:"DB_PORT"`
			Db       string `env:"DB_DB"`
		}{
			Driver: "sqlite",
			Host:   "file::memory:",
		},
	}

	lc := util.NewLifecycleParallel()
	db, err := database.New(lc)
	assert.NoError(t, err)
	assert.NotNil(t, db)

	// Close the db
	err = lc.Shutdown(context.Background())
	assert.NoError(t, err)
}

func TestNew_UnsupportedDriver(t *testing.T) {
	// Mock the config to use an unsupported driver
	config.Conf = config.Config{
		Db: struct {
			Driver   string `env:"DB_DRIVER"`
			User     string `env:"DB_USER"`
			Password string `env:"DB_PASSWORD"`
			Host     string `env:"DB_HOST"`
			Port     int    `env:"DB_PORT"`
			Db       string `env:"DB_DB"`
		}{
			Driver: "unsupported",
		},
	}

	lc := util.NewLifecycleParallel()
	_, err := database.New(lc)
	assert.Error(t, err)
}

func TestNew_Postgres(t *testing.T) {
	// Mock the config to use postgres
	config.Conf = config.Config{
		Db: struct {
			Driver   string `env:"DB_DRIVER"`
			User     string `env:"DB_USER"`
			Password string `env:"DB_PASSWORD"`
			Host     string `env:"DB_HOST"`
			Port     int    `env:"DB_PORT"`
			Db       string `env:"DB_DB"`
		}{
			Driver:   "postgres",
			Host:     "localhost",
			Port:     5432,
			User:     "user",
			Password: "password",
			Db:       "testdb",
		},
	}

	// We can't easily test the gorm.Open with a mock, as it involves a real connection attempt.
	// So we can only test that the function returns an error when it can't connect.
	// To truly test the connection string, we would need a more sophisticated test setup.
	lc := util.NewLifecycleParallel()
	_, err := database.New(lc)
	assert.Error(t, err)
}
