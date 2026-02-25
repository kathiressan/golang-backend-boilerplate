package database

import (
	"ovmsa-be/internal/entities"

	"gorm.io/gorm"
)

// RLSPlugin is a GORM plugin that enforces PostgreSQL Row Level Security.
type RLSPlugin struct{}

func (p *RLSPlugin) Name() string {
	return "rls_plugin"
}

// Initialize registers the callbacks for the plugin.
func (p *RLSPlugin) Initialize(db *gorm.DB) error {
	// Register before callbacks for all major operations
	db.Callback().Create().Before("gorm:create").Register("rls:before_create", p.before)
	db.Callback().Query().Before("gorm:query").Register("rls:before_query", p.before)
	db.Callback().Update().Before("gorm:update").Register("rls:before_update", p.before)
	db.Callback().Delete().Before("gorm:delete").Register("rls:before_delete", p.before)
	return nil
}

// rlsConfiguredKey is used to track if RLS has been configured for the current statement
const rlsConfiguredKey = "rls_plugin_configured"

// before is the function that injects SET LOCAL commands before the actual SQL runs.
func (p *RLSPlugin) before(db *gorm.DB) {
	if db.Error != nil || db.Statement.Context == nil {
		return
	}

	// Skip if already run for this statement (prevent duplicate SET commands in hooks)
	// sync.Map cannot be compared to nil, so we use Load to check
	if _, alreadySet := db.Statement.Settings.Load(rlsConfiguredKey); alreadySet {
		return
	}

	// Extract identity from context
	id, ok := db.Statement.Context.Value(entities.IdentityCtxKey).(*entities.Identity)
	if !ok || id == nil {
		return
	}

	isRoot := "false"
	if id.IsRoot {
		isRoot = "true"
	}

	// Apply session variables with safety delimiters
	// Use SET LOCAL within a transaction to ensure it only lasts for the current transaction
	path := id.OrgPath
	if path != "" && len(path) > 0 && path[len(path)-1] != '/' {
		path += "/"
	}

	// Execute SET LOCAL within the same transaction/scoped connection
	// Using db.Session to ensure the settings persist for this operation
	session := db.Session(&gorm.Session{})
	execInScope := func(query string, args ...any) {
		if err := session.Exec(query, args...).Error; err != nil {
			// Log but don't fail - let the actual query determine the outcome
			_ = err
		}
	}

	execInScope("SET LOCAL app.current_org_id TO ?", id.OrgID)
	execInScope("SET LOCAL app.current_org_path TO ?", path)
	execInScope("SET LOCAL app.user_id TO ?", id.UserID)
	execInScope("SET LOCAL app.user_role TO ?", id.Role)
	execInScope("SET LOCAL app.is_root TO ?", isRoot)

	// Mark as configured to prevent duplicate execution
	db.Statement.Settings.Store(rlsConfiguredKey, true)
}
