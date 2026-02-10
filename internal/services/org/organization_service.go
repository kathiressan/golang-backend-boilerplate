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

// CreateOrganization creates a new organization
func (s *OrganizationService) CreateOrganization(ctx context.Context, name string, parentID *string, tier string) (*entities.Organization, error) {
	org := &entities.Organization{
		Name:     name,
		ParentID: parentID,
		Tier:     tier,
	}

	// Handle hierarchical organization creation
	if parentID != nil && *parentID != "" {
		// Find parent organization to inherit its OrgID (tenant context) and set OrgPath
		parent, err := repository.Repo.Organization.FindByID(ctx, *parentID, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to find parent organization: %w", err)
		}
		if parent == nil {
			return nil, fmt.Errorf("parent organization not found")
		}

		// Inherit OrgID and build OrgPath
		org.OrgID = parent.OrgID
		org.OrgPath = parent.OrgPath + utils.Slugify(name) + "/"
	} else {
		// Top-level organization
		// OrgID will be set to ID in BeforeCreate hook of Organization entity
		org.OrgPath = "/" + utils.Slugify(name) + "/"
	}

	// Create original record
	if err := repository.Repo.Organization.Create(ctx, org); err != nil {
		return nil, fmt.Errorf("failed to create organization: %w", err)
	}

	return org, nil
}
