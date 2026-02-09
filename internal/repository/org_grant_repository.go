package repository

import (
	"context"
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

func (r *OrgGrantRepository) FindAllByUserID(ctx context.Context, userID string, tx ...*gorm.DB) ([]entities.OrgGrant, error) {
	return r.FindAll(ctx, map[string]any{"user_id": userID}, nil, tx...)
}

func (r *OrgGrantRepository) FindAllByOrgID(ctx context.Context, orgID string, tx ...*gorm.DB) ([]entities.OrgGrant, error) {
	return r.FindAll(ctx, map[string]any{"org_id": orgID}, nil, tx...)
}
