package database

import (
	"fmt"
	"log"

	"cosmic-go/pkg/env"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Database struct {
	conn string
}

type Option func(*Database)

func NewDatabase(opts ...Option) *gorm.DB {
	db := &Database{}
	for _, opt := range opts {
		opt(db)
	}

	if db.conn == "" {
		db.conn = buildConnectionString()
	}

	gormClient, err := gorm.Open(postgres.Open(db.conn), &gorm.Config{})
	if err != nil {
		log.Fatalf("error loading database configuration: %v", err)
	}

	initializeSchema(gormClient)
	return gormClient
}

func WithConnectionString(conn string) Option {
	return func(db *Database) {
		db.conn = conn
	}
}

func buildConnectionString() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		env.GetEnvOrSetDefault("DB_USER", "admin"),
		env.GetEnvOrSetDefault("DB_PASSWORD", "password"),
		env.GetEnvOrSetDefault("DB_HOST", "localhost"),
		env.GetEnvOrSetDefault("DB_PORT", "5432"),
		env.GetEnvOrSetDefault("DB_SCHEMA", "cosmic"))
}

func initializeSchema(gormClient *gorm.DB) {
	err := gormClient.AutoMigrate(&OrderLine{}, &Product{}, &Batch{}, &Allocation{})
	if err != nil {
		log.Fatalf("failed to migrate the database schema: %v", err)
	}
}
