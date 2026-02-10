package entities

import (
	"time"
)

// SigningKey represents a cryptographic key used for signing and validating JWTs.
// This allows for key rotation without application restarts.
type SigningKey struct {
	GlobalBaseEntity
	Version    string    `gorm:"uniqueIndex;not null;type:varchar(50)" json:"version"`
	Algorithm  string    `gorm:"type:varchar(10);not null" json:"algorithm"` // e.g., HS256, RS256
	KeyData    string    `gorm:"type:text;not null" json:"-"`                // Private key (PEM) or HMAC secret
	PublicKey  string    `gorm:"type:text" json:"public_key"`                // Public key (PEM) for asymmetric algos
	IsActive   bool      `gorm:"index;default:true" json:"is_active"`
	ExpiresAt  time.Time `gorm:"index" json:"expires_at"`
}

// IsExpired checks if the key has passed its expiration
func (k *SigningKey) IsExpired() bool {
	if k.ExpiresAt.IsZero() {
		return false
	}
	return time.Now().After(k.ExpiresAt)
}
