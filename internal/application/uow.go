package application

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
	mp     MessagePublisher
}

type (
	EventHandler   func(event domain.Event) error
	CommandHandler func(command domain.Command) error
)

type MessagePublisher interface {
	PublishEvent(event domain.Event) <-chan error
	PublishCommand(command domain.Command) <-chan error
	RegisterEventHandler(event domain.Event, handler EventHandler)
	RegisterCommandHandler(command domain.Command, handler CommandHandler)
}

func NewAllocationUnitOfWork(db *gorm.DB, mp MessagePublisher) *AllocationUoW {
	return &AllocationUoW{
		db: db,
		mp: mp,
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
		errChan := u.mp.PublishEvent(event)
		handleErr(errChan)
	}
}
