package services

import (
	"ovmsa-be/internal/services/auth"

	"gorm.io/gorm"
)

// Services holds all application services
type Services struct {
	Auth *auth.AuthService
}

// InitServices initializes all application services
func InitServices(db *gorm.DB) *Services {
	return &Services{
		Auth: auth.NewAuthService(db),
	}
}
