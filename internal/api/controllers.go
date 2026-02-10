package api

import (
	"ovmsa-be/internal/api/v1.0/ovmsa/auth"
	"ovmsa-be/internal/api/v1.0/ovmsa/org"
	"ovmsa-be/internal/services"
)

// InitializeControllers sets up all controllers with their required services
// This is called once during application startup
func InitializeControllers(svc *services.Services) {
	// Auth controllers
	auth.SetAuthService(svc.Auth)
	
	// Org controllers
	org.SetOrganizationService(svc.Organization)
}
