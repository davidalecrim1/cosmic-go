package server

import (
	"net/http"

	"cosmic-go/internal/application"
	"cosmic-go/internal/domain"
	"cosmic-go/internal/handler"
	"cosmic-go/internal/infra/events/publisher"
	unitofwork "cosmic-go/internal/uow"

	"cosmic-go/internal/infra/events/handlers/email"

	"gorm.io/gorm"
)

func InitializeServer(db *gorm.DB) *http.ServeMux {
	ep := initializeEventPublisher()
	uow := unitofwork.NewAllocationUnitOfWork(db, ep)
	svc := application.NewService(uow)
	handler := handler.NewHandler(svc)

	router := InitializeRouter(handler)
	return router
}

func initializeEventPublisher() *publisher.EventPublisher {
	ep := publisher.NewEventPublisher()
	ep.RegisterHandler(&domain.OutOfStockEvent{}, &email.EmailService{})
	return ep
}
