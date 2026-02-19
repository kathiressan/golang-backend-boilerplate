package auth

import (
	"ovmsa-be/internal/entities"
	serviceAuth "ovmsa-be/internal/services/auth"
	"ovmsa-be/pkg/helpers"
)

var RouteMatrices = []entities.TRouteMatrix{
	helpers.POST("/login").
		Unprotected().
		WithSchema(&LoginRequest{}).
		WithResponseSchema(&serviceAuth.LoginResult{}).
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
		WithResponseSchema(&serviceAuth.LoginResult{}).
		WithHandler(RefreshTokenHandler).
		Build(),

	helpers.GET("/me").
		ProtectedByJWT().
		WithResponseSchema(&serviceAuth.UserInfo{}).
		WithHandler(GetCurrentUserHandler).
		Build(),
}
