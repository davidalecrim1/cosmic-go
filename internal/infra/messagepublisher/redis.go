package messagepublisher

import (
	"cosmic-go/pkg/env"

	"github.com/redis/go-redis/v9"
)

func InitializeRedis() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: env.GetEnvOrSetDefault("REDIS_ENDPOINT", "localhost:6379"),
		DB:   0, // default
	})
}
