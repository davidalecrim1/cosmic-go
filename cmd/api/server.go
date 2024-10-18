package main

import (
	"log"
	"net/http"

	"cosmic-go/internal/infra/database"
	"cosmic-go/internal/server"
)

func main() {
	db := database.InitializeDatabase()
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	router := server.InitializeServer(db)
	err := http.ListenAndServe(":8080", router)
	if err != nil {
		log.Fatalln("Server failed to start:", err)
	}
}
