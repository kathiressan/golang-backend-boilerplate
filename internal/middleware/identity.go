package middleware

import (
	"errors"
	config "ovmsa-be/configs"
	"ovmsa-be/internal/entities"
	"ovmsa-be/internal/repository"
	cryptographyHelper "ovmsa-be/pkg/cryptography"
	"ovmsa-be/pkg/jwt"
	"ovmsa-be/pkg/response"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// TokenContext represents the token-related data stored in the context for external services
type TokenContext struct {
	Token       string
	RequesterID string
	Requester   config.Requester
}

// AuthMiddleware is a unified authentication handler that dispatches to either
// external service validation or user identity extraction based on token format.
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		// Expected format: "Bearer <token>"
		authParts := strings.SplitN(authHeader, " ", 2)
		if len(authParts) != 2 || authParts[0] != "Bearer" {
			// If format is invalid, we don't abort here because some routes might be public.
			c.Next()
			return
		}

		token := authParts[1]
		tokenParts := strings.Split(token, ":")

		// Dispatch based on token content
		// If it has 3 parts and the first part is a registered external requester, handle as service token.
		if len(tokenParts) == 3 {
			if _, isRequester := config.ValidRequesters[tokenParts[0]]; isRequester {
				handleExternalServiceToken(c, tokenParts)
				c.Next()
				return
			}
		}

		// Otherwise, handle as standard user JWT
		handleUserIdentity(c, token)
		c.Next()
	}
}

// handleExternalServiceToken handles tokens from known external services with format:
// requesterID : unixTimestamp : signature
func handleExternalServiceToken(c *gin.Context, tokenParts []string) {
	requesterID := tokenParts[0]
	timestampStr := tokenParts[1] // Unix Timestamp
	signature := tokenParts[2]    // HMAC Signature

	// 1. Lookup Requester
	requester, exists := config.ValidRequesters[requesterID]
	if !exists {
		response.UnauthorizedResponse(c, nil, "Unauthorized: Unknown requester ID")
		c.Abort()
		return
	}

	// 2. Verify signature
	// Proves the sender knows the secret key
	expectedSignature := cryptographyHelper.Base64HMAC(timestampStr, requester.SecretKey)
	if expectedSignature != signature {
		response.UnauthorizedResponse(c, nil, "Unauthorized: Invalid token signature")
		c.Abort()
		return
	}

	// 3. Time window check (Replay Protection)
	// Parse the timestamp and ensure it's within the last 5 minutes.
	unixTime, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		response.UnauthorizedResponse(c, nil, "Unauthorized: Invalid timestamp format")
		c.Abort()
		return
	}

	tokenTime := time.Unix(unixTime, 0)
	timeDiff := time.Since(tokenTime)

	// Enforce 5 minute window (allowing for minor clock skew in both directions)
	if timeDiff > 5*time.Minute || timeDiff < -5*time.Minute {
		response.UnauthorizedResponse(c, nil, "Unauthorized: Token has expired or timestamp is invalid")
		c.Abort()
		return
	}

	// Store validated context
	c.Set("tokenContext", TokenContext{
		Token:       strings.Join(tokenParts, ":"),
		RequesterID: requesterID,
		Requester:   requester,
	})

	// Unified identity for RLS and permissions
	c.Set("identity", &entities.Identity{
		UserID:  requesterID,
		OrgID:   "SYSTEM",
		OrgPath: "/system",
		Role:    "external_service",
		IsRoot:  true,
	})
}

// handleUserIdentity handles standard user JWTs using real validation
func handleUserIdentity(c *gin.Context, token string) {
	// Define how to lookup keys from the database by KID (Version)
	lookup := func(keyID string) (*jwt.JWTKey, error) {
		k, err := repository.Repo.SigningKey.GetKeyByVersion(c.Request.Context(), keyID)
		if err != nil {
			return nil, err
		}
		if k == nil {
			return nil, nil
		}

		return &jwt.JWTKey{
			ID:        k.Version,
			Algorithm: k.Algorithm,
			KeyData:   []byte(k.KeyData),
			PublicKey: []byte(k.PublicKey),
		}, nil
	}

	// Validate the token
	claims, err := jwt.ValidateAccessToken(token, lookup)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			response.UnauthorizedResponse(c, nil, "Unauthorized: Token has expired")
			c.Abort()
			return
		}
		// Generic unauthorized for other validation failures
		response.UnauthorizedResponse(c, nil, "Unauthorized: Invalid token")
		c.Abort()
		return
	}

	// Convert claims to internal Identity entity for RLS and permissions
	id := &entities.Identity{
		UserID:  claims.AccessUserID(),
		OrgID:   claims.OrgID,
		OrgPath: claims.OrgPath,
		Role:    claims.Role,
		IsRoot:  claims.IsRoot,
		Attributes: map[string]any{
			"session_id": claims.AccessSessionID(),
		},
	}

	// Immediate Revocation Check: Verify the session still exists in the database.
	if sessionID := claims.AccessSessionID(); sessionID != "" {
		exists, err := repository.Repo.Session.ExistsByID(c.Request.Context(), sessionID)
		if err != nil || !exists {
			response.UnauthorizedResponse(c, nil, "Unauthorized: Session has been revoked")
			c.Abort()
			return
		}
	}

	c.Set("identity", id)
}
