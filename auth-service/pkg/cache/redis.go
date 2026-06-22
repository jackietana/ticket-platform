package cache

import "github.com/redis/go-redis/v9"

func NewRedisConnection(addr string) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "",
		DB:       0,
		Protocol: 2,
	})
}
