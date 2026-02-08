package database

import (
	"context"
	"fmt"
	"ovmsa-be/internal/entities"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// DBConfig holds connection parameters
type DBConfig struct {
	Host        string
	Port        int
	User        string
	Password    string
	DBName      string
	SSLMode     string
	DatabaseURL string
	LogMode     logger.LogLevel
}

// Initialize opens a connection to PostgreSQL and sets up the RLS interceptor.
func Initialize(cfg DBConfig) (*gorm.DB, error) {
	dsn := cfg.DatabaseURL
	if dsn == "" {
		dsn = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
			cfg.Host, cfg.User, cfg.Password, cfg.DBName, cfg.Port, cfg.SSLMode)
	}

	logLevel := cfg.LogMode
	if logLevel == 0 {
		logLevel = logger.Silent
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, err
	}

	// Register our Enterprise RLS Plugin
	if err := db.Use(&RLSPlugin{}); err != nil {
		return nil, fmt.Errorf("failed to register RLS plugin: %w", err)
	}

	DB = db
	return db, nil
}

// ScopedDB takes a GORM DB (or transaction) and injects the identity into the context.
// The RLSPlugin will read this identity and set Postgres session variables.
func ScopedDB(db *gorm.DB, id *entities.Identity) *gorm.DB {
	if id != nil {
		ctx := context.WithValue(db.Statement.Context, "identity", id)
		return db.WithContext(ctx)
	}
	return db
}

// Migrate handles both GORM auto-migrations and custom SQL for RLS policies.
// It uses a distributed lock and tracking table to ensure it runs exactly once.
func Migrate() error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}

	// 1. Use a Postgres Advisory Lock to prevent concurrent migrations during startup
	// Lock ID: 123456789 (arbitrary 64-bit integer)
	if err := DB.Exec("SELECT pg_advisory_lock(123456789)").Error; err != nil {
		return fmt.Errorf("failed to acquire migration lock: %w", err)
	}
	defer DB.Exec("SELECT pg_advisory_unlock(123456789)")

	// 2. Ensure Migration Tracking table exists
	type MigrationRecord struct {
		ID        uint      `gorm:"primaryKey"`
		Version   string    `gorm:"uniqueIndex;not null"`
		AppliedAt time.Time `gorm:"autoCreateTime"`
	}
	if err := DB.AutoMigrate(&MigrationRecord{}); err != nil {
		return fmt.Errorf("failed to create migration tracking table: %w", err)
	}

	// 3. Run pending migrations in order
	for _, m := range Migrations {
		var count int64
		DB.Model(&MigrationRecord{}).Where("version = ?", m.Version).Count(&count)
		if count > 0 {
			continue // Skip already applied migrations
		}

		fmt.Printf("Applying database migration: %s\n", m.Version)
		err := DB.Transaction(func(tx *gorm.DB) error {
			if err := m.Action(tx); err != nil {
				return err
			}
			return tx.Create(&MigrationRecord{Version: m.Version}).Error
		})

		if err != nil {
			return fmt.Errorf("migration %s failed: %w", m.Version, err)
		}
	}

	return nil
}