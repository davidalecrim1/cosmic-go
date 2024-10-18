package application

import (
	"context"
	"errors"
	"os"
	"testing"

	"cosmic-go/internal/domain"
	"cosmic-go/internal/infra/database"
	"cosmic-go/internal/infra/repository"
	"cosmic-go/test/helpers"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
)

var db *pgxpool.Pool

func TestMain(m *testing.M) {
	db = database.InitializeDatabase()

	code := m.Run()
	os.Exit(code)
}

func TestUnitOfWork(t *testing.T) {
	t.Run("run a valid transaction with uow on repository", func(t *testing.T) {
		ctx := context.Background()
		uow := NewBatchUnitOfWork(db)

		sku := "ROUND-TABLE"
		product := domain.NewProduct(sku, []*domain.Batch{domain.NewBatch(
			"batch-001",
			sku,
			10,
			nil,
		)}, 0)

		_ = uow.Transact(ctx, func(adapters Adapters) error {
			err := adapters.Repository.AddProduct(ctx, product)
			assert.NoError(t, err)

			resultedProduct, err := adapters.Repository.GetProduct(ctx, sku)
			assert.Len(t, resultedProduct.Batches, 1)
			assert.NoError(t, err)
			return err
		})

		ensureTransactionWasCommited := func() error {
			return uow.Transact(ctx, func(adapters Adapters) error {
				product, err := adapters.Repository.GetProduct(ctx, sku)
				assert.NoError(t, err)
				assert.Len(t, product.Batches, 1)
				return err
			})
		}

		err := ensureTransactionWasCommited()
		assert.NoError(t, err)

		t.Cleanup(func() {
			helpers.CleanUpRepositoryHelper(db)
		})
	})

	t.Run("run a transaction that results in rollback", func(t *testing.T) {
		ctx := context.Background()
		uow := NewBatchUnitOfWork(db)

		sku := "ROUND-TABLE"
		product := domain.NewProduct(sku, []*domain.Batch{domain.NewBatch(
			"batch-001",
			sku,
			10,
			nil,
		)}, 0)

		expectedErr := uow.Transact(ctx, func(adapters Adapters) error {
			err := adapters.Repository.AddProduct(ctx, product)
			assert.NoError(t, err)

			product, err := adapters.Repository.GetProduct(ctx, sku)
			assert.NoError(t, err)
			assert.Len(t, product.Batches, 1)

			return errors.New("must rollback this transaction because err is not nil")
		})
		assert.Error(t, expectedErr)

		ensureTransactionWasRolledBack := func() error {
			return uow.Transact(ctx, func(adapters Adapters) error {
				product, err := adapters.Repository.GetProduct(ctx, sku)
				assert.Nil(t, product)
				return err
			})
		}

		err := ensureTransactionWasRolledBack()
		assert.ErrorIs(t, err, repository.ErrProductNotFound)

		t.Cleanup(func() {
			helpers.CleanUpRepositoryHelper(db)
		})
	})
}

type FakeUoW struct {
	adapters Adapters
}

func NewFakeUnitOfWorkFromRepository(repo Repository) *FakeUoW {
	return &FakeUoW{adapters: Adapters{Repository: repo}}
}

func (u *FakeUoW) Transact(ctx context.Context, txFunc func(_ Adapters) error) error {
	return txFunc(u.adapters)
}
