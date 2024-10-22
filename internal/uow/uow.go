package unitofwork

import (
	"context"

	"cosmic-go/internal/domain"
	"cosmic-go/internal/infra/repository"

	"gorm.io/gorm"
)

type Repository interface {
	AddProduct(ctx context.Context, p *domain.Product) error
	GetProduct(ctx context.Context, sku string) (*domain.Product, error)
	UpdateProduct(ctx context.Context, p *domain.Product) error
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
	Publish(event domain.Event)
	RegisterHandler(event domain.Event, handler EventHandler)
}

type EventHandler interface {
	Handle(event domain.Event) error
}

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

func (u *AllocationUoW) AddEvent(event domain.Event) {
	u.events = append(u.events, event)
}

func (u *AllocationUoW) dispatchEvents() {
	for _, event := range u.events {
		go func(e domain.Event) {
			u.ep.Publish(e)
		}(event)
	}
	u.events = nil
}
