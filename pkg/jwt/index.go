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

// TokenClaims represents the JWT claims for access tokens
type TokenClaims struct {
	UserID string `json:"user_id"`
	OrgID string `json:"org_id"`
	OrgPath string `json:"org_path"`
	Role string `json:"role"`
	IsRoot bool `json:"is_root"`
	jwt.RegisteredClaims
}

// Create a new JWT access token with user and organization context
func GenerateAccessToken(userID, orgID, orgPath, role string, isRoot bool) (string, error) {
	cfg := config.GetConfig()

	// Create claims with expiry
	claims := TokenClaims{
		UserID:  userID,
		OrgID:   orgID,
		OrgPath: orgPath,
		Role:    role,
		IsRoot:  isRoot,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * time.Duration(cfg.JWTExpiryHours))),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    cfg.AppName,
		},
	}

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token with secret
	tokenString, err := token.SignedString([]byte(cfg.JWTSecret))
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
		// Verify signing method
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