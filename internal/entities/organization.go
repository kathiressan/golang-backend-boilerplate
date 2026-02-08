package entities

import "gorm.io/gorm"

// Organization represents a business entity or a division within a hierarchy.
// It is the primary unit of multi-tenancy.
type Organization struct {
	BaseEntity
	Name     string  `gorm:"type:varchar(100);not null" json:"name"`
	ParentID *string `gorm:"type:varchar(26);index" json:"parent_id"`
	Tier     string  `gorm:"type:varchar(20);default:'free'" json:"tier"` // free, basic, enterprise
}

// BeforeCreate hook to ensure OrgID matches the record's own ID for root/level-1 orgs
func (o *Organization) BeforeCreate(tx *gorm.DB) error {
	if err := o.BaseEntity.BeforeCreate(tx); err != nil {
		return err
	}
	
	// For organizations, the OrgID is often its own ID if it's a top-level org
	// but can be the parent's OrgID if it's a sub-division.
	// This logic depends on whether you want "Self-Isolation" or "Parent-Isolation".
	if o.OrgID == "" {
		o.OrgID = o.ID
	}
	return nil
}
