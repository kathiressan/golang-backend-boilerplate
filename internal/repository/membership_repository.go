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

func (r *MembershipRepository) FindByUserAndOrg(userID, orgID string) (*entities.Membership, error) {
	return r.FindOne(map[string]any{
		"user_id": userID,
		"org_id":  orgID,
	})
}

func (r *MembershipRepository) FindFirstByUser(userID string) (*entities.Membership, error) {
	memberships, err := r.FindAll(map[string]any{"user_id": userID})
	if err != nil {
		return nil, err
	}
	if len(memberships) == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &memberships[0], nil
}

func (r *MembershipRepository) FindAllByUser(userID string) ([]entities.Membership, error) {
	return r.FindAll(map[string]any{"user_id": userID})
}

func (r *MembershipRepository) FindAllByOrg(orgID string) ([]entities.Membership, error) {
	return r.FindAll(map[string]any{"org_id": orgID})
}

func (r *MembershipRepository) UpdateRole(userID, orgID, newRole string) error {
	// For composite keys, we need a custom query
	return r.GetDB().Model(&entities.Membership{}).
		Where("user_id = ? AND org_id = ?", userID, orgID).
		Update("role", newRole).Error
}

func (r *MembershipRepository) DeleteByUserAndOrg(userID, orgID string) error {
	_, err := r.DeleteWhere(map[string]any{
		"user_id": userID,
		"org_id":  orgID,
	})
	return err
}