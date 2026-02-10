package repository

import (
	"context"
	"ovmsa-be/internal/entities"
	"ovmsa-be/pkg/database"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSessionRepository_RotateRefreshToken(t *testing.T) {
	db, err := database.InitializeTestDB()
	if err != nil {
		t.Fatalf("Failed to initialize test DB: %v", err)
	}
	Initialize(db)
	db.AutoMigrate(&entities.Session{})

	repo := Repo.Session
	ctx := context.Background()

	// 1. Create a session with an "old" token
	sessionID := "sess-123"
	oldToken := "old-hashed-token"
	session := entities.Session{
		UserID:       "user-1",
		RefreshToken: oldToken,
		ExpiresAt:    time.Now().Add(time.Hour),
	}
	session.ID = sessionID
	db.Create(&session)

	// 2. Rotate it successfully
	newToken := "new-hashed-token"
	newExpiry := time.Now().Add(2 * time.Hour)
	err = repo.RotateRefreshToken(ctx, sessionID, oldToken, newToken, newExpiry)
	assert.NoError(t, err)

	// Verify DB state
	var updated entities.Session
	db.First(&updated, "id = ?", sessionID)
	assert.Equal(t, newToken, updated.RefreshToken)
	assert.WithinDuration(t, newExpiry, updated.ExpiresAt, time.Second)

	// 3. Try to rotate again using the SAME old token (REUSE DETECTION)
	err = repo.RotateRefreshToken(ctx, sessionID, oldToken, "another-token", time.Now())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid or already used refresh token")

	// 4. Try to rotate with wrong session ID
	err = repo.RotateRefreshToken(ctx, "wrong-id", newToken, "another-token", time.Now())
	assert.Error(t, err)
}
