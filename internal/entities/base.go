package entities

import (
	"context"
	"ovmsa-be/pkg/utils"
	"time"

	"gorm.io/gorm"
)

// GlobalBaseEntity provides standard fields for system-wide models (no multi-tenancy).
type GlobalBaseEntity struct {
	ID        string         `gorm:"primaryKey;type:varchar(26)" json:"id"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// BeforeCreate ensure GlobalBaseEntity has a ULID.
func (g *GlobalBaseEntity) BeforeCreate(tx *gorm.DB) error {
	if g.ID == "" {
		g.ID = utils.NewULID()
	}
	return nil
}

// BaseEntity provides the standard fields for tenanted enterprise models.
// It includes RLS metadata (OrgID, OrgPath).
type BaseEntity struct {
	GlobalBaseEntity
	OrgID   string `gorm:"index;type:varchar(26);not null" json:"org_id"`
	OrgPath string `gorm:"index;column:org_path;type:text" json:"org_path"` // Hierarchical path
}

func (b *BaseEntity) BeforeCreate(tx *gorm.DB) error {
	if b.OrgPath != "" && b.OrgPath[len(b.OrgPath)-1] != '/' {
		b.OrgPath += "/"
	}
	return b.GlobalBaseEntity.BeforeCreate(tx)
}

func (b *BaseEntity) BeforeUpdate(tx *gorm.DB) error {
	if b.OrgPath != "" && b.OrgPath[len(b.OrgPath)-1] != '/' {
		b.OrgPath += "/"
	}
	return nil
}


// ScopeByOrg is a GORM scope that can be used manually if RLS is not active.
// We use this as a second layer of defense.
func ScopeByOrg(orgID string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("org_id = ?", orgID)
	}
}

// ScopeByPath is a GORM scope for hierarchical access.
// It ensures the path ends with a delimiter for prefix safety.
func ScopeByPath(path string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if path != "" && path[len(path)-1] != '/' {
			path += "/"
		}
		return db.Where("org_path LIKE ?", path+"%")
	}
}

// GetIdentity extracts the Identity from the context.
func GetIdentity(ctx context.Context) *Identity {
	if id, ok := ctx.Value(IdentityCtxKey).(*Identity); ok {
		return id
	}
	return nil
}
