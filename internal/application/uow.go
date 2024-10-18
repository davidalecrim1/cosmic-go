package application

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

type BatchUoW struct {
	db *gorm.DB
}

func NewBatchUnitOfWork(db *gorm.DB) *BatchUoW {
	return &BatchUoW{db: db}
}

func (u *BatchUoW) Transact(ctx context.Context, txFunc func(adapters Adapters) error) error {
	return u.db.Transaction(func(tx *gorm.DB) error {
		adapters := Adapters{
			Repository: repository.NewPostgresRepository(tx),
		}

		return txFunc(adapters)
	})
}
