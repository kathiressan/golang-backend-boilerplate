package health

import "ovmsa-be/internal/entities"

var RouteMatrices = []entities.TRouteMatrix{
	{
		Path:        "/",
		Method:      "GET",
		ProtectedBy: entities.UNPROTECTED,
		Controller: entities.TController{
			Handler: HealthCheckHandler,
		},
	},
}
