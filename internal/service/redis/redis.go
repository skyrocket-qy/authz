package redis

import (
	"authz/internal/cfg"
	"authz/internal/pkg"

	"github.com/redis/go-redis/v9"
)

func New(lc pkg.Lifecycle) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Cfg.Redis.Host + ":" + cfg.Cfg.Redis.Port,
		Password: cfg.Cfg.Redis.Password,
		DB:       0,
	})

	lc.Add(rdb.Close)
	return rdb
}
