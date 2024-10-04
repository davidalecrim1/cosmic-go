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
	query := `
		BEGIN;
			DELETE FROM allocations;
			DELETE FROM order_lines;
			DELETE FROM batches;
			DELETE FROM products;
		COMMIT;
		`
	_, err := db.Exec(ctx, query)
	if err != nil {
		log.Fatalf("error cleaning up database: %v", err)
	}
}
