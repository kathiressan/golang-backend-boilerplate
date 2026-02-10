package repository

import (
	"context"
	"errors"
	"ovmsa-be/internal/entities"
	appErrors "ovmsa-be/pkg/errors"
	"time"

	"gorm.io/gorm"
)

type SessionRepository struct {
	*BaseRepository[entities.Session]
}

func NewSessionRepository(db *gorm.DB) *SessionRepository {
	return &SessionRepository{
		BaseRepository: NewBaseRepository[entities.Session](db),
	}
}

func (r *SessionRepository) FindByRefreshToken(ctx context.Context, hashedToken string, tx ...*gorm.DB) (*entities.Session, error) {
	return r.FindOne(ctx, map[string]any{"refresh_token": hashedToken}, nil, tx...)
}

func (r *SessionRepository) FindByUserID(ctx context.Context, userID string, tx ...*gorm.DB) ([]entities.Session, error) {
	return r.FindAll(ctx, map[string]any{"user_id": userID}, nil, tx...)
}

func (r *SessionRepository) UpdateExpiry(ctx context.Context, sessionID string, expiresAt time.Time, tx ...*gorm.DB) error {
	return r.UpdateFields(ctx, sessionID, map[string]any{"expires_at": expiresAt}, tx...)
}

func (r *SessionRepository) DeleteByRefreshToken(ctx context.Context, hashedToken, userID string, tx ...*gorm.DB) (int64, error) {
	return r.DeleteWhere(ctx, map[string]any{
		"refresh_token": hashedToken,
		"user_id":       userID,
	}, tx...)
}

func (r *SessionRepository) DeleteAllByUserID(ctx context.Context, userID string, tx ...*gorm.DB) error {
	_, err := r.DeleteWhere(ctx, map[string]any{"user_id": userID}, tx...)
	return err
}

func (r *SessionRepository) DeleteExpired(ctx context.Context, tx ...*gorm.DB) error {
	err := r.GetDB(ctx, tx...).Where("expires_at < ?", time.Now()).Delete(&entities.Session{}).Error
	return r.wrapError(err, "Failed to delete expired sessions")
}

func (r *SessionRepository) CountActiveByUserID(ctx context.Context, userID string, tx ...*gorm.DB) (int64, error) {
	var count int64
	err := r.GetDB(ctx, tx...).Model(&entities.Session{}).
		Where("user_id = ? AND expires_at > ?", userID, time.Now()).
		Count(&count).Error
	return count, r.wrapError(err, "Failed to count active sessions")
}

// RotateRefreshToken atomically replaces an old hashed refresh token with a new one.
// It returns an error if the old token doesn't match, which helps detect token reuse/theft.
func (r *SessionRepository) RotateRefreshToken(ctx context.Context, sessionID, oldHashedToken, newHashedToken string, newExpiry time.Time, tx ...*gorm.DB) error {
	db := r.GetDB(ctx, tx...)
	
	result := db.Model(&entities.Session{}).
		Where("id = ? AND refresh_token = ?", sessionID, oldHashedToken).
		Updates(map[string]any{
			"refresh_token": newHashedToken,
			"expires_at":    newExpiry,
		})

	if result.Error != nil {
		return r.wrapError(result.Error, "Failed to rotate refresh token")
	}

	if result.RowsAffected == 0 {
		return appErrors.Unauthorized(errors.New("refresh token rotation failed: token mismatch or session not found"), "Invalid or already used refresh token")
	}

	return nil
}