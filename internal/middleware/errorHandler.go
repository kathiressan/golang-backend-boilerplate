package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorResponse represents the standard error response structure
// This struct defines the format for all error responses in the API
// Fields:
// - Success: Always false for error responses
// - Data: Typically nil for error responses
// - Error: Contains detailed error information
type ErrorResponse struct {
	Success bool        `json:"success"`
	Data    any         `json:"data"`
	Error   ErrorDetail `json:"error"`
}

// ErrorDetail represents the error details in the response
// This struct provides structured information about the error
// Fields:
// - Code: A unique error code for programmatic handling
// - Message: Human-readable error message
// - Type: The type/category of the error
// - Stack: Optional stack trace for debugging (only included in development)
type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Type    string `json:"type"`
	Stack   string `json:"stack,omitempty"`
}

// Error404n500Handler is a middleware that handles 404 and 500 errors
// This middleware:
// 1. Processes the request first to allow other middleware to run
// 2. Checks for internal server errors (500)
// 3. Checks for not found errors (404)
// 4. Returns standardized error responses
func Error404n500Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Process the request first
		// This allows other middleware to run and potentially set errors
		c.Next()

		// Check if there are any errors in the context
		if len(c.Errors) > 0 {
			lastError := c.Errors.Last()
			if lastError != nil {
				// Handle 500 Internal Server Error
				// This occurs when an unhandled error occurs during request processing
				errorResponse := ErrorResponse{
					Success: false,
					Data:    nil,
					Error: ErrorDetail{
						Code:    "INTERNAL_SERVER_ERROR",
						Message: lastError.Error(),
						Type:    "InternalServerError",
					},
				}
				c.JSON(http.StatusInternalServerError, errorResponse)
				return
			}
		}

		// Handle 404 Not Found errors
		// This occurs when the requested endpoint doesn't exist
		if c.Writer.Status() == http.StatusNotFound {
			errorResponse := ErrorResponse{
				Success: false,
				Data:    nil,
				Error: ErrorDetail{
					Code:    "NOT_FOUND",
					Message: "The endpoint " + c.Request.URL.Path + " not found.",
					Type:    "NotFound",
				},
			}
			c.JSON(http.StatusNotFound, errorResponse)
			return
		}
	}
}
