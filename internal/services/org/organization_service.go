package org

import (
	"context"
	"fmt"
	"ovmsa-be/internal/entities"
	"ovmsa-be/internal/repository"
	"ovmsa-be/pkg/utils"

	"gorm.io/gorm"
)

// OrganizationService handles organization-related business logic
type OrganizationService struct {
	db *gorm.DB
}

// NewOrganizationService creates a new organization service
func NewOrganizationService(db *gorm.DB) *OrganizationService {
	return &OrganizationService{
		db: db,
	}
}

var (
	// ErrOrganizationAlreadyExists is returned when an organization with the same path already exists
	ErrOrganizationAlreadyExists = fmt.Errorf("organization with this name already exists at this level")
)

// CreateOrganization creates a new organization
func (s *OrganizationService) CreateOrganization(ctx context.Context, name string, parentID *string, tier string) (*entities.Organization, error) {
	org := &entities.Organization{
		Name:     name,
		ParentID: parentID,
		Tier:     tier,
	}

	// Handle hierarchical organization creation
	var baseOrgPath string
	if parentID != nil && *parentID != "" {
		// Find parent organization to inherit its OrgID (tenant context) and set OrgPath
		parent, err := repository.Repo.Organization.FindByID(ctx, *parentID, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to find parent organization: %w", err)
		}
		if parent == nil {
			return nil, fmt.Errorf("parent organization not found")
		}

		// Inherit OrgID and set base OrgPath
		org.OrgID = parent.OrgID
		baseOrgPath = parent.OrgPath + utils.Slugify(name)
	} else {
		// Top-level organization
		// OrgID will be set to ID in BeforeCreate hook of Organization entity
		baseOrgPath = "/" + utils.Slugify(name)
	}

	// Ensure OrgPath is unique at this level
	finalOrgPath := baseOrgPath + "/"
	exists, err := repository.Repo.Organization.Exists(ctx, map[string]any{"org_path": finalOrgPath})
	if err != nil {
		return nil, fmt.Errorf("failed to check org path existence: %w", err)
	}
	if exists {
		return nil, ErrOrganizationAlreadyExists
	}
	org.OrgPath = finalOrgPath

	// Create original record
	if err := repository.Repo.Organization.Create(ctx, org); err != nil {
		return nil, fmt.Errorf("failed to create organization: %w", err)
	}

	return org, nil
}
