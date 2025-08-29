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
