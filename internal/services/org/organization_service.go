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

	// Calculate and validate OrgPath
	orgID, orgPath, err := s.calculateOrgPath(ctx, name, parentID)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate org path: %w", err)
	}

	// Double check uniqueness at this level (already checked in calculateOrgPath but safe to keep or rely on it)
	exists, err := repository.Repo.Organization.Exists(ctx, map[string]any{"org_path": orgPath})
	if err != nil {
		return nil, fmt.Errorf("failed to check org path existence: %w", err)
	}
	if exists {
		return nil, ErrOrganizationAlreadyExists
	}

	org.OrgID = orgID
	org.OrgPath = orgPath

	// Create original record
	if err := repository.Repo.Organization.Create(ctx, org); err != nil {
		return nil, fmt.Errorf("failed to create organization: %w", err)
	}

	return org, nil
}

// CheckOrgNameAvailability checks if an organization name is available at a given hierarchy level
func (s *OrganizationService) CheckOrgNameAvailability(ctx context.Context, name string, parentID *string) (bool, error) {
	_, orgPath, err := s.calculateOrgPath(ctx, name, parentID)
	if err != nil {
		return false, fmt.Errorf("failed to calculate org path: %w", err)
	}

	exists, err := repository.Repo.Organization.Exists(ctx, map[string]any{"org_path": orgPath})
	if err != nil {
		return false, fmt.Errorf("failed to check org path existence: %w", err)
	}

	return !exists, nil
}

// calculateOrgPath determines the OrgID and OrgPath for a new organization.
// If parentID is provided, it inherits the OrgID and appends the name slug to the path.
func (s *OrganizationService) calculateOrgPath(ctx context.Context, name string, parentID *string) (string, string, error) {
	var orgID, orgPath string

	if parentID != nil && *parentID != "" {
		// Find parent organization to inherit its OrgID (tenant context) and set OrgPath
		parent, err := repository.Repo.Organization.FindByID(ctx, *parentID, nil)
		if err != nil {
			return "", "", fmt.Errorf("failed to find parent organization: %w", err)
		}
		if parent == nil {
			return "", "", fmt.Errorf("parent organization not found")
		}

		// Inherit OrgID and build OrgPath
		orgID = parent.OrgID
		orgPath = parent.OrgPath + utils.Slugify(name) + "/"
	} else {
		// Top-level organization
		// OrgID will be set to ID in BeforeCreate hook of Organization entity
		orgID = ""
		orgPath = "/" + utils.Slugify(name) + "/"
	}

	return orgID, orgPath, nil
}
