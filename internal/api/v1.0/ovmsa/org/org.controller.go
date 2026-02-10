package org

import (
	"errors"
	"ovmsa-be/internal/entities"
	"ovmsa-be/internal/services/org"
	"ovmsa-be/pkg/helpers"
	"ovmsa-be/pkg/response"

	"github.com/gin-gonic/gin"
)

var orgService *org.OrganizationService

// SetOrganizationService sets the organization service for the controllers
func SetOrganizationService(svc *org.OrganizationService) {
	orgService = svc
}

// CreateOrganizationHandler handles POST /org
func CreateOrganizationHandler(ctx *gin.Context, payload entities.TValidatedPayload, jwtData *entities.TJwtData, params entities.TParams) (any, error, error) {
	req, err := helpers.ExtractPayload[CreateOrganizationRequest](payload)
	if err != nil {
		return nil, err, nil
	}

	// Set default tier if not provided
	tier := req.Tier
	if tier == "" {
		tier = "free"
	}

	// Create organization
	result, err := orgService.CreateOrganization(ctx.Request.Context(), req.Name, req.ParentID, tier)
	if err != nil {
		if errors.Is(err, org.ErrOrganizationAlreadyExists) {
			response.ConflictResponse(ctx, err, err.Error())
			ctx.Abort()
			return nil, nil, nil
		}
		return nil, err, nil
	}

	return result, nil, nil
}

// CheckOrgNameHandler handles GET /org/check-name
func CheckOrgNameHandler(ctx *gin.Context, payload entities.TValidatedPayload, jwtData *entities.TJwtData, params entities.TParams) (any, error, error) {
	req, err := helpers.ExtractPayload[CheckOrgNameRequest](payload)
	if err != nil {
		return nil, err, nil
	}

	// Check name availability
	available, err := orgService.CheckOrgNameAvailability(ctx.Request.Context(), req.Name, req.ParentID)
	if err != nil {
		return nil, err, nil
	}

	return map[string]bool{
		"available": available,
	}, nil, nil
}
