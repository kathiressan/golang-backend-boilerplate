package health

import (
	"ovmsa-be/internal/entities"
	"ovmsa-be/pkg/helpers"
)

var RouteMatrices = []entities.TRouteMatrix{
	helpers.GET("/").
		Unprotected().
		WithHandler(HealthCheckHandler).
		Build(),
}
