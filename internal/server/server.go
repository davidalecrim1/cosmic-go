package server

import (
	"net/http"

	"cosmic-go/internal/handler"
	"cosmic-go/internal/infra/repository"
	"cosmic-go/internal/service"

	"github.com/jackc/pgx/v5/pgxpool"
)

func InitializeServer(db *pgxpool.Pool) *http.ServeMux {
	repo := repository.NewPostgresRepository(db)
	svc := service.NewService(repo)
	handler := handler.NewHandler(svc)

	router := InitializeRouter(handler)
	return router
}
