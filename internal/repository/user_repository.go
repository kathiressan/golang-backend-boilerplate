package repository

import (
	"context"
	"ovmsa-be/internal/entities"

	"gorm.io/gorm"
)

type UserRepository struct {
	*BaseRepository[entities.User]
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{
		BaseRepository: NewBaseRepository[entities.User](db),
	}
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string, tx ...*gorm.DB) (*entities.User, error) {
	return r.FindOne(ctx, map[string]any{"email": email}, nil, tx...)
}

func (r *UserRepository) ExistsByEmail(ctx context.Context, email string, tx ...*gorm.DB) (bool, error) {
	return r.Exists(ctx, map[string]any{"email": email}, tx...)
}

func (r *UserRepository) UpdatePassword(ctx context.Context, userID, passwordHash string, tx ...*gorm.DB) error {
	return r.UpdateFields(ctx, userID, map[string]any{
		"password_hash": passwordHash,
	}, tx...)
}