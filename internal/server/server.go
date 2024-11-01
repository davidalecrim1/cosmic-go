package server

import (
	"net/http"

	"cosmic-go/internal/application"
	"cosmic-go/internal/domain"
	"cosmic-go/internal/handler"

	messagepublisher "cosmic-go/internal/infra/event_publisher"
	emailservice "cosmic-go/internal/infra/external/email_service"
	unitofwork "cosmic-go/internal/uow"

	"gorm.io/gorm"
)

func InitializeServer(db *gorm.DB) *http.ServeMux {
	mp := messagepublisher.NewMessagePublisher()
	uow := unitofwork.NewAllocationUnitOfWork(db, mp)
	svc := application.NewAllocationService(uow)
	handler := handler.NewAllocationHandler(mp)

	em := emailservice.EmailService{}
	mp.RegisterEventHandler(&domain.OutOfStock{}, em.SendEmail)

	mp.RegisterCommandHandler(&domain.CreateProduct{}, svc.AddProduct)
	mp.RegisterCommandHandler(&domain.Allocate{}, svc.Allocate)
	mp.RegisterCommandHandler(&domain.Deallocate{}, svc.Deallocate)

	router := InitializeRouter(handler)
	return router
}
