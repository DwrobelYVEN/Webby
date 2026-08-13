package db

import (
	"log"

	"github.com/yven/backend/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Connect opens the Postgres connection via GORM. In production, swap
// AutoMigrate for versioned migrations (e.g. golang-migrate) — it's used
// here only to get local/dev environments running quickly.
func Connect(dsn string) *gorm.DB {
	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	return database
}

// AutoMigrate creates/updates tables for every model. Run once at
// startup in dev; run explicitly via a migrate step in CI/prod.
func AutoMigrate(database *gorm.DB) error {
	return database.AutoMigrate(
		&models.User{},
		&models.Volunteer{},
		&models.Organization{},
		&models.OrgMembership{},
		&models.Event{},
		&models.EventRegistration{},
		&models.ServiceLog{},
		&models.StateTransition{},
		&models.Verification{},
		&models.VSR{},
		&models.VSRExport{},
		&models.Case{},
		&models.CaseNote{},
		&models.CaseResolution{},
		&models.Policy{},
		&models.PolicyVersion{},
		&models.PolicyAcknowledgment{},
	)
}
