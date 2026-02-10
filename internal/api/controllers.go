package api

import (
	"ovmsa-be/internal/services"
)

// InitializeControllers sets up all controllers with their required services
// This is called once during application startup
func InitializeControllers(svc *services.Services) {
	// Use factory pattern for better dependency injection
	InitializeControllersWithFactory(svc.Auth, svc.Organization)
}
