package auth

import "ovmsa-be/internal/entities"

var RouteMatrices = []entities.TRouteMatrix{
	{
		Path:        "/login",
		Method:      "POST",
		ProtectedBy: entities.UNPROTECTED,
		Schema:      &LoginRequest{},
		Controller: entities.TController{
			Handler: LoginHandler,
		},
	},
	{
		Path:        "/logout",
		Method:      "POST",
		ProtectedBy: entities.JWT,
		Controller: entities.TController{
			Handler: LogoutHandler,
		},
	},
	{
		Path:        "/logout-all",
		Method:      "POST",
		ProtectedBy: entities.RBAC_AUTH,
		Permissions: &entities.RBACConfig{
			AllowedRoles: []string{"root"},
		},
		Controller: entities.TController{
			Handler: LogoutAllHandler,
		},
	},
	{
		Path:        "/refresh",
		Method:      "POST",
		ProtectedBy: entities.UNPROTECTED,
		Schema:      &RefreshTokenRequest{},
		Controller: entities.TController{
			Handler: RefreshTokenHandler,
		},
	},
	{
		Path:        "/me",
		Method:      "GET",
		ProtectedBy: entities.JWT,
		Controller: entities.TController{
			Handler: GetCurrentUserHandler,
		},
	},
}
