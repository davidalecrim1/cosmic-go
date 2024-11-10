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

func InitializeServer(ctx context.Context, db *gorm.DB, pubsub *redis.Client) *http.ServeMux {
	internalMp := messagepublisher.NewMessagePublisher()

	uow := unitofwork.NewAllocationUnitOfWork(db, internalMp)
	svc := application.NewAllocationService(uow)
	hnr := handler.NewAllocationHandler(internalMp)

	externalMp := messagepublisher.NewExternalMessagePublisher(pubsub)

	// add here internal message publisher handlers
	internalMp.RegisterCommandHandler(&domain.CreateProduct{}, svc.AddProduct)
	internalMp.RegisterCommandHandler(&domain.Allocate{}, svc.Allocate)
	internalMp.RegisterCommandHandler(&domain.Deallocate{}, svc.Deallocate)
	internalMp.RegisterCommandHandler(&domain.ChangeBatchQuantity{}, svc.ChangeBatchQuantity)
	internalMp.RegisterEventHandler(&domain.Allocated{}, externalMp.PublishEvent)
	internalMp.RegisterEventHandler(&domain.BatchQuantityChangedRealocationIsNeeded{}, svc.AllocationIsNeeded)

	InitializeExternalMessageConsumer(ctx, pubsub, internalMp)

	router := InitializeRouter(hnr)
	return router
}

func InitializeExternalMessageConsumer(
	ctx context.Context,
	pubsub *redis.Client,
	imp *messagepublisher.MessagePublisher,
) {
	externalMc := handler.NewExternalMessageConsumer(pubsub, imp)

	// add here new consumers for external messages
	go externalMc.ConsumeChangeBatchQuantityCommand(ctx)
}
