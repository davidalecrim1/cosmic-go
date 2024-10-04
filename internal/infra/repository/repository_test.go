//go:build integration

package repository

import (
	"context"
	"cosmic-go/internal/bootstrap"
	"cosmic-go/internal/domain"
	"cosmic-go/test/helpers"
	"testing"
	"time"

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

			resultedBatch := &domain.Batch{}
			err = db.QueryRow(context.Background(), "SELECT * FROM batches;").
				Scan(&resultedBatch.Reference,
					&resultedBatch.Product.SKU,
					&resultedBatch.PurchasedQuantity,
					&resultedBatch.ETA)

			assert.NoError(t, err)
			assert.Equal(t, batch.Reference, resultedBatch.Reference)
			assert.Equal(t, batch.Product.SKU, resultedBatch.Product.SKU)
			assert.Equal(t, batch.PurchasedQuantity, resultedBatch.PurchasedQuantity)

			t.Cleanup(func() {
				helpers.CleanUpRepositoryHelper(db)
			})
		})

	t.Run("get a batch with allocations",
		func(t *testing.T) {
			ctx := context.Background()

			tx, err := db.Begin(ctx)
			assert.NoError(t, err)
			defer tx.Rollback(ctx)

			query := `INSERT INTO products (sku)
			VALUES ($1);`
			_, err = db.Exec(ctx, query, "SMALL-TABLE")
			assert.NoError(t, err)

			var orderlineId int
			query = `INSERT INTO order_lines (product_sku, quantity, orderid)
			VALUES ($1, $2, $3) RETURNING id;`

			err = tx.QueryRow(ctx, query, "SMALL-TABLE", 10, "order-001").
				Scan(&orderlineId)
			assert.NoError(t, err)

			query = `INSERT INTO batches (reference, product_sku, purchased_quantity, eta)
			VALUES ($1, $2, $3, $4);`
			_, err = tx.Exec(ctx, query, "batch-001", "SMALL-TABLE", 20, time.Time{})
			assert.NoError(t, err)

			query = `INSERT INTO allocations (orderline_id, batch_reference)
			VALUES ($1, $2);`
			_, err = tx.Exec(ctx, query, orderlineId, "batch-001")
			assert.NoError(t, err)

			tx.Commit(ctx)

			_, err = repo.GetBatchByReference(ctx, "batch-001")
			assert.NoError(t, err)

			t.Cleanup(func() {
				helpers.CleanUpRepositoryHelper(db)
			})
		})

	t.Run("list batches",
		func(t *testing.T) {
			ctx := context.Background()

			query := `
			INSERT INTO products (sku)
			VALUES 
			('SMALL-TABLE'), 
			('LARGE-TABLE');
			`
			_, err := db.Exec(ctx, query)
			assert.NoError(t, err)

			query = `
			INSERT INTO batches (reference, product_sku, purchased_quantity, eta) 
			VALUES 
			('batch-001', 'SMALL-TABLE', 200, $1),
			('batch-002', 'LARGE-TABLE', 100, $1);
			`
			_, err = db.Exec(ctx, query, time.Time{})
			assert.NoError(t, err)

			batches, err := repo.ListBatches(ctx)
			assert.NoError(t, err)
			assert.Equal(t, 2, len(batches))

			t.Cleanup(func() {
				helpers.CleanUpRepositoryHelper(db)
			})
		})

	t.Run("get batch with allocations by sku",
		func(t *testing.T) {
			ctx := context.Background()
			validSku := "SMALL-TABLE"

			tx, err := db.Begin(ctx)
			assert.NoError(t, err)
			defer tx.Rollback(ctx)

			query := `INSERT INTO products (sku)
			VALUES ($1);`
			_, err = db.Exec(ctx, query, validSku)
			assert.NoError(t, err)

			var orderlineId int
			query = `INSERT INTO order_lines (product_sku, quantity, orderid)
			VALUES ($1, $2, $3) RETURNING id;`

			err = tx.QueryRow(ctx, query, validSku, 10, "order-001").
				Scan(&orderlineId)
			assert.NoError(t, err)

			query = `INSERT INTO batches (reference, product_sku, purchased_quantity, eta)
			VALUES ($1, $2, $3, $4);`
			_, err = tx.Exec(ctx, query, "batch-001", validSku, 20, time.Time{})
			assert.NoError(t, err)

			query = `INSERT INTO allocations (orderline_id, batch_reference)
			VALUES ($1, $2);`
			_, err = tx.Exec(ctx, query, orderlineId, "batch-001")
			assert.NoError(t, err)

			tx.Commit(ctx)

			_, err = repo.getBatchBySku(ctx, validSku)
			assert.NoError(t, err)

			t.Cleanup(func() {
				helpers.CleanUpRepositoryHelper(db)
			})
		})
}

func newRepositoryHelper() (*pgxpool.Pool, *PostgresRepository) {
	db := bootstrap.InitializeDatabase()
	return db, NewPostgresRepository(db)
}
