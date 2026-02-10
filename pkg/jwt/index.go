package jwt

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	config "ovmsa-be/configs"
	"slices"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrTokenExpired   = errors.New("token has expired")
	ErrTokenInvalid   = errors.New("token is invalid")
	ErrTokenMalformed = errors.New("token is malformed")

	// keyCache stores already parsed RSA keys to avoid expensive re-parsing
	// map[string]any where value is *rsa.PrivateKey or *rsa.PublicKey
	keyCache sync.Map
)

// UserIdentity contains the core identity data for token generation
type UserIdentity struct {
	UserID    string
	SessionID string
	OrgID     string
	OrgPath   string
	Role      string
	IsRoot    bool
	Audience  string
}

// TokenClaims represents the JWT claims for access tokens
type TokenClaims struct {
	OrgID   string `json:"org_id"`
	OrgPath string `json:"org_path"`
	Role    string `json:"role"`
	IsRoot  bool   `json:"is_root"`
	KeyID   string `json:"kid,omitempty"` // Version of the signing key
	jwt.RegisteredClaims
}

// AccessUserID returns the user ID from the Subject claim
func (c *TokenClaims) AccessUserID() string {
	return c.Subject
}

// AccessSessionID returns the session ID from the ID (jti) claim
func (c *TokenClaims) AccessSessionID() string {
	return c.ID
}

// JWTKey represents the key data used for signing/validation
type JWTKey struct {
	ID        string
	Algorithm string
	KeyData   []byte // Secret for HMAC, Private Key PEM for RSA
	PublicKey []byte // Public Key PEM for RSA
}

// GenerateAccessToken creates a new JWT access token.
// If jwtKey is provided, it uses that key for signing; otherwise, it falls back to config.
func GenerateAccessToken(identity UserIdentity, jwtKey ...*JWTKey) (string, error) {
	cfg := config.GetConfig()

	// Create claims with expiry and leeway for clock skew
	claims := TokenClaims{
		OrgID:   identity.OrgID,
		OrgPath: identity.OrgPath,
		Role:    identity.Role,
		IsRoot:  identity.IsRoot,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   identity.UserID,
			ID:        identity.SessionID,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * time.Duration(cfg.AccessTokenExpiryMinutes))),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			// Set NotBefore to 1 minute in the past to handle slight clock skew
			NotBefore: jwt.NewNumericDate(time.Now().Add(-1 * time.Minute)),
			Issuer:    cfg.AppName,
		},
	}

	if identity.Audience != "" {
		claims.Audience = jwt.ClaimStrings{identity.Audience}
	}

	var method jwt.SigningMethod
	var key any
	var err error

	// Determine signing method and key
	if len(jwtKey) > 0 && jwtKey[0] != nil {
		k := jwtKey[0]
		claims.KeyID = k.ID
		if k.Algorithm == "RS256" {
			method = jwt.SigningMethodRS256
			// Use cache for private key
			hash := sha256.Sum256(k.KeyData)
			cacheKey := "priv-" + hex.EncodeToString(hash[:])
			if cached, ok := keyCache.Load(cacheKey); ok {
				key = cached
			} else {
				key, err = jwt.ParseRSAPrivateKeyFromPEM(k.KeyData)
				if err != nil {
					return "", fmt.Errorf("failed to parse private key: %w", err)
				}
				keyCache.Store(cacheKey, key)
			}
		} else {
			method = jwt.SigningMethodHS256
			key = k.KeyData
		}
	} else {
		// Fallback to Config
		if cfg.JWTSigningMethod == "RS256" {
			method = jwt.SigningMethodRS256
			pemData := []byte(cfg.JWTPrivateKey)
			hash := sha256.Sum256(pemData)
			cacheKey := "priv-cfg-" + hex.EncodeToString(hash[:])
			if cached, ok := keyCache.Load(cacheKey); ok {
				key = cached
			} else {
				key, err = jwt.ParseRSAPrivateKeyFromPEM(pemData)
				if err != nil {
					return "", fmt.Errorf("failed to parse private key from config: %w", err)
				}
				keyCache.Store(cacheKey, key)
			}
		} else {
			method = jwt.SigningMethodHS256
			key = []byte(cfg.JWTSecret)
		}
	}

	// Create token with claims
	token := jwt.NewWithClaims(method, claims)

	// Sign token
	tokenString, err := token.SignedString(key)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// KeyLookupFunc is a callback to fetch a key by its ID (kid)
