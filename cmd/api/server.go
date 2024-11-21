package main

import (
	"context"

	"cosmic-go/internal/infra/database"
	"cosmic-go/internal/infra/messagepublisher"
	"cosmic-go/internal/server"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db := database.InitializeDatabase()
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	redisClient := messagepublisher.InitializeRedis()

	httpServer := server.NewServer()
	httpServer.InitializeDependencies(ctx, db, redisClient)

	go httpServer.Run()
}
