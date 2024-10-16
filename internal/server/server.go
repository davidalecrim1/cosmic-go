package server

import (
	"net/http"

	"cosmic-go/internal/application"
	"cosmic-go/internal/handler"

	"github.com/jackc/pgx/v5/pgxpool"
)

func InitializeServer(db *pgxpool.Pool) *http.ServeMux {
	uow := application.NewBatchUnitOfWork(db)
	svc := application.NewService(uow)
	handler := handler.NewHandler(svc)

	router := InitializeRouter(handler)
	return router
}
