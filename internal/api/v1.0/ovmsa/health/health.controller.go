package health

import (
	"ovmsa-be/internal/entities"

	"github.com/gin-gonic/gin"
)

func HealthCheckHandler(ctx *gin.Context, payload entities.TValidatedPayload, jwtData *entities.TJwtData, params entities.TParams) (any, error, error) {
	return map[string]string{
		"status":  "ok",
		"message": "health check passed",
	}, nil, nil
}
