package cache

import "github.com/redis/go-redis/v9"

func NewRedisConnection(addr, pass string) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: pass,
		DB:       0,
	})
}
