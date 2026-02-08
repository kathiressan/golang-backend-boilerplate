package database

import (
	"fmt"
	"ovmsa-be/internal/entities"

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
}



// enableRLS executes the raw SQL to enable RLS and create the tenant policy for a table.
func enableRLS(db *gorm.DB, tableName string) error {
	// 1. Enable RLS on the table
	if err := db.Exec(fmt.Sprintf("ALTER TABLE %s ENABLE ROW LEVEL SECURITY;", tableName)).Error; err != nil {
		return err
	}

	// 2. Create the unified Tenant Isolation Policy
	// This policy allows access if:
	// - app.is_root is 'true' (Super Admin)
	// - The record's org_id matches app.current_org_id (Direct Tenancy)
	// - The record's org_path starts with app.current_org_path (Hierarchical Tenancy)
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