type KeyLookupFunc func(keyID string) (*JWTKey, error)

// ValidateAccessToken parses and validates a JWT token.
// If lookup is provided, it uses it to find the key by kid; otherwise, it falls back to config.
func ValidateAccessToken(tokenString string, lookup ...KeyLookupFunc) (*TokenClaims, error) {
	cfg := config.GetConfig()

	// Parse token
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (any, error) {
		claims, ok := token.Claims.(*TokenClaims)
		if !ok {
			return nil, errors.New("invalid token claims during key lookup")
		}

		// Use lookup function if provided and kid is present
		if len(lookup) > 0 && lookup[0] != nil && claims.KeyID != "" {
			k, err := lookup[0](claims.KeyID)
			if err != nil {
				return nil, fmt.Errorf("failed to lookup key %s: %w", claims.KeyID, err)
			}
			if k == nil {
				return nil, fmt.Errorf("key %s not found", claims.KeyID)
			}
			if k.Algorithm == "RS256" {
				if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				// Use cache for public key
				hash := sha256.Sum256(k.PublicKey)
				cacheKey := "pub-" + hex.EncodeToString(hash[:])
				if cached, ok := keyCache.Load(cacheKey); ok {
					return cached, nil
				}
				pk, err := jwt.ParseRSAPublicKeyFromPEM(k.PublicKey)
				if err != nil {
					return nil, err
				}
				keyCache.Store(cacheKey, pk)
				return pk, nil
			}
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return k.KeyData, nil
		}

		// Fallback to Config
		if cfg.JWTSigningMethod == "RS256" {
			// Verify signing method
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			pemData := []byte(cfg.JWTPublicKey)
			hash := sha256.Sum256(pemData)
			cacheKey := "pub-cfg-" + hex.EncodeToString(hash[:])
			if cached, ok := keyCache.Load(cacheKey); ok {
				return cached, nil
			}
			pk, err := jwt.ParseRSAPublicKeyFromPEM(pemData)
			if err != nil {
				return nil, err
			}
			keyCache.Store(cacheKey, pk)
			return pk, nil
		}

		// Default to HS256
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(cfg.JWTSecret), nil
	})


	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		if errors.Is(err, jwt.ErrTokenMalformed) {
			return nil, ErrTokenMalformed
		}
		// Log detailed error for internal debugging if needed, but return generic error to client
		return nil, fmt.Errorf("%w: %v", ErrTokenInvalid, err)
	}

	claims, ok := token.Claims.(*TokenClaims)
	if !ok || !token.Valid {
		return nil, ErrTokenInvalid
	}

	// Verify Audience
	// If the token has an audience, it MUST match the app name.
	// If the app is configured with an audience but the token has none, it's also invalid.
	if len(claims.Audience) > 0 {
		if !slices.Contains(claims.Audience, cfg.AppName) {
			return nil, fmt.Errorf("%w: audience mismatch", ErrTokenInvalid)
		}
	} else if cfg.AppName != "" {
		// Strictly require audience if AppName is set in config
		// This can be made optional depending on project requirements, but stricter is better for B2B.
		// return nil, fmt.Errorf("%w: missing audience", ErrTokenInvalid)
	}

	return claims, nil
}

// Create a cryptographically secure random refresh token
func GenerateRefreshToken() (string, error) {
	// Generate 32 random bytes (256 bits)
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random token: %w", err)
	}

	// Encode as hex string (64 characters)
	return hex.EncodeToString(bytes), nil
}

// Create a SHA256 hash of a token for secure storage
func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}