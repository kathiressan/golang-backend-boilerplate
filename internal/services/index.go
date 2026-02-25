package services

import (
	"ovmsa-be/internal/services/auth"
	"ovmsa-be/internal/services/org"
	"ovmsa-be/pkg/jwt"

	"gorm.io/gorm"
)

// Services holds all application services
type Services struct {
	Auth         *auth.AuthService
	Organization *org.OrganizationService
}

// InitServices initializes all application services
func InitServices(db *gorm.DB, jwtManager jwt.JWTManager) *Services {
	return &Services{
		Auth:         auth.NewAuthService(db, jwtManager),
		Organization: org.NewOrganizationService(db),
	}
}
