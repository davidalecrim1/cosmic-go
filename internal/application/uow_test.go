package application

import (
	"context"
	"errors"
	"os"
	"sync"
	"testing"

	"cosmic-go/internal/domain"
	"cosmic-go/internal/infra/database"
	"cosmic-go/internal/infra/repository"
	"cosmic-go/test/helpers"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

var db *gorm.DB

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

	t.Run("similate concorrent updates to version not allowed",
		func(t *testing.T) {
			helpers.CleanUpRepositoryHelper(db)
			ctx := context.Background()
			uow := NewBatchUnitOfWork(db)

			sku := "ROUND-TABLE"
			product := domain.NewProduct(sku, []*domain.Batch{domain.NewBatch(
				"batch-001",
				sku,
				100,
				nil,
			)}, 0)

			_ = uow.Transact(ctx, func(adapters Adapters) error {
				err := adapters.Repository.AddProduct(ctx, product)
				assert.NoError(t, err)
				return nil
			})

			var resultedProduct *domain.Product
			_ = uow.Transact(ctx, func(adapters Adapters) (err error) {
				resultedProduct, err = adapters.Repository.GetProduct(ctx, sku)
				assert.NoError(t, err)
				assert.Len(t, resultedProduct.Batches, 1)
				return nil
			})

			concorrentAllocateOperation := func(product domain.Product) error {
				return uow.Transact(ctx, func(adapters Adapters) error {
					_, err := product.Allocate(&domain.OrderLine{
						OrderId:  "order-001",
						SKU:      product.SKU,
						Quantity: 10,
					})
					assert.NoError(t, err)
					return adapters.Repository.UpdateProduct(ctx, &product)
				})
			}

			var wg sync.WaitGroup
			for i := 0; i < 30; i++ {
				wg.Add(1)

				go func(wg *sync.WaitGroup) {
					defer wg.Done()

					err := concorrentAllocateOperation(*resultedProduct)
					if err != nil {
						t.Logf("expected error on concurrent operation: %v", err)
					}
				}(&wg)
			}
			wg.Wait()

			_ = uow.Transact(ctx, func(adapters Adapters) error {
				product, err := adapters.Repository.GetProduct(ctx, sku)
				assert.NoError(t, err)

				assert.Equal(t, 90, product.Batches[0].AvailableQuantity())
				assert.Equal(t, 1, product.VersionId, "ensure that it was updated only once")

				return nil
			})

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
