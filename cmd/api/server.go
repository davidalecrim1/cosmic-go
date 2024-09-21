package main

import (
	"cosmic-go/internal/bootstrap"
	"cosmic-go/internal/domain"
	"cosmic-go/internal/handler"
	"cosmic-go/internal/infra/repository"
	"net/http"
)

func main() {
	db := bootstrap.InitializeDatabase()

	repo := repository.NewPostgresRepository(db)
	svc := domain.NewService(repo)
	handler := handler.NewHandler(svc)

	router := bootstrap.InitializeRouter(handler)
	http.ListenAndServe(":8080", router)
}
