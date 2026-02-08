package database

import (
	"context"
	"fmt"
	"ovmsa-be/internal/entities"

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
}

// Initialize opens a connection to PostgreSQL and sets up the RLS interceptor.
func Initialize(cfg DBConfig) (*gorm.DB, error) {
	dsn := cfg.DatabaseURL
	if dsn == "" {
		dsn = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
			cfg.Host, cfg.User, cfg.Password, cfg.DBName, cfg.Port, cfg.SSLMode)
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
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

// ScopedDB returns a DB instance that has the identity injected into the context.
// The RLSPlugin will automatically read this identity and set Postgres session variables.
func ScopedDB(ctx context.Context, id *entities.Identity) *gorm.DB {
	if id != nil {
		ctx = context.WithValue(ctx, "identity", id)
	}
	return DB.WithContext(ctx)
}
