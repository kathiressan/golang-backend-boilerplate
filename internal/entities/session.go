package entities

import (
	"time"
)

// Session represents an active login session.
// It tracks refresh tokens and security metadata.
type Session struct {
	BaseEntity
	UserID       string    `gorm:"index;type:varchar(26);not null" json:"user_id"`
	RefreshToken string    `gorm:"uniqueIndex;not null" json:"-"`      // Should be a hashed value in production
	IPAddress    string    `gorm:"type:varchar(45)" json:"ip_address"` // Supports IPv6
	UserAgent    string    `gorm:"type:text" json:"user_agent"`
	ExpiresAt    time.Time `gorm:"index" json:"expires_at"`
}

// IsExpired checks if the session has passed its expiration window
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}
