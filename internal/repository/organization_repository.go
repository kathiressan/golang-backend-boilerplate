package repository

import (
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

func (r *OrganizationRepository) FindByParentID(parentID string, tx ...*gorm.DB) ([]entities.Organization, error) {
	return r.FindAll(map[string]any{"parent_id": parentID}, tx...)
}

func (r *OrganizationRepository) FindChildren(parentPath string, tx ...*gorm.DB) ([]entities.Organization, error) {
	var orgs []entities.Organization
	err := r.GetDB(tx...).Where("org_path LIKE ?", parentPath+"%").Find(&orgs).Error
	return orgs, err
}
