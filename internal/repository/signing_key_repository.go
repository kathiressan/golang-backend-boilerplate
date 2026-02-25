package repository

import (
	"context"
	"ovmsa-be/internal/entities"

	"gorm.io/gorm"
)

type SigningKeyRepository struct {
	*BaseRepository[entities.SigningKey]
}

func NewSigningKeyRepository(db *gorm.DB) *SigningKeyRepository {
	return &SigningKeyRepository{
		BaseRepository: NewBaseRepository[entities.SigningKey](db),
	}
}

// GetActiveKey fetches the currently active signing key from the database.
// If multiple keys are active (e.g. during rotation), the most recently created one is preferred.
func (r *SigningKeyRepository) GetActiveKey(ctx context.Context, tx ...*gorm.DB) (*entities.SigningKey, error) {
	var key entities.SigningKey
	err := r.GetDB(ctx, tx...).Where("is_active = ?", true).Order("created_at desc").First(&key).Error
	if err != nil {
		return nil, r.wrapError(err, "Signing key not found")
	}
	return &key, nil
}

// GetKeyByVersion fetches a specific signing key by its version ID.
func (r *SigningKeyRepository) GetKeyByVersion(ctx context.Context, version string, tx ...*gorm.DB) (*entities.SigningKey, error) {
	return r.FindOne(ctx, map[string]any{"version": version}, nil, tx...)
}
