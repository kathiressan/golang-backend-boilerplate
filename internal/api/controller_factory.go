package api

import (
	"ovmsa-be/internal/api/v1.0/ovmsa/auth"
	"ovmsa-be/internal/api/v1.0/ovmsa/health"
	"ovmsa-be/internal/api/v1.0/ovmsa/org"
	"ovmsa-be/internal/services"
	authService "ovmsa-be/internal/services/auth"
	orgService "ovmsa-be/internal/services/org"
	"ovmsa-be/pkg/helpers"
)

// ControllerFactory creates and configures controllers with their dependencies
type ControllerFactory struct {
	services *services.Services
}

// NewControllerFactory creates a new controller factory with the given services
func NewControllerFactory(svc *services.Services) *ControllerFactory {
	return &ControllerFactory{
		services: svc,
	}
}

// PrepareAuthChain initializes the error handler chain for auth errors
func (f *ControllerFactory) PrepareAuthChain() *helpers.ErrorHandlerChain {
	return helpers.NewErrorHandlerChain().
		Add(helpers.NewSpecificErrorHandler(authService.ErrInvalidCredentials, 401, "Invalid email or password")).
		Add(helpers.NewSpecificErrorHandler(authService.ErrNoOrganizationAccess, 403, "User has no organization access")).
		Add(helpers.NewSpecificErrorHandler(authService.ErrSessionNotFound, 404, "Session not found")).
		Add(helpers.NewSpecificErrorHandler(authService.ErrInvalidRefreshToken, 401, "Invalid or expired refresh token")).
		Add(helpers.NewSpecificErrorHandler(authService.ErrUserNotFound, 404, "User not found"))
}

// PrepareOrgChain initializes the error handler chain for org errors
func (f *ControllerFactory) PrepareOrgChain() *helpers.ErrorHandlerChain {
	return helpers.NewErrorHandlerChain().
		Add(helpers.NewSpecificErrorHandler(orgService.ErrOrganizationAlreadyExists, 409, "Organization with this name already exists at this level"))
}

// PrepareHealthChain initializes the error handler chain for health errors
func (f *ControllerFactory) PrepareHealthChain() *helpers.ErrorHandlerChain {
	return helpers.NewErrorHandlerChain()
}

// InitializeControllersWithFactory initializes controllers using the factory pattern
// This replaces the old SetXXXService pattern with proper dependency injection
func InitializeControllersWithFactory(svc *services.Services) {
	factory := NewControllerFactory(svc)

	// Initialize auth module
	auth.SetAuthService(svc.Auth)
	auth.SetAuthErrorChain(factory.PrepareAuthChain())

	// Initialize org module
	org.SetOrganizationService(svc.Organization)
	org.SetOrgErrorChain(factory.PrepareOrgChain())

	// Initialize health module
	health.SetHealthErrorChain(factory.PrepareHealthChain())
}
