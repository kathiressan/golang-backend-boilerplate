package entities

// Membership links a User to an Organization with a specific Role.
type Membership struct {
	BaseEntity
	UserID     string `gorm:"index;type:varchar(26);not null" json:"user_id"`
	Role       string `gorm:"type:varchar(20);not null" json:"role"` // admin, editor, viewer
	IsOrgAdmin bool   `gorm:"default:false" json:"is_org_admin"`    // Power to manage the specific Org
}

// OrgGrant allows a user to access an organization without being a full member.
// Supports cross-organization administration.
// BaseEntity.OrgID refers to the TARGET organization the user is granted access to.
type OrgGrant struct {
	BaseEntity
	UserID    string `gorm:"index;type:varchar(26);not null" json:"user_id"`
	GrantType string `gorm:"type:varchar(20);not null" json:"grant_type"` // support, audit, temporary
}
