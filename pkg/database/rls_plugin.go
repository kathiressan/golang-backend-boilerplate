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

// before is the function that injects SET LOCAL commands before the actual SQL runs.
func (p *RLSPlugin) before(db *gorm.DB) {
	if db.Error != nil || db.Statement.Context == nil {
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
	path := id.OrgPath
	if path != "" && path[len(path)-1] != '/' {
		path += "/"
	}

	db.Exec("SELECT set_config('app.current_org_id', ?, true)", id.OrgID)
	db.Exec("SELECT set_config('app.current_org_path', ?, true)", path)
	db.Exec("SELECT set_config('app.user_id', ?, true)", id.UserID)
	db.Exec("SELECT set_config('app.user_role', ?, true)", id.Role)
	db.Exec("SELECT set_config('app.is_root', ?, true)", isRoot)
}
