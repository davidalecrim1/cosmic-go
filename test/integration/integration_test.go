//go:build integration

package integration

import (
	"context"
	"errors"
	"log"
	"os"
	"sync"
	"testing"
	"time"

	"cosmic-go/internal/domain"
	"cosmic-go/internal/infra/database"
	messagepublisher "cosmic-go/internal/infra/messagepublisher"

	"cosmic-go/internal/infra/repository"
	unitofwork "cosmic-go/internal/uow"
	"cosmic-go/test/helpers"

	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	testcontainerPostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/gorm"
)

var (
	db  *gorm.DB
	imp *messagepublisher.InternalMessagePublisher
)

func TestMain(m *testing.M) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dbContainer := createTestDatabase(ctx)
	defer func() {
		err := dbContainer.Terminate(ctx)
		if err != nil {
			log.Fatalf("failed to terminate container: %v", err)
		}
	}()
	conn, err := dbContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Fatalf("failed to retrieve the connection string: %v", err)
	}

	db = database.NewDatabase(
		database.WithConnectionString(conn),
	)

	imp = messagepublisher.NewInternalMessagePublisher()

	code := m.Run()
	os.Exit(code)
}

func createTestDatabase(ctx context.Context) (container *testcontainerPostgres.PostgresContainer) {
	pgContainer, err := testcontainerPostgres.Run(
		ctx,
		"docker.io/postgres:16.4-alpine3.20",
		testcontainerPostgres.WithDatabase("cosmic"),
		testcontainerPostgres.WithUsername("admin"),
		testcontainerPostgres.WithPassword("password"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)
	if err != nil {
		log.Fatalf("failed to create database for tests: %v", err)
	}
	return pgContainer
}

func TestUnitOfWork(t *testing.T) {
	t.Run("run a valid transaction with uow on repository", func(t *testing.T) {
		t.Cleanup(func() {
			helpers.CleanUpRepositoryHelper(db)
		})

		ctx := context.Background()
		uow := unitofwork.NewAllocationUnitOfWork(db, imp)

		sku := "ROUND-TABLE"
		product := domain.NewProduct(sku, []*domain.Batch{domain.NewBatch(
			"batch-001",
			sku,
			10,
			nil,
		)}, 0)

		_ = uow.Transact(ctx, func(adapters unitofwork.Adapters) error {
			err := adapters.Repository.AddProduct(ctx, product)
			assert.NoError(t, err)

			resultedProduct, err := adapters.Repository.GetProduct(ctx, sku)
			assert.Len(t, resultedProduct.Batches, 1)
			assert.NoError(t, err)
			return err
		})

		ensureTransactionWasCommited := func() error {
			return uow.Transact(ctx, func(adapters unitofwork.Adapters) error {
				product, err := adapters.Repository.GetProduct(ctx, sku)
				assert.NoError(t, err)
				assert.Len(t, product.Batches, 1)
				return err
			})
		}

		err := ensureTransactionWasCommited()
		assert.NoError(t, err)
	})

	t.Run("run a transaction that results in rollback", func(t *testing.T) {
		t.Cleanup(func() {
			helpers.CleanUpRepositoryHelper(db)
		})

		ctx := context.Background()
		uow := unitofwork.NewAllocationUnitOfWork(db, imp)

		sku := "ROUND-TABLE"
		product := domain.NewProduct(sku, []*domain.Batch{domain.NewBatch(
			"batch-001",
			sku,
			10,
			nil,
		)}, 0)

		expectedErr := uow.Transact(ctx, func(adapters unitofwork.Adapters) error {
			err := adapters.Repository.AddProduct(ctx, product)
			assert.NoError(t, err)

			product, err := adapters.Repository.GetProduct(ctx, sku)
			assert.NoError(t, err)
			assert.Len(t, product.Batches, 1)

			return errors.New("must rollback this transaction because err is not nil")
		})
		assert.Error(t, expectedErr)

		ensureTransactionWasRolledBack := func() error {
			return uow.Transact(ctx, func(adapters unitofwork.Adapters) error {
				product, err := adapters.Repository.GetProduct(ctx, sku)
				assert.Nil(t, product)
				return err
			})
		}

		err := ensureTransactionWasRolledBack()
		assert.ErrorIs(t, err, domain.ErrProductNotFound)
	})

	t.Run("similate concorrent updates to version not allowed",
		func(t *testing.T) {
			t.Cleanup(func() {
				helpers.CleanUpRepositoryHelper(db)
			})

			ctx := context.Background()
			uow := unitofwork.NewAllocationUnitOfWork(db, imp)

			sku := "ROUND-TABLE"
			product := domain.NewProduct(sku, []*domain.Batch{domain.NewBatch(
				"batch-001",
				sku,
				100,
				nil,
			)}, 0)

			_ = uow.Transact(ctx, func(adapters unitofwork.Adapters) error {
				err := adapters.Repository.AddProduct(ctx, product)
				assert.NoError(t, err)
				return nil
			})

			var resultedProduct *domain.Product
			_ = uow.Transact(ctx, func(adapters unitofwork.Adapters) (err error) {
				resultedProduct, err = adapters.Repository.GetProduct(ctx, sku)
				assert.NoError(t, err)
				assert.Len(t, resultedProduct.Batches, 1)
				return nil
			})

			concorrentAllocateOperation := func(product domain.Product) error {
				return uow.Transact(ctx, func(adapters unitofwork.Adapters) error {
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

			_ = uow.Transact(ctx, func(adapters unitofwork.Adapters) error {
				product, err := adapters.Repository.GetProduct(ctx, sku)
				assert.NoError(t, err)

				assert.Equal(t, 90, product.Batches[0].AvailableQuantity())
				assert.Equal(t, 1, product.VersionId, "ensure that it was updated only once")

				return nil
			})
		})
}

func TestRepository(t *testing.T) {
	t.Run("add a product",
		func(t *testing.T) {
			ctx := context.Background()

			tx := db.WithContext(ctx).Begin()
			assert.NoError(t, tx.Error)
			repoCreate := repository.NewPostgresRepository(tx)

			sku := "SMALL-TABLE"
			batch := domain.NewBatch("batch-001", sku, 10, nil)
			product := domain.NewProduct(sku, []*domain.Batch{batch}, 0)

			err := repoCreate.AddProduct(ctx, product)
			assert.NoError(t, err)

			resultedProduct, err := repoCreate.GetProduct(ctx, sku)
			assert.NoError(t, err)

			assert.Equal(t, product, resultedProduct)

			t.Cleanup(func() {
				tx.WithContext(ctx).Commit()
				helpers.CleanUpRepositoryHelper(db)
			})
		})

	t.Run("get a product with batches and allocations",
		func(t *testing.T) {
			ctx := context.Background()
			tx := db.WithContext(ctx).Begin()
			assert.NoError(t, tx.Error)
			repoCreate := repository.NewPostgresRepository(tx)

			sku := "SMALL-TABLE"
			eta := time.Now()
			initialBatchReference := "batch-001"
			initialBatch := domain.NewBatch(
				initialBatchReference,
				sku,
				50,
				&eta,
			)

			firstOrder := &domain.OrderLine{
				OrderId:  "order-001",
				SKU:      sku,
				Quantity: 25,
			}

			secondOrder := &domain.OrderLine{
				OrderId:  "order-002",
				SKU:      sku,
				Quantity: 25,
			}

			err := initialBatch.Allocate(firstOrder)
			assert.NoError(t, err)

			err = initialBatch.Allocate(secondOrder)
			assert.NoError(t, err)

			product := domain.NewProduct(sku, []*domain.Batch{initialBatch}, 0)

			err = repoCreate.AddProduct(ctx, product)
			assert.NoError(t, err)

			resultedProduct, err := repoCreate.GetProduct(ctx, sku)
			assert.NoError(t, err)

			assert.EqualExportedValues(t, product, resultedProduct)

			t.Cleanup(func() {
				tx.WithContext(ctx).Commit()
				helpers.CleanUpRepositoryHelper(db)
			})
		})

	t.Run("get product with no allocations",
		func(t *testing.T) {
			ctx := context.Background()
			tx := db.WithContext(ctx).Begin()
			assert.NoError(t, tx.Error)
			repoCreate := repository.NewPostgresRepository(tx)

			createdBatches := 2
			sku := "SMALL-TABLE"

			etaOne := time.Now().Add(24 * time.Hour)
			batchOne := domain.NewBatch("batch-001", sku, 50, &etaOne)

			etaTwo := time.Now().Add(48 * time.Hour)
			batchTwo := domain.NewBatch("batch-002", sku, 100, &etaTwo)

			product := domain.NewProduct(sku, []*domain.Batch{batchOne, batchTwo}, 0)
			err := repoCreate.AddProduct(ctx, product)
			assert.NoError(t, err)

			resultedProduct, err := repoCreate.GetProduct(ctx, sku)
			assert.NoError(t, err)
			assert.Equal(t, len(resultedProduct.Batches), createdBatches)

			t.Cleanup(func() {
				tx.WithContext(ctx).Commit()
				helpers.CleanUpRepositoryHelper(db)
			})
		})

	t.Run("get product wth batch with allocations by sku",
		func(t *testing.T) {
			ctx := context.Background()
			tx := db.WithContext(ctx).Begin()
			assert.NoError(t, tx.Error)
			repoCreate := repository.NewPostgresRepository(tx)

			eta := time.Now()
			initialSKU := "SMALL-TABLE"
			initialBatch := domain.NewBatch(
				"batch-001",
				initialSKU,
				50,
				&eta,
			)

			firstOrder := &domain.OrderLine{
				SKU:      initialSKU,
				Quantity: 25,
				OrderId:  "order-001",
			}

			secondOrder := &domain.OrderLine{
				SKU:      initialSKU,
				Quantity: 25,
				OrderId:  "order-002",
			}

			err := initialBatch.Allocate(firstOrder)
			assert.NoError(t, err)

			err = initialBatch.Allocate(secondOrder)
			assert.NoError(t, err)

			product := domain.NewProduct(initialSKU, []*domain.Batch{initialBatch}, 0)

			err = repoCreate.AddProduct(ctx, product)
			assert.NoError(t, err)

			resultedProduct, err := repoCreate.GetProduct(ctx, initialSKU)
			assert.NoError(t, err)

			assert.NotEqual(t, resultedProduct, nil)
			assert.EqualExportedValues(t, product, resultedProduct)

			t.Cleanup(func() {
				tx.WithContext(ctx).Commit()
				helpers.CleanUpRepositoryHelper(db)
			})
		})

	t.Run("update a product with new allocation", func(t *testing.T) {
		ctx := context.Background()
		tx := db.WithContext(ctx).Begin()
		assert.NoError(t, tx.Error)
		repoCreate := repository.NewPostgresRepository(tx)

		eta := time.Now()

		initialProduct := domain.NewProduct("SMALL-TABLE", []*domain.Batch{
			domain.NewBatch("batch-001", "SMALL-TABLE", 50, nil),
			domain.NewBatch("batch-002", "SMALL-TABLE", 100, &eta),
		}, 0)

		err := repoCreate.AddProduct(ctx, initialProduct)
		assert.NoError(t, err)

		assert.NoError(t, tx.WithContext(ctx).Commit().Error)

		tx = db.WithContext(ctx).Begin()
		assert.NoError(t, tx.Error)
		repoUpdate := repository.NewPostgresRepository(tx)

		updatedProduct := *initialProduct

		_, err = updatedProduct.Allocate(&domain.OrderLine{
			OrderId:  "order-001",
			SKU:      "SMALL-TABLE",
			Quantity: 25,
		})
		assert.NoError(t, err)

		err = repoUpdate.UpdateProduct(ctx, &updatedProduct)
		assert.NoError(t, err)

		resultedProduct, err := repoUpdate.GetProduct(ctx, updatedProduct.SKU)
		assert.NoError(t, err)
		assert.EqualExportedValues(t, &updatedProduct, resultedProduct)

		t.Cleanup(func() {
			tx.WithContext(ctx).Commit()
			helpers.CleanUpRepositoryHelper(db)
		})
	})

	t.Run("update a product with deallocation", func(t *testing.T) {
		ctx := context.Background()
		tx := db.WithContext(ctx).Begin()
		assert.NoError(t, tx.Error)
		repoCreate := repository.NewPostgresRepository(tx)

		eta := time.Now()

		initialProduct := domain.NewProduct("SMALL-TABLE", []*domain.Batch{
			domain.NewBatch("batch-001", "SMALL-TABLE", 50, nil),
			domain.NewBatch("batch-002", "SMALL-TABLE", 100, &eta),
		}, 0)

		_, err := initialProduct.Allocate(&domain.OrderLine{
			OrderId:  "order-001",
			SKU:      "SMALL-TABLE",
			Quantity: 25,
		})
		assert.NoError(t, err)

		err = repoCreate.AddProduct(ctx, initialProduct)
		assert.NoError(t, err)
		err = tx.WithContext(ctx).Commit().Error
		assert.NoError(t, err)

		tx = db.WithContext(ctx).Begin()
		assert.NoError(t, tx.Error)
		repoUpdate := repository.NewPostgresRepository(tx)

		updatedProduct := initialProduct

		err = updatedProduct.Deallocate("order-001")
		assert.NoError(t, err)

		err = repoUpdate.UpdateProduct(ctx, updatedProduct)
		assert.NoError(t, err)

		resultedProduct, err := repoUpdate.GetProduct(ctx, initialProduct.SKU)
		assert.NoError(t, err)
		assert.EqualExportedValues(t, updatedProduct, resultedProduct)

		t.Cleanup(func() {
			tx.WithContext(ctx).Commit()
			helpers.CleanUpRepositoryHelper(db)
		})
	})

	t.Run("get product by batch reference with no allocations",
		func(t *testing.T) {
			ctx := context.Background()
			tx := db.WithContext(ctx).Begin()
			assert.NoError(t, tx.Error)
			repoCreate := repository.NewPostgresRepository(tx)

			sku := "SMALL-TABLE"

			etaOne := time.Now().Add(24 * time.Hour)
			batchOne := domain.NewBatch("batch-001", sku, 50, &etaOne)

			etaTwo := time.Now().Add(48 * time.Hour)
			batchTwo := domain.NewBatch("batch-002", sku, 100, &etaTwo)

			product := domain.NewProduct(sku, []*domain.Batch{batchOne, batchTwo}, 0)
			err := repoCreate.AddProduct(ctx, product)
			assert.NoError(t, err)

			resultedProduct, err := repoCreate.GetProductByBatchReference(ctx, batchOne.Reference)
			assert.NoError(t, err)

			foundBatch := false
			for _, batch := range resultedProduct.Batches {
				if batch.Reference == batchOne.Reference {
					foundBatch = true
					break
				}
			}
			assert.True(t, foundBatch)
			assert.Len(t, resultedProduct.Batches, 2)

			t.Cleanup(func() {
				tx.WithContext(ctx).Commit()
				helpers.CleanUpRepositoryHelper(db)
			})
		})
}
