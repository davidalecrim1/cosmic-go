package helpers

import (
	"context"
	"log"
	"time"

	"gorm.io/gorm"
)

func CleanUpRepositoryHelper(db *gorm.DB) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	tx := db.WithContext(ctx).Begin()
	if tx.Error != nil {
		log.Fatalf("error creating transaction: %v", tx.Error)
	}
	defer tx.WithContext(ctx).Rollback()

	query := `
		DELETE FROM allocations;
		DELETE FROM order_lines;
		DELETE FROM batches;
		DELETE FROM products;
		`
	if err := tx.WithContext(ctx).Exec(query).Error; err != nil {
		log.Fatalf("error cleaning up database: %v", err)
	}

	if err := tx.WithContext(ctx).Commit().Error; err != nil {
		log.Fatalf("failed on commit: %v", err)
	}
}
