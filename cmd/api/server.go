package main

import (
	"cosmic-go/internal/bootstrap"
	"cosmic-go/internal/handler"
	"cosmic-go/internal/infra/repository"
	"cosmic-go/internal/service"
	"log"
	"net/http"
)

func main() {
	db := bootstrap.InitializeDatabase()
	defer db.Close()

	repo := repository.NewPostgresRepository(db)
	svc := service.NewService(repo)
	handler := handler.NewHandler(svc)

	router := bootstrap.InitializeRouter(handler)

	err := http.ListenAndServe(":8080", router)
	if err != nil {
		log.Fatalln("Server failed to start:", err)
	}
}
