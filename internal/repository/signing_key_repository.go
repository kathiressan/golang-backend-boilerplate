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
func (r *SigningKeyRepository) GetActiveKey(ctx context.Context, tx ...*gorm.DB) (*entities.SigningKey, error) {
	return r.FindOne(ctx, map[string]any{"is_active": true}, nil, tx...)
}

// GetKeyByVersion fetches a specific signing key by its version ID.
func (r *SigningKeyRepository) GetKeyByVersion(ctx context.Context, version string, tx ...*gorm.DB) (*entities.SigningKey, error) {
	return r.FindOne(ctx, map[string]any{"version": version}, nil, tx...)
}
