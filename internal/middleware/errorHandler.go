package middleware

import (
	"errors"
	"net/http"
	"ovmsa-be/pkg/response"

	"github.com/gin-gonic/gin"
)

// Error404n500Handler is a middleware that handles 404 and 500 errors
func Error404n500Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Process the request first
		c.Next()

		// Check if there are any errors in the context
		if len(c.Errors) > 0 {
			lastError := c.Errors.Last()
			if lastError != nil {
				// Handle 500 Internal Server Error
				response.InternalServerErrorResponse(c, lastError.Err, "Internal Server Error")
				return
			}
		}

		// Handle 404 Not Found errors
		if c.Writer.Status() == http.StatusNotFound {
			response.NotFoundResponse(c, errors.New("route not found"), "The requested endpoint was not found")
			return
		}
	}
}
