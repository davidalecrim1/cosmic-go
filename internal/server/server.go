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

func InitializeServer(ctx context.Context, db *gorm.DB, redisClient *redis.Client) *http.ServeMux {
	imp := messagepublisher.NewInternalMessagePublisher()

	uow := unitofwork.NewAllocationUnitOfWork(db, imp)
	svc := application.NewAllocationService(uow)
	hnr := handler.NewAllocationHandler(imp)
	hnrv := handler.NewAllocationViewHandler(db)

	emp := messagepublisher.NewExternalMessagePublisher(redisClient)

	// add here new internal message publisher services
	imp.RegisterCommandHandler(&domain.CreateProduct{}, svc.AddProduct)
	imp.RegisterCommandHandler(&domain.Allocate{}, svc.Allocate)
	imp.RegisterCommandHandler(&domain.Deallocate{}, svc.Deallocate)
	imp.RegisterCommandHandler(&domain.ChangeBatchQuantity{}, svc.ChangeBatchQuantity)

	imp.RegisterEventHandler(&domain.Allocated{}, emp.PublishEvent)
	imp.RegisterEventHandler(&domain.BatchQuantityChangedRealocationIsNeeded{}, svc.AllocationIsNeeded)

	InitializeExternalMessageConsumer(ctx, redisClient, imp)

	router := InitializeRouter(hnr, hnrv)
	return router
}

func InitializeExternalMessageConsumer(
	ctx context.Context,
	redisClient *redis.Client,
	imp *messagepublisher.InternalMessagePublisher,
) {
	emc := handler.NewExternalMessageConsumer(redisClient, imp)

	// add here new consumers for external messages
	go emc.ConsumeChangeBatchQuantityCommand(ctx)
}
