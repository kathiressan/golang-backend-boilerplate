package api

import (
	"ovmsa-be/internal/api/v1.0/ovmsa/auth"
	"ovmsa-be/internal/api/v1.0/ovmsa/org"
	authService "ovmsa-be/internal/services/auth"
	orgService "ovmsa-be/internal/services/org"
	"ovmsa-be/pkg/helpers"
)

// ControllerFactory creates and configures controllers with their dependencies
type ControllerFactory struct {
	authService *authService.AuthService
	orgService  *orgService.OrganizationService
}

// NewControllerFactory creates a new controller factory with the given services
func NewControllerFactory(authSvc *authService.AuthService, orgSvc *orgService.OrganizationService) *ControllerFactory {
	return &ControllerFactory{
		authService: authSvc,
		orgService:  orgSvc,
	}
}

// AuthController provides auth-related handlers with dependency injection
type AuthController struct {
	service    *authService.AuthService
	errorChain *helpers.ErrorHandlerChain
}

// NewAuthController creates a new auth controller with its dependencies
func (f *ControllerFactory) NewAuthController() *AuthController {
	// Initialize error handler chain for auth errors
	errorChain := helpers.NewErrorHandlerChain()
	errorChain.Add(helpers.NewSpecificErrorHandler(authService.ErrInvalidCredentials, 401, "Invalid email or password"))
	errorChain.Add(helpers.NewSpecificErrorHandler(authService.ErrNoOrganizationAccess, 403, "User has no organization access"))
	errorChain.Add(helpers.NewSpecificErrorHandler(authService.ErrSessionNotFound, 404, "Session not found"))
	errorChain.Add(helpers.NewSpecificErrorHandler(authService.ErrInvalidRefreshToken, 401, "Invalid or expired refresh token"))
	errorChain.Add(helpers.NewSpecificErrorHandler(authService.ErrUserNotFound, 404, "User not found"))

	return &AuthController{
		service:    f.authService,
		errorChain: errorChain,
	}
}

// GetService returns the auth service (for backward compatibility with existing handlers)
func (c *AuthController) GetService() *authService.AuthService {
	return c.service
}

// GetErrorChain returns the error handler chain
func (c *AuthController) GetErrorChain() *helpers.ErrorHandlerChain {
	return c.errorChain
}

// OrgController provides organization-related handlers with dependency injection
type OrgController struct {
	service    *orgService.OrganizationService
	errorChain *helpers.ErrorHandlerChain
}

// NewOrgController creates a new org controller with its dependencies
func (f *ControllerFactory) NewOrgController() *OrgController {
	// Initialize error handler chain for org errors
	errorChain := helpers.NewErrorHandlerChain()
	errorChain.Add(helpers.NewSpecificErrorHandler(orgService.ErrOrganizationAlreadyExists, 409, "Organization with this name already exists at this level"))

	return &OrgController{
		service:    f.orgService,
		errorChain: errorChain,
	}
}

// GetService returns the org service (for backward compatibility with existing handlers)
func (c *OrgController) GetService() *orgService.OrganizationService {
	return c.service
}

// GetErrorChain returns the error handler chain
func (c *OrgController) GetErrorChain() *helpers.ErrorHandlerChain {
	return c.errorChain
}

// InitializeControllersWithFactory initializes controllers using the factory pattern
// This replaces the old SetXXXService pattern with proper dependency injection
func InitializeControllersWithFactory(authSvc *authService.AuthService, orgSvc *orgService.OrganizationService) {
	factory := NewControllerFactory(authSvc, orgSvc)
	
	// Initialize auth controller
	authCtrl := factory.NewAuthController()
	auth.SetAuthService(authCtrl.GetService())
	auth.SetAuthErrorChain(authCtrl.GetErrorChain())
	
	// Initialize org controller
	orgCtrl := factory.NewOrgController()
	org.SetOrganizationService(orgCtrl.GetService())
	org.SetOrgErrorChain(orgCtrl.GetErrorChain())
}
