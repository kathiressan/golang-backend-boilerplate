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

func (r *SessionRepository) FindByRefreshToken(hashedToken string, tx ...*gorm.DB) (*entities.Session, error) {
	return r.FindOne(map[string]any{"refresh_token": hashedToken}, tx...)
}

func (r *SessionRepository) FindByUserID(userID string, tx ...*gorm.DB) ([]entities.Session, error) {
	return r.FindAll(map[string]any{"user_id": userID}, tx...)
}

func (r *SessionRepository) UpdateExpiry(sessionID string, expiresAt time.Time, tx ...*gorm.DB) error {
	return r.UpdateFields(sessionID, map[string]any{"expires_at": expiresAt}, tx...)
}

func (r *SessionRepository) DeleteByRefreshToken(hashedToken, userID string, tx ...*gorm.DB) (int64, error) {
	return r.DeleteWhere(map[string]any{
		"refresh_token": hashedToken,
		"user_id":       userID,
	}, tx...)
}

func (r *SessionRepository) DeleteAllByUserID(userID string, tx ...*gorm.DB) error {
	_, err := r.DeleteWhere(map[string]any{"user_id": userID}, tx...)
	return err
}

func (r *SessionRepository) DeleteExpired(tx ...*gorm.DB) error {
	return r.GetDB(tx...).Where("expires_at < ?", time.Now()).Delete(&entities.Session{}).Error
}

func (r *SessionRepository) CountActiveByUserID(userID string, tx ...*gorm.DB) (int64, error) {
	var count int64
	err := r.GetDB(tx...).Model(&entities.Session{}).
		Where("user_id = ? AND expires_at > ?", userID, time.Now()).
		Count(&count).Error
	return count, err
}