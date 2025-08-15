package redis

import (
	"authz/internal/cfg"
	"authz/internal/pkg"
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

func New(lc *pkg.LifecycleParallel) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Cfg.Redis.Host + ":" + cfg.Cfg.Redis.Port,
		Password: cfg.Cfg.Redis.Password,
		DB:       0,
	})

	lc.Add(rdb, rdb.Close)
	return rdb
}

// TODO: use distributed lock to do update graph check point
type DistributedLock struct {
	client *redis.Client
	key    string
	value  string
	ttl    time.Duration
}

func NewDistributedLock(client *redis.Client, key string, ttl time.Duration) *DistributedLock {
	return &DistributedLock{
		client: client,
		key:    key,
		ttl:    ttl,
	}
}

// TryLock attempts to acquire the lock. Returns true if successful.
func (l *DistributedLock) TryLock(ctx context.Context) (bool, error) {
	l.value = uuid.NewString() // unique per holder
	ok, err := l.client.SetNX(ctx, l.key, l.value, l.ttl).Result()
	return ok, err
}

// Unlock releases the lock only if the value matches (safe unlock)
func (l *DistributedLock) Unlock(ctx context.Context) (bool, error) {
	// Lua script ensures atomic check-and-delete
	script := redis.NewScript(`
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		else
			return 0
		end
	`)
	res, err := script.Run(ctx, l.client, []string{l.key}, l.value).Int()
	return res == 1, err
}
