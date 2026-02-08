package entities

// User represents a global person in the system.
// Users can belong to many organizations via Memberships.

type User struct {
	GlobalBaseEntity
	Name         string `gorm:"type:varchar(100)" json:"name"`
	Email        string `gorm:"uniqueIndex;type:varchar(100);not null" json:"email"`
	PasswordHash string `gorm:"not null" json:"-"`
	IsRoot       bool   `gorm:"default:false" json:"is_root"` // Global super-admin
}