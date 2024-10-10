//go:build integration

package repository

import (
	"context"
	"testing"
	"time"

	"cosmic-go/internal/domain"
	"cosmic-go/internal/infra/database"
	"cosmic-go/test/helpers"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
)

func TestRepository(t *testing.T) {
	db, repo := newRepositoryHelper()
	defer db.Close()

	t.Run("add a batch",
		func(t *testing.T) {
			ctx := context.Background()

			batch := domain.NewBatchWithoutETA("batch-001", domain.Product{SKU: "SMALL-TABLE"}, 10)
			err := repo.AddBatch(ctx, batch)
			assert.NoError(t, err)

			resultedBatch, err := repo.GetBatchByReference(ctx, batch.Reference)
			assert.NoError(t, err)

			assert.Equal(t, batch, resultedBatch)

			t.Cleanup(func() {
				helpers.CleanUpRepositoryHelper(db)
			})
		})

	t.Run("get a batch with allocations",
		func(t *testing.T) {
			ctx := context.Background()

			eta := time.Now()
			initialBatchReference := "batch-001"
			initialBatch := domain.NewBatch(
				initialBatchReference,
				domain.Product{SKU: "SMALL-TABLE"},
				50,
				&eta,
			)

			firstOrder := &domain.OrderLine{
				Product: domain.Product{
					SKU: "SMALL-TABLE",
				},
				Quantity: 25,
				OrderId:  "order-001",
			}

			secondOrder := &domain.OrderLine{
				Product: domain.Product{
					SKU: "SMALL-TABLE",
				},
				Quantity: 25,
				OrderId:  "order-002",
			}

			err := initialBatch.Allocate(firstOrder)
			assert.NoError(t, err)

			err = initialBatch.Allocate(secondOrder)
			assert.NoError(t, err)

			err = repo.AddBatch(ctx, initialBatch)
			assert.NoError(t, err)

			resultedBatch, err := repo.GetBatchByReference(ctx, initialBatchReference)
			assert.NoError(t, err)

			assert.Equal(t, initialBatch.Reference, resultedBatch.Reference)
			assert.Equal(t, initialBatch.Product, resultedBatch.Product)
			assert.Equal(t, initialBatch.PurchasedQuantity, resultedBatch.PurchasedQuantity)
			assert.Equal(t, initialBatch.Allocations, resultedBatch.Allocations)
			assert.Equal(t, initialBatch.GetETA().Format(time.RFC3339), resultedBatch.GetETA().Format(time.RFC3339))

			t.Cleanup(func() {
				helpers.CleanUpRepositoryHelper(db)
			})
		})

	t.Run("list batches that have no allocations",
		func(t *testing.T) {
			ctx := context.Background()
			createdBatches := 2

			etaOne := time.Now().Add(24 * time.Hour)
			batchOne := domain.NewBatch("batch-001", domain.Product{SKU: "SMALL-TABLE"}, 50, &etaOne)

			etaTwo := time.Now().Add(48 * time.Hour)
			batchTwo := domain.NewBatch("batch-002", domain.Product{SKU: "LARGE-TABLE"}, 100, &etaTwo)

			err := repo.AddBatch(ctx, batchOne)
			assert.NoError(t, err)

			err = repo.AddBatch(ctx, batchTwo)
			assert.NoError(t, err)

			batches, err := repo.ListBatches(ctx)
			assert.NoError(t, err)
			assert.Equal(t, createdBatches, len(batches))

			t.Cleanup(func() {
				helpers.CleanUpRepositoryHelper(db)
			})
		})

	t.Run("get batch with allocations by sku",
		func(t *testing.T) {
			ctx := context.Background()

			eta := time.Now()
			initialSKU := "SMALL-TABLE"
			initialBatch := domain.NewBatch(
				"batch-001",
				domain.Product{SKU: initialSKU},
				50,
				&eta,
			)

			firstOrder := &domain.OrderLine{
				Product: domain.Product{
					SKU: initialSKU,
				},
				Quantity: 25,
				OrderId:  "order-001",
			}

			secondOrder := &domain.OrderLine{
				Product: domain.Product{
					SKU: initialSKU,
				},
				Quantity: 25,
				OrderId:  "order-002",
			}

			err := initialBatch.Allocate(firstOrder)
			assert.NoError(t, err)

			err = initialBatch.Allocate(secondOrder)
			assert.NoError(t, err)

			err = repo.AddBatch(ctx, initialBatch)
			assert.NoError(t, err)

			resultedBatch, err := repo.GetBatchBySku(ctx, initialSKU)
			assert.NoError(t, err)

			assert.NotEqual(t, resultedBatch, nil)
			assert.Equal(t, initialBatch.Reference, resultedBatch.Reference)
			assert.Equal(t, initialBatch.Product, resultedBatch.Product)
			assert.Equal(t, initialBatch.PurchasedQuantity, resultedBatch.PurchasedQuantity)
			assert.Equal(t, initialBatch.Allocations, resultedBatch.Allocations)
			assert.Equal(t, initialBatch.GetETA().Format(time.RFC3339), resultedBatch.GetETA().Format(time.RFC3339))

			t.Cleanup(func() {
				helpers.CleanUpRepositoryHelper(db)
			})
		})
}

func newRepositoryHelper() (*pgxpool.Pool, *PostgresRepository) {
	db := database.InitializeDatabase()
	return db, NewPostgresRepository(db)
}
