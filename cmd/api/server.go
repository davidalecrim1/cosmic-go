package main

import (
	"context"

	"cosmic-go/internal/infra/database"
	"cosmic-go/internal/infra/messagepublisher"
	"cosmic-go/internal/server"
	"cosmic-go/pkg/env"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db := database.NewDatabase()
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	redisClient := messagepublisher.InitializeRedis(
		env.GetEnvOrSetDefault("REDIS_ENDPOINT", "localhost:6379"),
	)

	httpServer := server.NewServer()
	httpServer.InitializeDependencies(ctx, db, redisClient)

	go httpServer.Run()
}
