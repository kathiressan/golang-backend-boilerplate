package repository

import (
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

func (r *UserRepository) FindByEmail(email string, tx ...*gorm.DB) (*entities.User, error) {
	return r.FindOne(map[string]any{"email": email}, tx...)
}

func (r *UserRepository) ExistsByEmail(email string, tx ...*gorm.DB) (bool, error) {
	return r.Exists(map[string]any{"email": email}, tx...)
}

func (r *UserRepository) UpdatePassword(userID, passwordHash, passwordSalt string, tx ...*gorm.DB) error {
	return r.UpdateFields(userID, map[string]any{
		"password_hash": passwordHash,
		"password_salt": passwordSalt,
	}, tx...)
}