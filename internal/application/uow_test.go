package application

import (
	"context"
	"errors"
	"os"
	"testing"

	"cosmic-go/internal/domain"
	"cosmic-go/internal/infra/database"
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

		_ = uow.Transact(ctx, func(adapters Adapters) error {
			err := adapters.Repository.AddBatch(ctx, domain.NewBatch(
				"batch-001",
				domain.Product{SKU: "ROUND-TABLE"},
				10,
				nil,
			))
			assert.NoError(t, err)

			batches, err := adapters.Repository.ListBatches(ctx)
			assert.Len(t, batches, 1)
			assert.NoError(t, err)
			return err
		})

		ensureTransactionWasCommited := func() error {
			return uow.Transact(ctx, func(adapters Adapters) error {
				batches, err := adapters.Repository.ListBatches(ctx)
				assert.Len(t, batches, 1)
				assert.NoError(t, err)
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

		expectedErr := uow.Transact(ctx, func(adapters Adapters) error {
			err := adapters.Repository.AddBatch(ctx, domain.NewBatch(
				"batch-001",
				domain.Product{SKU: "ROUND-TABLE"},
				10,
				nil,
			))
			assert.NoError(t, err)

			batches, err := adapters.Repository.ListBatches(ctx)
			assert.Len(t, batches, 1)
			assert.NoError(t, err)

			return errors.New("must rollback this transaction because err is not nil")
		})
		assert.Error(t, expectedErr)

		ensureTransactionWasRolledBack := func() error {
			return uow.Transact(ctx, func(adapters Adapters) error {
				batches, err := adapters.Repository.ListBatches(ctx)
				assert.Len(t, batches, 0)
				return err
			})
		}

		err := ensureTransactionWasRolledBack()
		assert.NoError(t, err)

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
