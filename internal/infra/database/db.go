package database

import (
	"fmt"
	"log"

	"cosmic-go/pkg/env"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitializeDatabase() *gorm.DB {
	dbEndpoint := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		env.GetEnvOrSetDefault("DB_USER", "admin"),
		env.GetEnvOrSetDefault("DB_PASSWORD", "password"),
		env.GetEnvOrSetDefault("DB_HOST", "localhost"),
		env.GetEnvOrSetDefault("DB_PORT", "5432"),
		env.GetEnvOrSetDefault("DB_SCHEMA", "cosmic"))

	db, err := gorm.Open(postgres.Open(dbEndpoint), &gorm.Config{})
	if err != nil {
		log.Fatalf("error loading database configuration: %v", err)
	}

	initializeSchema(db)
	return db
}

func initializeSchema(db *gorm.DB) {
	err := db.AutoMigrate(&OrderLine{}, &Product{}, &Batch{}, &Allocation{})
	if err != nil {
		log.Fatalf("failed to migrate the database schema: %v", err)
	}
}
