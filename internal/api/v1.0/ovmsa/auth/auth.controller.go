package auth

import (
	"ovmsa-be/internal/entities"
	"ovmsa-be/internal/services/auth"
	"ovmsa-be/pkg/helpers"
	"ovmsa-be/pkg/response"

	"github.com/gin-gonic/gin"
)

var authService *auth.AuthService
var authErrorChain *helpers.ErrorHandlerChain

// SetAuthService sets the auth service for the auth controllers
func SetAuthService(svc *auth.AuthService) {
	authService = svc
}

// SetAuthErrorChain sets the error handler chain for auth controllers
func SetAuthErrorChain(chain *helpers.ErrorHandlerChain) {
	authErrorChain = chain
}

// LoginHandler handles POST /auth/login
func LoginHandler(ctx *gin.Context, payload entities.TValidatedPayload, identity *entities.Identity, params entities.TParams) (any, error, error) {
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
		if handled, _ := helpers.HandleServiceError(ctx, err, authErrorChain); handled {
			return nil, nil, nil
		}
		return nil, err, nil
	}

	return result, nil, nil
}

// LogoutHandler handles POST /auth/logout
func LogoutHandler(ctx *gin.Context, payload entities.TValidatedPayload, identity *entities.Identity, params entities.TParams) (any, error, error) {
	// Get identity from context (set by middleware)
	id, err := helpers.MustGetIdentity(ctx)
	if err != nil {
		response.UnauthorizedResponse(ctx, err, "Unauthorized")
		ctx.Abort()
		return nil, nil, nil
	}

	// Logout the current session
	if err := authService.Logout(ctx.Request.Context(), id.SessionID); err != nil {
		if handled, _ := helpers.HandleServiceError(ctx, err, authErrorChain); handled {
			return nil, nil, nil
		}
		return nil, err, nil
	}

	return map[string]string{
		"message": "Logged out successfully",
	}, nil, nil
}

// LogoutAllHandler handles POST /auth/logout-all
func LogoutAllHandler(ctx *gin.Context, payload entities.TValidatedPayload, identity *entities.Identity, params entities.TParams) (any, error, error) {
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
func RefreshTokenHandler(ctx *gin.Context, payload entities.TValidatedPayload, identity *entities.Identity, params entities.TParams) (any, error, error) {
	req, err := helpers.ExtractPayload[RefreshTokenRequest](payload)
	if err != nil {
		return nil, err, nil
	}

	// Refresh the token
	result, err := authService.RefreshToken(ctx.Request.Context(), req.RefreshToken, req.Audience)
	if err != nil {
		if handled, _ := helpers.HandleServiceError(ctx, err, authErrorChain); handled {
			return nil, nil, nil
		}
		return nil, err, nil
	}

	return result, nil, nil
}

// GetCurrentUserHandler handles GET /auth/me
func GetCurrentUserHandler(ctx *gin.Context, payload entities.TValidatedPayload, identity *entities.Identity, params entities.TParams) (any, error, error) {
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
		if handled, _ := helpers.HandleServiceError(ctx, err, authErrorChain); handled {
			return nil, nil, nil
		}
		return nil, err, nil
	}

	return userInfo, nil, nil
}
