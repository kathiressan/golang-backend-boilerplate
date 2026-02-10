package jwt

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	config "ovmsa-be/configs"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// UserIdentity contains the core identity data for token generation
type UserIdentity struct {
	UserID    string
	SessionID string
	OrgID     string
	OrgPath   string
	Role      string
	IsRoot    bool
}

// TokenClaims represents the JWT claims for access tokens
type TokenClaims struct {
	OrgID   string `json:"org_id"`
	OrgPath string `json:"org_path"`
	Role    string `json:"role"`
	IsRoot  bool   `json:"is_root"`
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

// Create a new JWT access token with user and organization context
func GenerateAccessToken(identity UserIdentity) (string, error) {
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

	var method jwt.SigningMethod
	var key any
	var err error

	if cfg.JWTSigningMethod == "RS256" {
		method = jwt.SigningMethodRS256
		key, err = jwt.ParseRSAPrivateKeyFromPEM([]byte(cfg.JWTPrivateKey))
		if err != nil {
			return "", fmt.Errorf("failed to parse private key: %w", err)
		}
	} else {
		method = jwt.SigningMethodHS256
		key = []byte(cfg.JWTSecret)
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

// Validate and parse a JWT access token
func ValidateAccessToken(tokenString string) (*TokenClaims, error) {
	cfg := config.GetConfig()

	// Parse token
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (any, error) {
		if cfg.JWTSigningMethod == "RS256" {
			// Verify signing method
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return jwt.ParseRSAPublicKeyFromPEM([]byte(cfg.JWTPublicKey))
		}

		// Default to HS256
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(cfg.JWTSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	// Extract claims
	claims, ok := token.Claims.(*TokenClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
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