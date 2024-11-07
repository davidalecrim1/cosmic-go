package server

import (
	"context"
	"net/http"

	"cosmic-go/internal/application"
	"cosmic-go/internal/domain"
	"cosmic-go/internal/handler"

	messagepublisher "cosmic-go/internal/infra/messagepublisher"
	unitofwork "cosmic-go/internal/uow"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func InitializeServer(ctx context.Context, db *gorm.DB) *http.ServeMux {
	internalMp := messagepublisher.NewMessagePublisher()

	uow := unitofwork.NewAllocationUnitOfWork(db, internalMp)
	svc := application.NewAllocationService(uow)
	hnr := handler.NewAllocationHandler(internalMp)

	redis := messagepublisher.InitializeRedis()
	externalMp := messagepublisher.NewExternalMessagePublisher(redis)

	internalMp.RegisterCommandHandler(&domain.CreateProduct{}, svc.AddProduct)
	internalMp.RegisterCommandHandler(&domain.Allocate{}, svc.Allocate)
	internalMp.RegisterCommandHandler(&domain.Deallocate{}, svc.Deallocate)
	internalMp.RegisterCommandHandler(&domain.ChangeBatchQuantity{}, svc.ChangeBatchQuantity)
	internalMp.RegisterEventHandler(&domain.Allocated{}, externalMp.PublishEvent)

	InitializeExternalMessageConsumer(ctx, redis)

	router := InitializeRouter(hnr)
	return router
}

func InitializeExternalMessageConsumer(ctx context.Context, redis *redis.Client) {
	externalMc := handler.NewExternalMessageConsumer(redis)
	go externalMc.ConsumeChangeBatchQuantityCommand(ctx)
}
