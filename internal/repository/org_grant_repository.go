package repository

import (
	"ovmsa-be/internal/entities"

	"gorm.io/gorm"
)

type OrgGrantRepository struct {
	*BaseRepository[entities.OrgGrant]
}

func NewOrgGrantRepository(db *gorm.DB) *OrgGrantRepository {
	return &OrgGrantRepository{
		BaseRepository: NewBaseRepository[entities.OrgGrant](db),
	}
}

func (r *OrgGrantRepository) FindAllByUserID(userID string, tx ...*gorm.DB) ([]entities.OrgGrant, error) {
	return r.FindAll(map[string]any{"user_id": userID}, tx...)
}

func (r *OrgGrantRepository) FindAllByOrgID(orgID string, tx ...*gorm.DB) ([]entities.OrgGrant, error) {
	return r.FindAll(map[string]any{"org_id": orgID}, tx...)
}
