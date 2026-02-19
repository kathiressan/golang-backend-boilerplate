package health

import (
	"ovmsa-be/internal/entities"
	"ovmsa-be/pkg/helpers"

	"github.com/gin-gonic/gin"
)

var healthErrorChain *helpers.ErrorHandlerChain

// SetHealthErrorChain sets the error handler chain for health controllers
func SetHealthErrorChain(chain *helpers.ErrorHandlerChain) {
	healthErrorChain = chain
}

func HealthCheckHandler(ctx *gin.Context, payload entities.TValidatedPayload, identity *entities.Identity, params entities.TParams) (any, error, error) {
	return map[string]string{
		"status":  "ok",
		"message": "health check passed",
	}, nil, nil
}
