package middleware

import (
	"errors"
	config "ovmsa-be/configs"
	"ovmsa-be/internal/entities"
	"ovmsa-be/internal/repository"
	cryptographyHelper "ovmsa-be/pkg/cryptography"
	"ovmsa-be/pkg/jwt"
	log "ovmsa-be/pkg/logger"
	"ovmsa-be/pkg/response"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hashicorp/golang-lru/v2/expirable"
)

var (
	// signingKeyCache avoids DB hits for public keys (TTL 5m)
	signingKeyCache = expirable.NewLRU[string, *jwt.JWTKey](100, nil, 5*time.Minute)

	// sessionCache avoids redundant session existence checks (TTL 30s)
	// Key: sessionID, Value: bool (exists)
	sessionCache = expirable.NewLRU[string, bool](1000, nil, 30*time.Second)

	// replayCache tracks used signatures for external service tokens to prevent replays (TTL 5m)
	// Key: base64(signature), Value: bool
	replayCache = expirable.NewLRU[string, bool](2000, nil, 5*time.Minute)
)

// PurgeCaches clears all in-memory caches used by the middleware.
// This is primarily intended for use in unit tests to ensure a clean state.
func PurgeCaches() {
	signingKeyCache.Purge()
	sessionCache.Purge()
	replayCache.Purge()
}

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
		// If it has 4 parts (new format) or 3 parts (old format), handle as service token.
		if len(tokenParts) == 4 || len(tokenParts) == 3 {
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
// New: requesterID : unixTimestamp : nonce : signature
// Old: requesterID : unixTimestamp : signature
func handleExternalServiceToken(c *gin.Context, tokenParts []string) {
	var requesterID, timestampStr, nonce, signature string

	if len(tokenParts) == 4 {
		requesterID = tokenParts[0]
		timestampStr = tokenParts[1]
		nonce = tokenParts[2]
		signature = tokenParts[3]
	} else {
		requesterID = tokenParts[0]
		timestampStr = tokenParts[1]
		signature = tokenParts[2]
	}

	// 1. Replay Protection: Check if this signature has been seen recently
	if _, seen := replayCache.Get(signature); seen {
		response.UnauthorizedResponse(c, nil, "Unauthorized: Replay detected")
		c.Abort()
		return
	}
	replayCache.Add(signature, true)

	// 2. Lookup Requester
	requester, exists := config.ValidRequesters[requesterID]
	if !exists {
		response.UnauthorizedResponse(c, nil, "Unauthorized: Unknown requester ID")
		c.Abort()
		return
	}

	// 3. Verify signature
	// Proves the sender knows the secret key
	var expectedSignature string
	if nonce != "" {
		expectedSignature = cryptographyHelper.Base64HMAC(timestampStr+":"+nonce, requester.SecretKey)
	} else {
		expectedSignature = cryptographyHelper.Base64HMAC(timestampStr, requester.SecretKey)
	}

	if expectedSignature != signature {
		response.UnauthorizedResponse(c, nil, "Unauthorized: Invalid token signature")
		c.Abort()
		return
	}

	// 4. Time window check (Replay Protection)
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

	// Unified identity for RLS and permissions.
	// CAUTION: IsRoot: true gives full access. In a production system with multiple
	// external services, you should assign specific roles/permissions per requester.
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
		// Check Cache First
		if k, ok := signingKeyCache.Get(keyID); ok {
			return k, nil
		}

		k, err := repository.Repo.SigningKey.GetKeyByVersion(c.Request.Context(), keyID)
		if err != nil {
			return nil, err
		}
		if k == nil {
			return nil, nil
		}

		jwtKey := &jwt.JWTKey{
			ID:        k.Version,
			Algorithm: k.Algorithm,
			KeyData:   []byte(k.KeyData),
			PublicKey: []byte(k.PublicKey),
		}

		// Store in Cache
		signingKeyCache.Add(keyID, jwtKey)

		return jwtKey, nil
	}

	// Validate the token
	claims, err := jwt.ValidateAccessToken(token, lookup)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			response.UnauthorizedResponse(c, nil, "Unauthorized: Token has expired")
			c.Abort()
			return
		}
		
		// Log detailed error for internal troubleshooting
		log.Error("JWT validation failed", "error", err)
		
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
		// Check Cache First
		if _, ok := sessionCache.Get(sessionID); !ok {
			exists, err := repository.Repo.Session.ExistsByID(c.Request.Context(), sessionID)
			if err != nil || !exists {
				response.UnauthorizedResponse(c, nil, "Unauthorized: Session has been revoked")
				c.Abort()
				return
			}
			// Store success in cache for 30s
			sessionCache.Add(sessionID, true)
		}
	}

	c.Set("identity", id)
}
