package org

import (
	"errors"
	"ovmsa-be/internal/entities"
	"ovmsa-be/internal/services/org"
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
	req, ok := payload.(*CreateOrganizationRequest)
	if !ok {
		return nil, errors.New("invalid payload type"), nil
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
