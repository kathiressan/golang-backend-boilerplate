package jwt

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
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

	// Default Setup for HS256
	os.Setenv("APP_NAME", "TestApp")
	os.Setenv("PROJECT_ROOT", ".")
	os.Setenv("PLATFORM_NAME", "TestPlatform")
	os.Setenv("JWT_SECRET", "test-secret")
	os.Setenv("ACCESS_TOKEN_EXPIRY_MINUTES", "60")
	os.Setenv("JWT_SIGNING_METHOD", "HS256")

	exitVal := m.Run()
	os.Exit(exitVal)
}

func TestGenerateAndValidateAccessTokenHS256(t *testing.T) {
	os.Setenv("JWT_SIGNING_METHOD", "HS256")
	config.ResetConfigForTest()
	
	identity := UserIdentity{
		UserID:    "user-123",
		SessionID: "sess-999",
		OrgID:     "org-456",
		OrgPath:   "/org/456",
		Role:      "admin",
		IsRoot:    true,
	}

	token, err := GenerateAccessToken(identity)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	claims, err := ValidateAccessToken(token)
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
	os.Setenv("JWT_SIGNING_METHOD", "RS256")
	os.Setenv("JWT_PRIVATE_KEY", testRSAPrivateKey)
	os.Setenv("JWT_PUBLIC_KEY", testRSAPublicKey)
	config.ResetConfigForTest()
	
	defer func() {
		os.Setenv("JWT_SIGNING_METHOD", "HS256")
		config.ResetConfigForTest()
	}() // Reset

	identity := UserIdentity{
		UserID:    "user-rs256",
		SessionID: "sess-rs256",
		OrgID:     "org-rs256",
		OrgPath:   "/rs256",
		Role:      "superuser",
		IsRoot:    true,
	}

	token, err := GenerateAccessToken(identity)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	claims, err := ValidateAccessToken(token)
	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, identity.UserID, claims.AccessUserID())
	assert.Equal(t, identity.SessionID, claims.AccessSessionID())
	assert.Equal(t, identity.OrgID, claims.OrgID)
}

func TestValidateInvalidTokenRS256(t *testing.T) {
	os.Setenv("JWT_SIGNING_METHOD", "RS256")
	os.Setenv("JWT_PRIVATE_KEY", testRSAPrivateKey)
	os.Setenv("JWT_PUBLIC_KEY", testRSAPublicKey)
	config.ResetConfigForTest()

	defer func() {
		os.Setenv("JWT_SIGNING_METHOD", "HS256")
		config.ResetConfigForTest()
	}()

	// Sign with a different key
	otherPriv, _ := rsa.GenerateKey(rand.Reader, 2048)
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{"sub": "123"})
	tokenString, _ := token.SignedString(otherPriv)
	
	claims, err := ValidateAccessToken(tokenString)
	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Contains(t, err.Error(), "crypto/rsa: verification error")
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
	// Simulate a database-backed key
	dbKey := &JWTKey{
		ID:        "v2-key",
		Algorithm: "HS256",
		KeyData:   []byte("db-secret-key"),
	}

	identity := UserIdentity{
		UserID:    "db-user",
		SessionID: "db-sess",
	}

	// 1. Generate with DB key
	token, err := GenerateAccessToken(identity, dbKey)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// 2. Validate with Lookup function
	lookup := func(keyID string) (*JWTKey, error) {
		if keyID == dbKey.ID {
			return dbKey, nil
		}
		return nil, nil
	}

	claims, err := ValidateAccessToken(token, lookup)
	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, identity.UserID, claims.AccessUserID())
	assert.Equal(t, dbKey.ID, claims.KeyID)
}

