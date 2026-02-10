package org

import "ovmsa-be/internal/entities"

var RouteMatrices = []entities.TRouteMatrix{
	{
		Path:        "",
		Method:      "POST",
		ProtectedBy: entities.RBAC_AUTH,
		Permissions: &entities.RBACConfig{
			AllowedRoles: []string{"root"},
		},
		Schema:      &CreateOrganizationRequest{},
		Controller: entities.TController{
			Handler: CreateOrganizationHandler,
		},
	},
	{
		Path:        "/check-name",
		Method:      "POST",
		ProtectedBy: entities.RBAC_AUTH,
		Permissions: &entities.RBACConfig{
			AllowedRoles: []string{"root"},
		},
		Schema:      &CheckOrgNameRequest{},
		Controller: entities.TController{
			Handler: CheckOrgNameHandler,
		},
	},
}
