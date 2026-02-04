// Package response provides standardized response handling for the API
// This package defines consistent response structures and helper functions
// for handling both success and error responses across the application
package response

import (
	appErrors "ovmsa-be/pkg/errors"
	"ovmsa-be/pkg/logger"

	"github.com/gin-gonic/gin"
)

// SuccessResponse represents the standard success response structure
// This struct is used to format all successful API responses consistently
// Fields:
// - Success: Always true for success responses
// - Data: The response payload (can be any type)
// - Message: A human-readable success message
type SuccessResponse struct {
	Success bool   `json:"success"`
	Data    any    `json:"data"`
	Message string `json:"message"`
}

// ErrorResponse represents the standard error response structure
// This struct is used to format all error responses consistently
// Fields:
// - Success: Always false for error responses
// - Data: Usually nil for error responses
// - Error: Contains detailed error information using AppError
type ErrorResponse struct {
	Success bool               `json:"success"`
	Data    any                `json:"data"`
	Error   appErrors.AppError `json:"error"`
}

// Success returns a standard success response
// This helper function creates a consistent success response format
// Parameters:
// - msg: A human-readable success message
// - data: The response payload (can be any type)
// Returns:
// - gin.H: A map containing the formatted success response
func Success(msg string, data any) gin.H {
	return gin.H{
		"success": true,
		"message": msg,
		"data":    data,
	}
}

// Error handles error responses with the appropriate status code
// This function processes errors and returns them in a standardized format
// It handles both application errors and unexpected errors
// Parameters:
// - c: The Gin context
// - err: The error to process
func Error(c *gin.Context, err error) {
	requestID, _ := c.Get("X-Request-ID")

	// If it's an application error, use it directly
	if appErr, ok := err.(*appErrors.AppError); ok {
		logger.Error("Request error",
			"requestID", requestID,
			"statusCode", appErr.StatusCode,
			"errorCode", appErr.Code,
			"errorType", appErr.Type,
			"message", appErr.Message,
		)

		c.JSON(appErr.StatusCode, ErrorResponse{
			Success: false,
			Data:    nil,
			Error:   *appErr,
		})
		return
	}

	// Otherwise, wrap it as an internal server error
	appErr := appErrors.WrapError(err)

	logger.Error("Internal server error",
		"requestID", requestID,
		"error", err.Error(),
	)

	c.JSON(appErr.StatusCode, ErrorResponse{
		Success: false,
		Data:    nil,
		Error:   *appErr,
	})
}

// BadRequestResponse returns a 400 Bad Request response
// This helper function creates a standardized bad request error response
// Parameters:
// - c: The Gin context
// - err: The underlying error
// - msg: A human-readable error message
func BadRequestResponse(c *gin.Context, err error, msg string) {
	appErr := appErrors.BadRequest(err, msg)
	Error(c, appErr)
}

// UnauthorizedResponse returns a 401 Unauthorized response
// This helper function creates a standardized unauthorized error response
// Parameters:
// - c: The Gin context
// - err: The underlying error
// - msg: A human-readable error message
func UnauthorizedResponse(c *gin.Context, err error, msg string) {
	appErr := appErrors.Unauthorized(err, msg)
	Error(c, appErr)
}

// ForbiddenResponse returns a 403 Forbidden response
// This helper function creates a standardized forbidden error response
// Parameters:
// - c: The Gin context
// - err: The underlying error
// - msg: A human-readable error message
func ForbiddenResponse(c *gin.Context, err error, msg string) {
	appErr := appErrors.Forbidden(err, msg)
	Error(c, appErr)
}

// NotFoundResponse returns a 404 Not Found response
// This helper function creates a standardized not found error response
// Parameters:
// - c: The Gin context
// - err: The underlying error
// - msg: A human-readable error message
func NotFoundResponse(c *gin.Context, err error, msg string) {
	appErr := appErrors.NotFound(err, msg)
	Error(c, appErr)
}

// ValidationErrorResponse returns a 422 Validation Error response
// This helper function creates a standardized validation error response
// Parameters:
// - c: The Gin context
// - err: The underlying error
// - msg: A human-readable error message
func ValidationErrorResponse(c *gin.Context, err error, msg string) {
	appErr := appErrors.ValidationError(err, msg)
	Error(c, appErr)
}

// ValidationErrorWithDetailsResponse returns a 422 Validation Error response with detailed validation errors
// This helper function creates a standardized validation error response with additional details
// Parameters:
// - c: The Gin context
// - err: The underlying error
// - msg: A human-readable error message
// - details: Additional validation error details
func ValidationErrorWithDetailsResponse(c *gin.Context, err error, msg string, details any) {
	appErr := appErrors.ValidationErrorWithDetails(err, msg, details)
	Error(c, appErr)
}

// InternalServerErrorResponse returns a 500 Internal Server Error response
// This helper function creates a standardized internal server error response
// Parameters:
// - c: The Gin context
// - err: The underlying error
// - msg: A human-readable error message
func InternalServerErrorResponse(c *gin.Context, err error, msg string) {
	appErr := appErrors.InternalServerError(err, msg)
	Error(c, appErr)
}
