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
			batch := domain.NewBatchWithoutETA("batch-001", domain.Product{SKU: "SMALL-TABLE"}, 10)
			err := repo.AddBatch(batch)
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

			helpers.CleanUpRepositoryHelper(db)
		})

	t.Run("get a batch with allocations",
		func(t *testing.T) {
			tx, err := db.Begin(context.Background())
			assert.NoError(t, err)
			defer tx.Rollback(context.Background())

			query := `INSERT INTO products (sku)
			VALUES ($1);`
			_, err = db.Exec(context.Background(), query, "SMALL-TABLE")
			assert.NoError(t, err)

			var orderlineId int
			query = `INSERT INTO order_lines (product_sku, quantity, orderid)
			VALUES ($1, $2, $3) RETURNING id;`

			err = tx.QueryRow(context.Background(), query, "SMALL-TABLE", 10, "order-001").
				Scan(&orderlineId)
			assert.NoError(t, err)

			query = `INSERT INTO batches (reference, product_sku, purchased_quantity, eta)
			VALUES ($1, $2, $3, $4);`
			_, err = tx.Exec(context.Background(), query, "batch-001", "SMALL-TABLE", 20, time.Time{})
			assert.NoError(t, err)

			query = `INSERT INTO allocations (orderline_id, batch_reference)
			VALUES ($1, $2);`
			_, err = tx.Exec(context.Background(), query, orderlineId, "batch-001")
			assert.NoError(t, err)

			tx.Commit(context.Background())

			_, err = repo.GetBatch("batch-001")
			assert.NoError(t, err)

			helpers.CleanUpRepositoryHelper(db)
		})

	t.Run("list batches",
		func(t *testing.T) {
			query := `
			INSERT INTO products (sku)
			VALUES 
			('SMALL-TABLE'), 
			('LARGE-TABLE');
			`
			_, err := db.Exec(context.Background(), query)
			assert.NoError(t, err)

			query = `
			INSERT INTO batches (reference, product_sku, purchased_quantity, eta) 
			VALUES 
			('batch-001', 'SMALL-TABLE', 200, $1),
			('batch-002', 'LARGE-TABLE', 100, $1);
			`
			_, err = db.Exec(context.Background(), query, time.Time{})
			assert.NoError(t, err)

			batches, err := repo.ListBatches()
			assert.NoError(t, err)
			assert.Equal(t, 2, len(batches))

			helpers.CleanUpRepositoryHelper(db)
		})
}

func newRepositoryHelper() (*pgxpool.Pool, *PostgresRepository) {
	db := bootstrap.InitializeDatabase()
	return db, NewPostgresRepository(db)
}
