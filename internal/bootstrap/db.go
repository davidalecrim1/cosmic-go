package bootstrap

import (
	"context"
	"cosmic-go/pkg/env"
	"fmt"
	"log"

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

	return db
}
