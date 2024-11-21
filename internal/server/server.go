package server

import (
	"context"
	"log"
	"net/http"

	"cosmic-go/internal/application"
	"cosmic-go/internal/domain"
	"cosmic-go/internal/handler"

	messagepublisher "cosmic-go/internal/infra/messagepublisher"
	unitofwork "cosmic-go/internal/uow"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Server struct {
	Router *http.ServeMux
}

func NewServer() *Server {
	router := http.NewServeMux()

	return &Server{
		Router: router,
	}
}

func (s *Server) InitializeDependencies(ctx context.Context, db *gorm.DB, redisClient *redis.Client) {
	imp := messagepublisher.NewInternalMessagePublisher()

	uow := unitofwork.NewAllocationUnitOfWork(db, imp)
	svc := application.NewAllocationService(uow)
	hnr := handler.NewAllocationHandler(imp)
	hnrv := handler.NewAllocationViewHandler(db)

	emp := messagepublisher.NewExternalMessagePublisher(redisClient)

	SetupInternalMessagePublisher(imp, svc, emp)
	StartExternalMessageConsumer(ctx, redisClient, imp)
	s.setupRoutes(hnr, hnrv)
}

func (s *Server) setupRoutes(
	h *handler.AllocationHandler,
	hv *handler.AllocationViewHandler,
) {
	s.Router.HandleFunc("POST /products/allocations/allocate", h.Allocate)
	s.Router.HandleFunc("POST /products/allocations/deallocate", h.Deallocate)
	s.Router.HandleFunc("POST /products", h.AddProduct)
	s.Router.HandleFunc("GET /products/allocations/{id}", hv.GetAllocation)
}

func (s *Server) Run() {
	err := http.ListenAndServe(":8080", s.Router)
	if err != nil {
		log.Fatalln("Server failed to start:", err)
	}
}

func StartExternalMessageConsumer(
	ctx context.Context,
	redisClient *redis.Client,
	imp *messagepublisher.InternalMessagePublisher,
) {
	emc := handler.NewExternalMessageConsumer(redisClient, imp)
	go emc.ConsumeChangeBatchQuantityCommand(ctx)
}

func SetupInternalMessagePublisher(
	imp *messagepublisher.InternalMessagePublisher,
	svc *application.AllocationService,
	emp *messagepublisher.ExternalMessagePublisher,
) {
	imp.RegisterCommandHandler(&domain.CreateProduct{}, svc.AddProduct)
	imp.RegisterCommandHandler(&domain.Allocate{}, svc.Allocate)
	imp.RegisterCommandHandler(&domain.Deallocate{}, svc.Deallocate)
	imp.RegisterCommandHandler(&domain.ChangeBatchQuantity{}, svc.ChangeBatchQuantity)

	imp.RegisterEventHandler(&domain.Allocated{}, emp.PublishEvent)
	imp.RegisterEventHandler(&domain.BatchQuantityChangedRealocationIsNeeded{}, svc.AllocationIsNeeded)
}
