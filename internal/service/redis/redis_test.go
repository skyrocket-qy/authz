package redis_test

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"testing"
	"time"

	"authz/internal/config"
	"authz/internal/service/redis"
	"authz/internal/util"
	"github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	// Mock the config
	config.Conf.Redis.Host = "localhost"
	config.Conf.Redis.Port = "6379"

	lc := util.NewLifecycleParallel()
	rdb := redis.New(lc)
	assert.NotNil(t, rdb)
}

func TestDistributedLock_TryLock(t *testing.T) {
	db, mock := redismock.NewClientMock()
	key := "test-lock"
	ttl := 10 * time.Second
	lock := redis.NewDistributedLock(db, key, ttl)

	// Mock the UUID generation
	originalNewUUID := redis.NewUUID

	defer func() { redis.NewUUID = originalNewUUID }()

	redis.NewUUID = func() string { return "test-uuid" }

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

func TestDistributedLock_Unlock(t *testing.T) {
	db, mock := redismock.NewClientMock()
	key := "test-lock"
	ttl := 10 * time.Second
	lock := redis.NewDistributedLock(db, key, ttl)

	// Mock the UUID generation to have a predictable value
	originalNewUUID := redis.NewUUID
	defer func() { redis.NewUUID = originalNewUUID }()
	redis.NewUUID = func() string { return "test-uuid" }

	// First, lock the lock to set the value
	_, _ = lock.TryLock(context.Background())

	script := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		else
			return 0
		end
	`
	h := sha1.New()
	_, _ = h.Write([]byte(script))
	sha := hex.EncodeToString(h.Sum(nil))

	t.Run("Unlock success", func(t *testing.T) {
		mock.ExpectEvalSha(sha, []string{key}, "test-uuid").SetVal(int64(1))

		unlocked, err := lock.Unlock(context.Background())
		assert.NoError(t, err)
		assert.True(t, unlocked)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Unlock failure", func(t *testing.T) {
		mock.ExpectEvalSha(sha, []string{key}, "test-uuid").SetVal(int64(0))

		unlocked, err := lock.Unlock(context.Background())
		assert.NoError(t, err)
		assert.False(t, unlocked)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
