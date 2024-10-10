package database

import (
	"context"
	"fmt"
	"log"

	"cosmic-go/pkg/env"

	"github.com/jackc/pgx/v5/pgxpool"
)

func InitializeDatabase() *pgxpool.Pool {
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

	initializeSchema(db)
	return db
}

func initializeSchema(db *pgxpool.Pool) {
	ctx := context.Background()
	schema := `
	BEGIN;

	CREATE TABLE IF NOT EXISTS products (
		sku TEXT PRIMARY KEY
		);

	CREATE TABLE IF NOT EXISTS batches (
		reference TEXT PRIMARY KEY NOT NULL, 
		product_sku TEXT REFERENCES products(sku),
		purchased_quantity INT NOT NULL,
		eta TIMESTAMPTZ
		);

	CREATE TABLE IF NOT EXISTS order_lines (
		id SERIAL PRIMARY KEY NOT NULL, 
		product_sku TEXT REFERENCES products(sku),
		quantity INT NOT NULL,
		orderid TEXT
		);

	CREATE TABLE IF NOT EXISTS allocations (
		id SERIAL PRIMARY KEY NOT NULL, 
		orderline_id INT REFERENCES order_lines(id),
		batch_reference TEXT REFERENCES batches(reference)
		);

	COMMIT;
	`

	_, err := db.Exec(ctx, schema)
	if err != nil {
		log.Fatal("failed to initiliaze schema with error:", err)
	}
}
