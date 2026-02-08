package database

import (
	"fmt"
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
	id, ok := db.Statement.Context.Value("identity").(*entities.Identity)
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

	db.Exec(fmt.Sprintf("SET LOCAL app.current_org_id = '%s'", id.OrgID))
	db.Exec(fmt.Sprintf("SET LOCAL app.current_org_path = '%s'", path))
	db.Exec(fmt.Sprintf("SET LOCAL app.user_id = '%s'", id.UserID))
	db.Exec(fmt.Sprintf("SET LOCAL app.user_role = '%s'", id.Role))
	db.Exec(fmt.Sprintf("SET LOCAL app.is_root = '%s'", isRoot))
}
