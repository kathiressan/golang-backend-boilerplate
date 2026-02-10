package api

import (
	"ovmsa-be/internal/api/v1.0/ovmsa/auth"
	"ovmsa-be/internal/api/v1.0/ovmsa/health"
	"ovmsa-be/internal/api/v1.0/ovmsa/org"
	"ovmsa-be/internal/entities"
)

// Registry defines the hierarchy of routes: Platform -> Version -> Groups
type Registry struct {
	Platform string
	Version  string
	Groups   []entities.TGroup
}

// RouteRegistry is the central store for all versioned platform routes
var RouteRegistry = []Registry{
	{
		Platform: "ovmsa",
		Version:  "v1.0",
		Groups: []entities.TGroup{
			{
				Group:         "/health",
				RouteMatrices: health.RouteMatrices,
			},
			{
				Group:         "/auth",
				RouteMatrices: auth.RouteMatrices,
			},
			{
				Group:         "/org",
				RouteMatrices: org.RouteMatrices,
			},
		},
	},
}
