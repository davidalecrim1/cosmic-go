package main

import (
	"cosmic-go/internal/bootstrap"
	"cosmic-go/internal/handler"
	"cosmic-go/internal/infra/repository"
	"cosmic-go/internal/service"
	"net/http"
)

func main() {
	db := bootstrap.InitializeDatabase()

	repo := repository.NewPostgresRepository(db)
	svc := service.NewService(repo)
	handler := handler.NewHandler(svc)

	router := bootstrap.InitializeRouter(handler)
	http.ListenAndServe(":8080", router)
}
