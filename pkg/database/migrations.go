package database

import (
	"fmt"
	"os"
	config "ovmsa-be/configs"
	"ovmsa-be/internal/entities"
	"ovmsa-be/pkg/password"

	"gorm.io/gorm"
)

// MigrationStep defines a single versioned database change
type MigrationStep struct {
	Version string
	Action  func(tx *gorm.DB) error
}

var Migrations = []MigrationStep{
	{
		Version: "v1.0.0-core-enterprise-rls",
		Action: func(tx *gorm.DB) error {
			// Initial Schema
			if err := tx.AutoMigrate(
				&entities.User{},
				&entities.Organization{},
				&entities.Membership{},
				&entities.OrgGrant{},
				&entities.Session{},
				&entities.SigningKey{},
			); err != nil {
				return err
			}

			// Global RLS Setup
			tenantedTables := []string{"organizations", "memberships", "org_grants", "sessions"}
			for _, table := range tenantedTables {
				if err := enableRLS(tx, table); err != nil {
					return err
				}
			}
			return nil
		},
	},
	{
		Version: "v1.0.1-seed-system-root-user",
		Action: func(tx *gorm.DB) error {
			// Ensure schema is updated before seeding
			if err := tx.AutoMigrate(&entities.User{}); err != nil {
				return err
			}

			// Check if root user already exists
			var existingUser entities.User
			if err := tx.Where("email = ?", "root@system.local").First(&existingUser).Error; err == nil {
				// Root user already exists, skip seeding
				return nil
			}

			// Seed the Root User as a pure SYSTEM ROOT user.
			// This user belongs to no specific org but has global bypass powers.
			// Password must be set via ROOT_PASSWORD environment variable
			rootPassword := os.Getenv("ROOT_PASSWORD")
			if rootPassword == "" {
				return fmt.Errorf("ROOT_PASSWORD environment variable must be set")
			}

			hashedPassword, err := password.HashPassword(rootPassword)
			if err != nil {
				return fmt.Errorf("failed to hash root password: %w", err)
			}

			adminUser := &entities.User{
				Name:         "System Root",
				Email:        "root@system.local",
				PasswordHash: hashedPassword,
				IsRoot:       true, // Global System Admin
			}
			return tx.Create(adminUser).Error
		},
	},
	{
		Version: "v1.0.2-seed-initial-signing-key",
		Action: func(tx *gorm.DB) error {
			// Ensure schema is updated before seeding
			if err := tx.AutoMigrate(&entities.SigningKey{}); err != nil {
				return err
			}
			// Seed initial key from environment variables if they exist
			// This provides a smooth transition from env-based to db-based keys.
			cfg := config.GetConfig()

			initialKey := &entities.SigningKey{
				Version:   "v1",
				Algorithm: cfg.JWTSigningMethod,
				IsActive:  true,
			}

			if cfg.JWTSigningMethod == "RS256" {
				initialKey.KeyData = cfg.JWTPrivateKey
				initialKey.PublicKey = cfg.JWTPublicKey
			} else {
				initialKey.KeyData = cfg.JWTSecret
			}

			if initialKey.KeyData == "" {
				// No key in env, just skip seeding
				return nil
			}

			return tx.Create(initialKey).Error
		},
	},
}

// enableRLS executes the raw SQL to enable RLS and create the tenant policy for a table.
func enableRLS(db *gorm.DB, tableName string) error {
	// 1. Enable RLS on the table
	if err := db.Exec(fmt.Sprintf("ALTER TABLE %s ENABLE ROW LEVEL SECURITY;", tableName)).Error; err != nil {
		return err
	}

	// 2. Create the unified Tenant Isolation Policy
	policySQL := fmt.Sprintf(`
		DO $$
		BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_policies WHERE tablename = '%s' AND policyname = 'tenant_isolation_policy') THEN
				CREATE POLICY tenant_isolation_policy ON %s
				USING (
					current_setting('app.is_root', true) = 'true' OR
					org_id = current_setting('app.current_org_id', true) OR
					org_path LIKE current_setting('app.current_org_path', true) || '%%'
				);
			END IF;
		END $$;
	`, tableName, tableName)

	return db.Exec(policySQL).Error
}
