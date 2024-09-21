package helpers

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func CleanUpRepositoryHelper(db *pgxpool.Pool) {
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
