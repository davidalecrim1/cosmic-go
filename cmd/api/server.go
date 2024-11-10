package main

import (
	"context"
	"log"
	"net/http"

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

	pubsub := messagepublisher.InitializeRedis()

	router := server.InitializeServer(ctx, db, pubsub)
	err := http.ListenAndServe(":8080", router)
	if err != nil {
		log.Fatalln("Server failed to start:", err)
	}
}
