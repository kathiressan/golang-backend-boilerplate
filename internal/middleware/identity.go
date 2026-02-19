package middleware

import (
	"context"
	"errors"
	"fmt"
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

	// identityCheckCache avoids redundant user/membership existence and permission checks (TTL 1m)
	// Key: userID:orgID:role:isRoot, Value: bool (valid)
	identityCheckCache = expirable.NewLRU[string, bool](1000, nil, 1*time.Minute)
)

// identityCtxKey is an unexported type for context keys in this package.
// Using a typed key prevents collisions with other packages using the same string.
type identityCtxKey struct{}

// PurgeCaches clears all in-memory caches used by the middleware.
// This is primarily intended for use in unit tests to ensure a clean state.
func PurgeCaches() {
	signingKeyCache.Purge()
	sessionCache.Purge()
	replayCache.Purge()
	identityCheckCache.Purge()
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
		// If it has 4 parts (new format), handle as service token.
		if len(tokenParts) == 4 {
			if _, isRequester := config.ValidRequesters[tokenParts[0]]; isRequester {
				handleExternalServiceToken(c, tokenParts)
				c.Next()
				return
			}
		}

		// Otherwise, handle as standard user JWT
		handleUserIdentity(c, token)
		
		// Add identity to request context for RLS (if it exists)
		if identity, exists := c.Get("identity"); exists {
			if id, ok := identity.(*entities.Identity); ok {
				// Update the request context with typed key to prevent collisions
				ctx := context.WithValue(c.Request.Context(), identityCtxKey{}, id)
				c.Request = c.Request.WithContext(ctx)
			}
		}
		
		c.Next()
	}
}

// handleExternalServiceToken handles tokens from known external services with format:
// Required: requesterID : unixTimestamp : nonce : signature
func handleExternalServiceToken(c *gin.Context, tokenParts []string) {
	// Enforce 4-part format only (requesterID:timestamp:nonce:signature)
	if len(tokenParts) != 4 {
		response.UnauthorizedResponse(c, nil, "Unauthorized: Invalid token format. Expected format: requesterID:timestamp:nonce:signature")
		c.Abort()
		return
	}

	requesterID := tokenParts[0]
	timestampStr := tokenParts[1]
	nonce := tokenParts[2]
	signature := tokenParts[3]

	// 1. Replay Protection: Check if this canonical message has been seen recently.
	// Key on the full message (requesterID:timestamp:nonce) rather than just the
	// signature, since a bare HMAC output is not a reliable unique fingerprint.
	replayKey := requesterID + ":" + timestampStr + ":" + nonce
	if _, seen := replayCache.Get(replayKey); seen {
		response.UnauthorizedResponse(c, nil, "Unauthorized: Replay detected")
		c.Abort()
		return
	}

	// 2. Lookup Requester
	requester, exists := config.ValidRequesters[requesterID]
	if !exists {
		response.UnauthorizedResponse(c, nil, "Unauthorized: Unknown requester ID")
		c.Abort()
		return
	}

	// 3. Verify signature
	// Proves the sender knows the secret key
	expectedSignature := cryptographyHelper.Base64HMAC(timestampStr+":"+nonce, requester.SecretKey)

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

	// 5. Add to Replay Cache (AFTER validation)
	replayCache.Add(replayKey, true)

	// Store validated context
	c.Set("tokenContext", TokenContext{
		Token:       strings.Join(tokenParts, ":"),
		RequesterID: requesterID,
		Requester:   requester,
	})

	// Unified identity for RLS and permissions.
	// We now use the role configured for the requester instead of hardcoding root access.
	role := requester.Role
	isRoot := false
	if role == "root" || role == "admin" {
		// Caution: Historically these might have been treated as root.
		// We maintain root if specifically configured or fallback to a safer default.
		if role == "root" {
			isRoot = true
		}
	}

	c.Set("identity", &entities.Identity{
		UserID:  requesterID,
		OrgID:   "SYSTEM",
		OrgPath: "/system",
		Role:    role,
		IsRoot:  isRoot,
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
			return nil, fmt.Errorf("failed to lookup signing key from database: %w", err)
		}
		if k == nil {
			return nil, nil
		}

		// Validate key is active and not expired
		if !k.IsActive {
			return nil, fmt.Errorf("signing key %s is inactive", keyID)
		}
		if k.IsExpired() {
			return nil, fmt.Errorf("signing key %s has expired", keyID)
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
		UserID:    claims.AccessUserID(),
		SessionID: claims.AccessSessionID(),
		OrgID:     claims.OrgID,
		OrgPath:   claims.OrgPath,
		Role:      claims.Role,
		IsRoot:    claims.IsRoot,
	}

	// Identity Consistency Check: Verify the identity data in the JWT is still valid.
	// This prevents "Identity Staleness" where a demoted user still has root/admin access.
	if claims.Subject != "" {
		// Cache Key: sessionId|userId|orgId|role|isRoot
		// Session ID is included so a new session after a role change is always re-validated.
		cacheKey := claims.ID + "|" + claims.Subject + "|" + claims.OrgID + "|" + claims.Role + "|" + strconv.FormatBool(claims.IsRoot)

		if _, ok := identityCheckCache.Get(cacheKey); !ok {
			// 1. Verify Global Root Status
			user, err := repository.Repo.User.FindByID(c.Request.Context(), claims.Subject, nil)
			if err != nil || user == nil {
				response.UnauthorizedResponse(c, nil, "Unauthorized: User not found")
				c.Abort()
				return
			}
			if user.IsRoot != claims.IsRoot {
				response.UnauthorizedResponse(c, nil, "Unauthorized: Identity has changed")
				c.Abort()
				return
			}

			// 2. Verify Org-Specific Role (if not root)
			if !user.IsRoot && claims.OrgID != "" {
				membership, err := repository.Repo.Membership.FindByUserAndOrg(c.Request.Context(), claims.Subject, claims.OrgID)
				if err != nil || membership == nil {
					response.UnauthorizedResponse(c, nil, "Unauthorized: No access to organization")
					c.Abort()
					return
				}
				if membership.Role != claims.Role {
					response.UnauthorizedResponse(c, nil, "Unauthorized: Permissions have changed")
					c.Abort()
					return
				}
			}

			// Store success in cache for 1m
			identityCheckCache.Add(cacheKey, true)
		}
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
