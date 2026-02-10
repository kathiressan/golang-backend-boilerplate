package org

import (
	"ovmsa-be/internal/entities"
	"ovmsa-be/pkg/helpers"
)

var RouteMatrices = []entities.TRouteMatrix{
	helpers.POST("").
		ProtectedByRBAC("root").
		WithSchema(&CreateOrganizationRequest{}).
		WithHandler(CreateOrganizationHandler).
		Build(),

	helpers.POST("/check-name").
		ProtectedByRBAC("root").
		WithSchema(&CheckOrgNameRequest{}).
		WithHandler(CheckOrgNameHandler).
		Build(),
}
