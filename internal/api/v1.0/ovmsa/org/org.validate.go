package org

import "ovmsa-be/internal/entities"

// CreateOrganizationRequest represents the payload for creating a new organization
type CreateOrganizationRequest struct {
	entities.TValidatedPayload `json:"-"`
	Name                       string  `json:"name" validate:"required,min=2,max=100"`
	ParentID                   *string `json:"parent_id" validate:"omitempty,len=26"`
	Tier                       string  `json:"tier" validate:"omitempty,oneof=free basic enterprise"`
}

// CheckOrgNameRequest represents the query parameters for checking organization name availability
type CheckOrgNameRequest struct {
	entities.TValidatedPayload `json:"-"`
	Name                       string  `json:"name" validate:"required,min=2,max=100"`
	ParentID                   *string `json:"parent_id" validate:"omitempty,len=26"`
}
