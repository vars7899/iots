package redis

import (
	"github.com/redis/go-redis/v9"
	"github.com/vars7899/iots/config"
)

func NewRedisClient(cfg *config.RedisConfig) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
}
