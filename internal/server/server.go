package server

import (
	"net/http"

	"cosmic-go/internal/application"
	"cosmic-go/internal/domain"
	"cosmic-go/internal/handler"

	eventpublisher "cosmic-go/internal/infra/event_publisher"
	emailservice "cosmic-go/internal/infra/external/email_service"
	unitofwork "cosmic-go/internal/uow"

	"gorm.io/gorm"
)

func InitializeServer(db *gorm.DB) *http.ServeMux {
	ep := eventpublisher.NewEventPublisher()
	uow := unitofwork.NewAllocationUnitOfWork(db, ep)
	svc := application.NewAllocationService(uow)
	handler := handler.NewAllocationHandler(ep)

	em := emailservice.EmailService{}
	ep.RegisterHandler(&domain.OutOfStock{}, em.SendEmail)
	ep.RegisterHandler(&domain.CreateProduct{}, svc.AddProduct)
	ep.RegisterHandler(&domain.AllocationRequired{}, svc.Allocate)
	ep.RegisterHandler(&domain.DeallocationRequired{}, svc.Deallocate)

	router := InitializeRouter(handler)
	return router
}
