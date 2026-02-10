package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	config "ovmsa-be/configs"
	"ovmsa-be/internal/entities"
	"ovmsa-be/internal/repository"
	"ovmsa-be/pkg/database"
	"ovmsa-be/pkg/jwt"
	log "ovmsa-be/pkg/logger"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestAuthMiddleware_Revocation(t *testing.T) {
	// Setup Gin in test mode
	gin.SetMode(gin.TestMode)

	// Setup mock environment for config
	os.Setenv("APP_NAME", "TestApp")
	os.Setenv("PROJECT_ROOT", ".")
	os.Setenv("PLATFORM_NAME", "TestPlatform")
	os.Setenv("JWT_SECRET", "test-secret")
	os.Setenv("JWT_SIGNING_METHOD", "HS256")
	config.ResetConfigForTest()
	log.Initialize(config.EnvDevelopment)

	// Initialize a real SQLite in-memory DB for repository testing
	db, err := database.InitializeTestDB()
	if err != nil {
		t.Fatalf("Failed to initialize test DB: %v", err)
	}
	repository.Initialize(db)

	// Auto-migrate the necessary tables
	db.AutoMigrate(&entities.Session{}, &entities.User{}, &entities.SigningKey{})

	// 1. Create a valid session in the DB
	sessionID := uuid.New().String()
	user := entities.User{Name: "Test User", Email: "test@example.com"}
	db.Create(&user)

	session := entities.Session{
		UserID:     user.ID,
		ExpiresAt:  time.Now().Add(time.Hour),
	}
	session.ID = sessionID
	db.Create(&session)

	// 2. Generate a JWT linked to this session
	identity := jwt.UserIdentity{
		UserID:    user.ID,
		SessionID: sessionID,
	}
	token, _ := jwt.GenerateAccessToken(identity)

	// 3. Setup router with middleware
	router := gin.New()
	router.Use(AuthMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// 4. Test Case: Valid token, valid session
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// 5. Revoke the session (delete from DB)
	db.Delete(&session)

	// 6. Test Case: Valid token, REVOKED session
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/test", nil)
	req2.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusUnauthorized, w2.Code)
	assert.Contains(t, w2.Body.String(), "Session has been revoked")
}
