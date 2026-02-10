package auth

import (
	"errors"
	"ovmsa-be/internal/entities"
	"ovmsa-be/internal/services/auth"
	"ovmsa-be/pkg/helpers"
	"ovmsa-be/pkg/response"

	"github.com/gin-gonic/gin"
)

var authService *auth.AuthService

// SetAuthService sets the auth service for the auth controllers
func SetAuthService(svc *auth.AuthService) {
	authService = svc
}

// LoginHandler handles POST /auth/login
func LoginHandler(ctx *gin.Context, payload entities.TValidatedPayload, jwtData *entities.TJwtData, params entities.TParams) (any, error, error) {
	req, err := helpers.ExtractPayload[LoginRequest](payload)
	if err != nil {
		return nil, err, nil
	}

	// Get client IP and user agent for session tracking
	ipAddress := ctx.ClientIP()
	userAgent := ctx.Request.UserAgent()

	// Prepare login input
	input := auth.LoginInput{
		Email:    req.Email,
		Password: req.Password,
		OrgID:    req.OrgID,
		Audience: req.Audience,
	}

	// Perform login
	result, err := authService.Login(ctx.Request.Context(), input, ipAddress, userAgent)
	if err != nil {
		// Check for specific errors
		if errors.Is(err, auth.ErrInvalidCredentials) {
			response.UnauthorizedResponse(ctx, err, "Invalid email or password")
			ctx.Abort()
			return nil, nil, nil
		}
		if errors.Is(err, auth.ErrNoOrganizationAccess) {
			response.ForbiddenResponse(ctx, err, "User has no organization access")
			ctx.Abort()
			return nil, nil, nil
		}
		return nil, err, nil
	}

	return result, nil, nil
}

// LogoutHandler handles POST /auth/logout
func LogoutHandler(ctx *gin.Context, payload entities.TValidatedPayload, jwtData *entities.TJwtData, params entities.TParams) (any, error, error) {
	// Get identity from context (set by middleware)
	id, err := helpers.MustGetIdentity(ctx)
	if err != nil {
		response.UnauthorizedResponse(ctx, err, "Unauthorized")
		ctx.Abort()
		return nil, nil, nil
	}

	// Logout the current session
	if err := authService.Logout(ctx.Request.Context(), id.SessionID); err != nil {
		if errors.Is(err, auth.ErrSessionNotFound) {
			response.NotFoundResponse(ctx, err, "Session not found")
			ctx.Abort()
			return nil, nil, nil
		}
		return nil, err, nil
	}

	return map[string]string{
		"message": "Logged out successfully",
	}, nil, nil
}

// LogoutAllHandler handles POST /auth/logout-all
func LogoutAllHandler(ctx *gin.Context, payload entities.TValidatedPayload, jwtData *entities.TJwtData, params entities.TParams) (any, error, error) {
	// Get identity from context (set by middleware)
	id, err := helpers.MustGetIdentity(ctx)
	if err != nil {
		response.UnauthorizedResponse(ctx, err, "Unauthorized")
		ctx.Abort()
		return nil, nil, nil
	}

	// Logout all sessions for the user
	if err := authService.LogoutAll(ctx.Request.Context(), id.UserID); err != nil {
		return nil, err, nil
	}

	return map[string]string{
		"message": "All sessions logged out successfully",
	}, nil, nil
}

// RefreshTokenHandler handles POST /auth/refresh
func RefreshTokenHandler(ctx *gin.Context, payload entities.TValidatedPayload, jwtData *entities.TJwtData, params entities.TParams) (any, error, error) {
	req, err := helpers.ExtractPayload[RefreshTokenRequest](payload)
	if err != nil {
		return nil, err, nil
	}

	// Refresh the token
	result, err := authService.RefreshToken(ctx.Request.Context(), req.RefreshToken, req.Audience)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidRefreshToken) {
			response.UnauthorizedResponse(ctx, err, "Invalid or expired refresh token")
			ctx.Abort()
			return nil, nil, nil
		}
		return nil, err, nil
	}

	return result, nil, nil
}

// GetCurrentUserHandler handles GET /auth/me
func GetCurrentUserHandler(ctx *gin.Context, payload entities.TValidatedPayload, jwtData *entities.TJwtData, params entities.TParams) (any, error, error) {
	// Get identity from context (set by middleware)
	id, err := helpers.MustGetIdentity(ctx)
	if err != nil {
		response.UnauthorizedResponse(ctx, err, "Unauthorized")
		ctx.Abort()
		return nil, nil, nil
	}

	// Get user information
	userInfo, err := authService.GetCurrentUser(ctx.Request.Context(), id)
	if err != nil {
		if errors.Is(err, auth.ErrUserNotFound) {
			response.NotFoundResponse(ctx, err, "User not found")
			ctx.Abort()
			return nil, nil, nil
		}
		return nil, err, nil
	}

	return userInfo, nil, nil
}
