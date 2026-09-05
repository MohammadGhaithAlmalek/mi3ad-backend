package database

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	//"mi3ad/internal/database/models"
)

// Connect opens a GORM connection to Postgres.
func Connect(databaseURL string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	return db, nil
}

// AutoMigrate registers every model's schema with GORM. Add each new
// model here as it's created - this is the single place that knows
// about the full schema.
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		// &models.User{},
		// &models.Appointment{}, // add when the appointments module exists
	)
}