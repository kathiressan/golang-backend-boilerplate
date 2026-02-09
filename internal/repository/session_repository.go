package repository

import (
	"ovmsa-be/internal/entities"
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

func (r *SessionRepository) FindByRefreshToken(hashedToken string) (*entities.Session, error) {
	return r.FindOne(map[string]any{"refresh_token": hashedToken})
}

func (r *SessionRepository) FindByUserID(userID string) ([]entities.Session, error) {
	return r.FindAll(map[string]any{"user_id": userID})
}

func (r *SessionRepository) UpdateExpiry(sessionID string, expiresAt time.Time) error {
	return r.UpdateFields(sessionID, map[string]any{"expires_at": expiresAt})
}

func (r *SessionRepository) DeleteByRefreshToken(hashedToken, userID string) (int64, error) {
	return r.DeleteWhere(map[string]any{
		"refresh_token": hashedToken,
		"user_id":       userID,
	})
}

func (r *SessionRepository) DeleteAllByUserID(userID string) error {
	_, err := r.DeleteWhere(map[string]any{"user_id": userID})
	return err
}

func (r *SessionRepository) DeleteExpired() error {
	return r.GetDB().Where("expires_at < ?", time.Now()).Delete(&entities.Session{}).Error
}

func (r *SessionRepository) CountActiveByUserID(userID string) (int64, error) {
	var count int64
	err := r.GetDB().Model(&entities.Session{}).
		Where("user_id = ? AND expires_at > ?", userID, time.Now()).
		Count(&count).Error
	return count, err
}