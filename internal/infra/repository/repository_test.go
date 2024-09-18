package repository

import (
	"context"
	"cosmic-go/internal/domain"
	"cosmic-go/pkg/env"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
)

func TestRepository(t *testing.T) {
	db, repo := newPostgresRepositoryHelper(t)

	t.Run("repository can save a batch",
		func(t *testing.T) {
			batch := domain.NewBatchWithoutETA("batch-001", domain.Product{SKU: "SMALL-TABLE"}, 10)
			err := repo.Add(batch)
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

			cleanUpRepository(db)
		})

	t.Run("repository can retrieve a batch with allocations",
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

			_, err = repo.Get("batch-001")
			assert.NoError(t, err)

			cleanUpRepository(db)
		})
}

func newPostgresRepositoryHelper(t *testing.T) (*pgxpool.Pool, *PostgresRepository) {
	t.Helper()

	dbEndpoint := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		env.GetEnvOrSetDefault("DB_USER", "admin"),
		env.GetEnvOrSetDefault("DB_PASSWORD", "password"),
		env.GetEnvOrSetDefault("DB_HOST", "localhost"),
		env.GetEnvOrSetDefault("DB_PORT", "5432"),
		env.GetEnvOrSetDefault("DB_SCHEMA", "cosmic"))

	db, err := pgxpool.New(context.Background(), dbEndpoint)
	if err != nil {
		log.Fatalf("error loading database configuration: %v", err)
	}

	return db, NewPostgresRepository(db)
}

func cleanUpRepository(db *pgxpool.Pool) {
	query := `
		DELETE FROM allocations;
		DELETE FROM order_lines;
		DELETE FROM batches;
		DELETE FROM products;
		`
	_, err := db.Exec(context.Background(), query)
	if err != nil {
		log.Fatalf("error cleaning up database: %v", err)
	}
}

type FakeRepository struct {
	batches map[string]*domain.Batch
}

func (r *FakeRepository) Add(batch *domain.Batch) error {
	r.batches[string(batch.Reference)] = batch
	return nil
}

func (r *FakeRepository) Get(reference string) (*domain.Batch, error) {
	return r.batches[reference], nil
}
