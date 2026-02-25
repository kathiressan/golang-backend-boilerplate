package jwt

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"
	config "ovmsa-be/configs"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

var (
	testRSAPrivateKey string
	testRSAPublicKey  string
)

func TestMain(m *testing.M) {
	// Generate temporary RSA key pair for testing RS256
	priv, _ := rsa.GenerateKey(rand.Reader, 2048)

	privBytes := x509.MarshalPKCS1PrivateKey(priv)
	privBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privBytes,
	}
	testRSAPrivateKey = string(pem.EncodeToMemory(privBlock))

	pubBytes, _ := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	pubBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubBytes,
	}
	testRSAPublicKey = string(pem.EncodeToMemory(pubBlock))

	exitVal := m.Run()
	os.Exit(exitVal)
}

func getTestConfig() *config.Config {
	return &config.Config{
		AppName:                  "TestApp",
		JWTSecret:                "test-secret",
		AccessTokenExpiryMinutes: 60,
		JWTSigningMethod:         "HS256",
		PlatformName:             "TestPlatform",
	}
}

func TestGenerateAndValidateAccessTokenHS256(t *testing.T) {
	cfg := getTestConfig()
	manager := NewJWTManager(cfg)

	identity := UserIdentity{
		UserID:    "user-123",
		SessionID: "sess-999",
		OrgID:     "org-456",
		OrgPath:   "/org/456",
		Role:      "admin",
		IsRoot:    true,
		Audience:  "TestApp",
	}

	token, err := manager.GenerateAccessToken(identity)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	claims, err := manager.ValidateAccessToken(token)
	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, identity.UserID, claims.AccessUserID())
	assert.Equal(t, identity.SessionID, claims.AccessSessionID())
	assert.Equal(t, identity.OrgID, claims.OrgID)
	assert.Equal(t, identity.Role, claims.Role)

	// Verify clock leeway (should be valid even if issued "now")
	assert.WithinDuration(t, time.Now(), claims.NotBefore.Time, 61*time.Second)
}

func TestGenerateAndValidateAccessTokenRS256(t *testing.T) {
	cfg := getTestConfig()
	cfg.JWTSigningMethod = "RS256"
	cfg.JWTPrivateKey = testRSAPrivateKey
	cfg.JWTPublicKey = testRSAPublicKey
	manager := NewJWTManager(cfg)

	identity := UserIdentity{
		UserID:    "user-rs256",
		SessionID: "sess-rs256",
		OrgID:     "org-rs256",
		OrgPath:   "/rs256",
		Role:      "superuser",
		IsRoot:    true,
		Audience:  "TestApp",
	}

	token, err := manager.GenerateAccessToken(identity)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	claims, err := manager.ValidateAccessToken(token)
	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, identity.UserID, claims.AccessUserID())
	assert.Equal(t, identity.SessionID, claims.AccessSessionID())
	assert.Equal(t, identity.OrgID, claims.OrgID)
}

func TestValidateInvalidTokenRS256(t *testing.T) {
	cfg := getTestConfig()
	cfg.JWTSigningMethod = "RS256"
	cfg.JWTPrivateKey = testRSAPrivateKey
	cfg.JWTPublicKey = testRSAPublicKey
	manager := NewJWTManager(cfg)

	// Sign with a different key
	otherPriv, _ := rsa.GenerateKey(rand.Reader, 2048)
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{"sub": "123"})
	tokenString, _ := token.SignedString(otherPriv)

	claims, err := manager.ValidateAccessToken(tokenString)
	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.True(t, errors.Is(err, ErrTokenInvalid))
}

func TestGenerateRefreshToken(t *testing.T) {
	token, err := GenerateRefreshToken()
	assert.NoError(t, err)
	assert.Len(t, token, 64)
}

func TestHashToken(t *testing.T) {
	token := "sample-token"
	hash := HashToken(token)
	assert.Equal(t, hash, HashToken(token))
}

func TestGenerateAndValidateAccessTokenWithDBKey(t *testing.T) {
	cfg := getTestConfig()
	manager := NewJWTManager(cfg)
	// Simulate a database-backed key
	dbKey := &JWTKey{
		ID:        "v2-key",
		Algorithm: "HS256",
		KeyData:   []byte("db-secret-key"),
	}

	identity := UserIdentity{
		UserID:    "db-user",
		SessionID: "db-sess",
		Audience:  "TestApp",
	}

	// 1. Generate with DB key
	token, err := manager.GenerateAccessToken(identity, dbKey)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// 2. Validate with Lookup function
	lookup := func(keyID string) (*JWTKey, error) {
		if keyID == dbKey.ID {
			return dbKey, nil
		}
		return nil, nil
	}

	claims, err := manager.ValidateAccessToken(token, lookup)
	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, identity.UserID, claims.AccessUserID())
	assert.Equal(t, dbKey.ID, claims.KeyID)
}

func TestValidateAudienceMismatch(t *testing.T) {
	cfg := getTestConfig()
	cfg.EnforceAudienceValidation = true
	manager := NewJWTManager(cfg)

	// Standard audience is cfg.AppName ("TestApp" in TestMain)
	identity := UserIdentity{
		UserID:   "user-1",
		Audience: "wrong-aud",
	}

	token, _ := manager.GenerateAccessToken(identity)

	// Validate (should fail because "wrong-aud" != "TestApp")
	claims, err := manager.ValidateAccessToken(token)
	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.True(t, errors.Is(err, ErrTokenInvalid))

	// Validate with correct audience
	identity.Audience = "TestApp"
	token2, _ := manager.GenerateAccessToken(identity)
	claims2, err := manager.ValidateAccessToken(token2)
	assert.NoError(t, err)
	assert.NotNil(t, claims2)
	assert.Equal(t, "TestApp", claims2.Audience[0])
}

func TestTokenExpired(t *testing.T) {
	cfg := getTestConfig()
	cfg.AccessTokenExpiryMinutes = -1 // Expired in the past
	manager := NewJWTManager(cfg)

	identity := UserIdentity{UserID: "user-exp", Audience: "TestApp"}
	token, _ := manager.GenerateAccessToken(identity)

	claims, err := manager.ValidateAccessToken(token)
	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.True(t, errors.Is(err, ErrTokenExpired))
}

