//go:build integration

package repository

import (
	"context"
	"os"
	"testing"
	"time"

	"cosmic-go/internal/domain"
	"cosmic-go/internal/infra/database"
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

func TestRepository(t *testing.T) {
	t.Run("add a product",
		func(t *testing.T) {
			ctx := context.Background()

			tx := db.WithContext(ctx).Begin()
			assert.NoError(t, tx.Error)
			repoCreate := NewPostgresRepository(tx)

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
			repoCreate := NewPostgresRepository(tx)

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
			repoCreate := NewPostgresRepository(tx)

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
			repoCreate := NewPostgresRepository(tx)

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
		repoCreate := NewPostgresRepository(tx)

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
		repoUpdate := NewPostgresRepository(tx)

		updatedProduct := *initialProduct

		updatedProduct.Allocate(&domain.OrderLine{
			OrderId:  "order-001",
			SKU:      "SMALL-TABLE",
			Quantity: 25,
		})

		err = repoUpdate.UpdateProduct(ctx, initialProduct)
		assert.NoError(t, err)

		resultedProduct, err := repoUpdate.GetProduct(ctx, initialProduct.SKU)
		assert.NoError(t, err)
		assert.EqualExportedValues(t, initialProduct, resultedProduct)

		t.Cleanup(func() {
			tx.WithContext(ctx).Commit()
			helpers.CleanUpRepositoryHelper(db)
		})
	})

	t.Run("update a product with deallocation", func(t *testing.T) {
		ctx := context.Background()
		tx := db.WithContext(ctx).Begin()
		assert.NoError(t, tx.Error)
		repoCreate := NewPostgresRepository(tx)

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
		repoUpdate := NewPostgresRepository(tx)

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
}
