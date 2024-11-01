package unitofwork

import (
	"context"
	"log"

	"cosmic-go/internal/domain"
	"cosmic-go/internal/infra/repository"

	"gorm.io/gorm"
)

type Repository interface {
	AddProduct(ctx context.Context, p *domain.Product) error
	GetProduct(ctx context.Context, sku string) (*domain.Product, error)
	UpdateProduct(ctx context.Context, p *domain.Product) error
	GetProductByBatchReference(ctx context.Context, batchReference string) (*domain.Product, error)
}

type Adapters struct {
	Repository Repository
}

type AllocationUoW struct {
	db     *gorm.DB
	events []domain.Event
	ep     EventPublisher
}

type EventPublisher interface {
	Publish(event domain.Event) <-chan error
	RegisterHandler(event domain.Event, handler EventHandler)
}

type EventHandler func(event domain.Event) error

func NewAllocationUnitOfWork(db *gorm.DB, ep EventPublisher) *AllocationUoW {
	return &AllocationUoW{
		db: db,
		ep: ep,
	}
}

func (u *AllocationUoW) Transact(ctx context.Context, txFunc func(adapters Adapters) error) error {
	err := u.db.Transaction(func(tx *gorm.DB) error {
		adapters := Adapters{
			Repository: repository.NewPostgresRepository(tx),
		}

		return txFunc(adapters)
	})

	u.dispatchEvents()
	return err
}

func (u *AllocationUoW) AddEvents(events []domain.Event) {
	u.events = append(u.events, events...)
}

func (u *AllocationUoW) dispatchEvents() {
	events := u.events
	u.events = nil

	handleErr := func(errChan <-chan error) {
		for err := range errChan {
			if err != nil {
				log.Printf("error publishing event: %v", err)
			}
		}
	}

	for _, event := range events {
		errChan := u.ep.Publish(event)
		handleErr(errChan)
	}
}
