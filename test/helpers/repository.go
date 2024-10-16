package helpers

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func CleanUpRepositoryHelper(db *pgxpool.Pool) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	tx, err := db.Begin(ctx)
	if err != nil {
		log.Fatalf("error creating transaction: %v", err)
	}
	defer tx.Rollback(ctx)

	query := `
		DELETE FROM allocations;
		DELETE FROM order_lines;
		DELETE FROM batches;
		DELETE FROM products;
		`
	_, err = tx.Exec(ctx, query)
	if err != nil {
		log.Fatalf("error cleaning up database: %v", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		log.Fatalf("failed on commit: %v", err)
	}
}
