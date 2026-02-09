package repository

import (
	"ovmsa-be/internal/entities"

	"gorm.io/gorm"
)

type MembershipRepository struct {
	*BaseRepository[entities.Membership]
}

func NewMembershipRepository(db *gorm.DB) *MembershipRepository {
	return &MembershipRepository{
		BaseRepository: NewBaseRepository[entities.Membership](db),
	}
}

func (r *MembershipRepository) FindByUserAndOrg(userID, orgID string, tx ...*gorm.DB) (*entities.Membership, error) {
	return r.FindOne(map[string]any{
		"user_id": userID,
		"org_id":  orgID,
	}, tx...)
}

func (r *MembershipRepository) FindFirstByUser(userID string, tx ...*gorm.DB) (*entities.Membership, error) {
	memberships, err := r.FindAll(map[string]any{"user_id": userID}, tx...)
	if err != nil {
		return nil, err
	}
	if len(memberships) == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &memberships[0], nil
}

func (r *MembershipRepository) FindAllByUser(userID string, tx ...*gorm.DB) ([]entities.Membership, error) {
	return r.FindAll(map[string]any{"user_id": userID}, tx...)
}

func (r *MembershipRepository) FindAllByOrg(orgID string, tx ...*gorm.DB) ([]entities.Membership, error) {
	return r.FindAll(map[string]any{"org_id": orgID}, tx...)
}

func (r *MembershipRepository) UpdateRole(userID, orgID, newRole string, tx ...*gorm.DB) error {
	// For composite keys, we need a custom query
	return r.GetDB(tx...).Model(&entities.Membership{}).
		Where("user_id = ? AND org_id = ?", userID, orgID).
		Update("role", newRole).Error
}

func (r *MembershipRepository) DeleteByUserAndOrg(userID, orgID string, tx ...*gorm.DB) error {
	_, err := r.DeleteWhere(map[string]any{
		"user_id": userID,
		"org_id":  orgID,
	}, tx...)
	return err
}