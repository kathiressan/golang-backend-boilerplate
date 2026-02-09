package repository

import (
	"context"
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

func (r *MembershipRepository) FindByUserAndOrg(ctx context.Context, userID, orgID string, tx ...*gorm.DB) (*entities.Membership, error) {
	return r.FindOne(ctx, map[string]any{
		"user_id": userID,
		"org_id":  orgID,
	}, nil, tx...)
}

func (r *MembershipRepository) FindFirstByUser(ctx context.Context, userID string, tx ...*gorm.DB) (*entities.Membership, error) {
	memberships, err := r.FindAll(ctx, map[string]any{"user_id": userID}, nil, tx...)
	if err != nil {
		return nil, err
	}
	if len(memberships) == 0 {
		return nil, r.wrapError(gorm.ErrRecordNotFound, "Membership not found")
	}
	return &memberships[0], nil
}

func (r *MembershipRepository) FindAllByUser(ctx context.Context, userID string, tx ...*gorm.DB) ([]entities.Membership, error) {
	return r.FindAll(ctx, map[string]any{"user_id": userID}, nil, tx...)
}

func (r *MembershipRepository) FindAllByOrg(ctx context.Context, orgID string, tx ...*gorm.DB) ([]entities.Membership, error) {
	return r.FindAll(ctx, map[string]any{"org_id": orgID}, nil, tx...)
}

func (r *MembershipRepository) UpdateRole(ctx context.Context, userID, orgID, newRole string, tx ...*gorm.DB) error {
	// For composite keys, we need a custom query
	err := r.GetDB(ctx, tx...).Model(&entities.Membership{}).
		Where("user_id = ? AND org_id = ?", userID, orgID).
		Update("role", newRole).Error
	return r.wrapError(err, "Failed to update membership role")
}

func (r *MembershipRepository) DeleteByUserAndOrg(ctx context.Context, userID, orgID string, tx ...*gorm.DB) error {
	_, err := r.DeleteWhere(ctx, map[string]any{
		"user_id": userID,
		"org_id":  orgID,
	}, tx...)
	return err
}