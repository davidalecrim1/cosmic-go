package application

import (
	"context"
	"errors"

	"cosmic-go/internal/domain"
	"cosmic-go/internal/infra/repository"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	AddBatch(ctx context.Context, b *domain.Batch) error
	GetBatchByReference(ctx context.Context, batchRef string) (*domain.Batch, error)
	ListBatches(ctx context.Context) ([]*domain.Batch, error)
	GetBatchBySku(ctx context.Context, sku string) (*domain.Batch, error)
	UpdateBatch(ctx context.Context, existingB *domain.Batch, updatedB *domain.Batch) error
}

type Adapters struct {
	Repository Repository
}

type BatchUoW struct {
	db *pgxpool.Pool
}

func NewBatchUnitOfWork(db *pgxpool.Pool) *BatchUoW {
	return &BatchUoW{db: db}
}

func (u *BatchUoW) Transact(ctx context.Context, txFunc func(adapters Adapters) error) error {
	return runWithTransaction(ctx, u.db, func(tx pgx.Tx) error {
		adapters := Adapters{
			Repository: repository.NewPostgresRepository(tx),
		}

		return txFunc(adapters)
	})
}

func runWithTransaction(
	ctx context.Context,
	db *pgxpool.Pool,
	fn func(tx pgx.Tx) error,
) error {
	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}

	err = fn(tx)
	if err == nil {
		return tx.Commit(ctx)
	}

	rollbackErr := tx.Rollback(ctx)
	if rollbackErr != nil {
		return errors.Join(err, rollbackErr)
	}

	return err
}
