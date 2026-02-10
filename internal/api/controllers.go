package api

import (
	"ovmsa-be/internal/api/v1.0/ovmsa/auth"
	"ovmsa-be/internal/services"
)

// InitializeControllers sets up all controllers with their required services
// This is called once during application startup
func InitializeControllers(svc *services.Services) {
	// Auth controllers
	auth.SetAuthService(svc.Auth)
	
	// Add more controller initializations here as you add services
	// Example:
	// user.SetUserService(svc.User)
	// org.SetOrgService(svc.Org)
}
