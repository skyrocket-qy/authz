package redis

import (
	"context"
	"testing"
	"time"

	"authz/internal/config"
	"authz/internal/util"
	"github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	// Mock the config
	config.Conf.Redis.Host = "localhost"
	config.Conf.Redis.Port = "6379"

	lc := util.NewLifecycleParallel()
	rdb := New(lc)
	assert.NotNil(t, rdb)
}

func TestDistributedLock_TryLock(t *testing.T) {
	db, mock := redismock.NewClientMock()
	key := "test-lock"
	ttl := 10 * time.Second
	lock := NewDistributedLock(db, key, ttl)

	// Mock the UUID generation
	originalNewUUID := newUUID

	defer func() { newUUID = originalNewUUID }()

	newUUID = func() string { return "test-uuid" }

	t.Run("TryLock success", func(t *testing.T) {
		mock.ExpectSetNX(key, "test-uuid", ttl).SetVal(true)

		locked, err := lock.TryLock(context.Background())
		assert.NoError(t, err)
		assert.True(t, locked)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("TryLock failure", func(t *testing.T) {
		mock.ExpectSetNX(key, "test-uuid", ttl).SetVal(false)

		locked, err := lock.TryLock(context.Background())
		assert.NoError(t, err)
		assert.False(t, locked)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
