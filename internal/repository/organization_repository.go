package repository

import (
	"context"
	"ovmsa-be/internal/entities"

	"gorm.io/gorm"
)

type OrganizationRepository struct {
	*BaseRepository[entities.Organization]
}

func NewOrganizationRepository(db *gorm.DB) *OrganizationRepository {
	return &OrganizationRepository{
		BaseRepository: NewBaseRepository[entities.Organization](db),
	}
}

func (r *OrganizationRepository) FindByParentID(ctx context.Context, parentID string, tx ...*gorm.DB) ([]entities.Organization, error) {
	return r.FindAll(ctx, map[string]any{"parent_id": parentID}, nil, tx...)
}

func (r *OrganizationRepository) FindChildren(ctx context.Context, parentPath string, tx ...*gorm.DB) ([]entities.Organization, error) {
	var orgs []entities.Organization
	err := r.GetDB(ctx, tx...).Where("org_path LIKE ?", parentPath+"%").Find(&orgs).Error
	return orgs, r.wrapError(err, "Failed to fetch child organizations")
}
