package auth

import (
	"ovmsa-be/internal/entities"
	"ovmsa-be/pkg/helpers"
)

var RouteMatrices = []entities.TRouteMatrix{
	helpers.POST("/login").
		Unprotected().
		WithSchema(&LoginRequest{}).
		WithHandler(LoginHandler).
		Build(),

	helpers.POST("/logout").
		ProtectedByJWT().
		WithHandler(LogoutHandler).
		Build(),

	helpers.POST("/logout-all").
		ProtectedByRBAC("root").
		WithHandler(LogoutAllHandler).
		Build(),

	helpers.POST("/refresh").
		Unprotected().
		WithSchema(&RefreshTokenRequest{}).
		WithHandler(RefreshTokenHandler).
		Build(),

	helpers.GET("/me").
		ProtectedByJWT().
		WithHandler(GetCurrentUserHandler).
		Build(),
}
