package server

import (
	"net/http"

	"cosmic-go/internal/application"
	"cosmic-go/internal/handler"

	"gorm.io/gorm"
)

func InitializeServer(db *gorm.DB) *http.ServeMux {
	uow := application.NewBatchUnitOfWork(db)
	svc := application.NewService(uow)
	handler := handler.NewHandler(svc)

	router := InitializeRouter(handler)
	return router
}
