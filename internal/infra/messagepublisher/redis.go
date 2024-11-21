package messagepublisher

import (
	"github.com/redis/go-redis/v9"
)

func InitializeRedis(addr string) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: addr,
		DB:   0, // default
	})
}
