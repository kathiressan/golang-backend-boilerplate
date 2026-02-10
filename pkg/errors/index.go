// Package errors provides a standardized error handling system for the application
// It creates structured errors with consistent formatting, status codes, and additional context
// This approach helps ensure that errors throughout the application are:
// - Consistent in format and behavior
// - Easy to identify and classify
// - Properly formatted for API responses
// - Traceable for debugging
package errors

import (
	"errors"
	"fmt"
	"net/http"
	config "ovmsa-be/configs"
	"runtime"
	"strings"
)

var cfg = config.GetConfig()

// AppError is the custom error type used throughout the application
// This struct contains all the information needed for both
// internal error handling and external error responses
type AppError struct {
	// HTTP Status code (e.g., 400, 404, 500)
	// Not included in JSON responses (indicated by json:"-")
	StatusCode int `json:"-"`

	// Public facing error code
	// A consistent string identifier for the error type
	// Examples: "NOT_FOUND", "VALIDATION_ERROR"
	Code string `json:"code"`

	// User facing message
	// A human-readable description of what went wrong
	Message string `json:"message"`

	// Error type classification
	// Used to categorize errors (e.g., "ValidationError", "AuthenticationError")
	Type string `json:"type"`

	// Stack trace (only included in non-production)
	// Shows where the error occurred in the code
	Stack string `json:"stack,omitempty"`

	// Validation details for validation errors
	// Used to provide specific information about validation failures
	// For example, which fields failed validation and why
	ValidationDetails any `json:"details,omitempty"`

	// Original error
	// The underlying error that was wrapped (not included in JSON)
	Err error `json:"-"`
}

// Error satisfies the error interface
// This method allows AppError to be used anywhere a standard Go error is expected
// It returns the error message as the string representation of the error
func (e *AppError) Error() string {
	return e.Message
}

// Unwrap allows using errors.Is and errors.As with wrapped errors
// This enables the standard Go error handling functions to work with AppError
// It returns the original error that was wrapped
func (e *AppError) Unwrap() error {
	return e.Err
}

// callers is an internal helper function that generates a stack trace
// It captures the call stack at the point where the error was created
// and formats it as a string, only including application files
func callers() string {
	// Allocate space for 32 stack frames
	var pc [32]uintptr
	// Skip 3 frames to exclude this function and New/NewWithDetails
	n := runtime.Callers(3, pc[:])

	// Build a formatted stack trace string
	var builder strings.Builder
	for _, pc := range pc[:n] {
		fn := runtime.FuncForPC(pc)
		if fn == nil {
			continue
		}
		file, line := fn.FileLine(pc)
		// Only include application files (exclude standard library and third-party packages)
		if strings.Contains(file, cfg.ProjectRoot) {
			fmt.Fprintf(&builder, "%s:%d\n", file, line)
		}
	}
	return builder.String()
}

// New creates a new application error
// This is the core function for creating AppErrors
// Parameters:
// - err: The original error (can be nil)
// - statusCode: HTTP status code to use in responses
// - code: A machine-readable error code string
// - message: A human-readable error message
// - errorType: Category of the error
func New(err error, statusCode int, code string, message string, errorType string) *AppError {
	stack := ""
	// Only include stack traces in non-production environments
	// This keeps production payloads smaller and avoids exposing implementation details
	if !config.IsProduction() {
		stack = callers()
	}

	return &AppError{
		StatusCode: statusCode,
		Code:       code,
		Message:    message,
		Type:       errorType,
		Stack:      stack,
		Err:        err,
	}
}

// NewWithDetails creates a new application error with additional details
// This is useful for validation errors where you want to include more information
// about what specifically failed validation
func NewWithDetails(err error, statusCode int, code string, message string, errorType string, details any) *AppError {
	appErr := New(err, statusCode, code, message, errorType)
	appErr.ValidationDetails = details
	return appErr
}

// ========== Specific error constructors ==========
// These functions provide convenient ways to create common error types
// They use standard HTTP status codes and consistent error codes

// BadRequest creates a 400 Bad Request error
// Use for invalid input from clients (malformed requests)
// Example: BadRequest(err, "Invalid query parameter 'limit'")
func BadRequest(err error, message string) *AppError {
	return New(err, http.StatusBadRequest, "BAD_REQUEST", message, "BadRequestError")
}

// Unauthorized creates a 401 Unauthorized error
// Use when authentication is required but not provided or invalid
// Example: Unauthorized(err, "Invalid API key")
func Unauthorized(err error, message string) *AppError {
	return New(err, http.StatusUnauthorized, "UNAUTHORIZED", message, "UnauthorizedError")
}

// Forbidden creates a 403 Forbidden error
// Use when a user is authenticated but doesn't have permission
// Example: Forbidden(err, "You don't have permission to access this resource")
func Forbidden(err error, message string) *AppError {
	return New(err, http.StatusForbidden, "FORBIDDEN", message, "ForbiddenError")
}

// NotFound creates a 404 Not Found error
// Use when a requested resource doesn't exist
// Example: NotFound(err, "User with ID 123 not found")
func NotFound(err error, message string) *AppError {
	return New(err, http.StatusNotFound, "NOT_FOUND", message, "NotFoundError")
}

// InternalServerError creates a 500 Internal Server Error
// Use for unexpected server-side errors that aren't the client's fault
// Example: InternalServerError(err, "Database connection failed")
func InternalServerError(err error, message string) *AppError {
	if message == "" {
		message = "An unexpected error occurred"
	}
	return New(err, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", message, "InternalServerError")
}

// ValidationError creates a 422 Unprocessable Entity error
// Use when request data fails validation rules
// Example: ValidationError(err, "Invalid user data")
func ValidationError(err error, message string) *AppError {
	return New(err, http.StatusUnprocessableEntity, "VALIDATION_ERROR", message, "ValidationError")
}

// ValidationErrorWithDetails creates a 422 Unprocessable Entity error with validation details
// Use when you need to provide specific information about validation failures
// Example: ValidationErrorWithDetails(err, "Invalid user data", {"email": "Invalid email format"})
func ValidationErrorWithDetails(err error, message string, details any) *AppError {
	appErr := ValidationError(err, message)
	appErr.ValidationDetails = details
	return appErr
}

// TooManyRequests creates a 429 Too Many Requests error
// Use when a client has exceeded rate limits
// Example: TooManyRequests(nil, "Rate limit exceeded. Please try again later.")
func TooManyRequests(err error, message string) *AppError {
	return New(err, http.StatusTooManyRequests, "TOO_MANY_REQUESTS", message, "RateLimitError")
}

// ========== Error handling utilities ==========

// Is checks if an error is of a specific AppError type
// This function allows checking the type of an error without type assertions
// Example: if errors.Is(err, "ValidationError") { ... }
func Is(err error, errorType string) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Type == errorType
	}
	return false
}

// AsAppError tries to convert a standard error to an AppError
// This function is useful when you need to access AppError-specific fields
// Example:
//
//	if appErr, ok := errors.AsAppError(err); ok {
//	    statusCode = appErr.StatusCode
//	}
func AsAppError(err error) (*AppError, bool) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}

// WrapError wraps a standard error as an internal server error
// This is a convenience function for quickly converting standard errors
// Example: return errors.WrapError(err)
func WrapError(err error) *AppError {
	return InternalServerError(err, err.Error())
}

// ========== How to use this error system ==========
//
// 1. Creating basic errors:
//    return errors.NotFound(nil, "User not found")
//
// 2. Wrapping existing errors:
//    if err := db.GetUser(id); err != nil {
//        return errors.InternalServerError(err, "Failed to get user")
//    }
//
// 3. Validation errors with details:
//    validationErrors := map[string]string{
//        "email": "Invalid email format",
//        "password": "Password must be at least 8 characters"
//    }
//    return errors.ValidationErrorWithDetails(nil, "Invalid input", validationErrors)
//
// 4. Checking error types:
//    if errors.Is(err, "ValidationError") {
//        // Handle validation errors differently
//    }
//
// 5. Accessing error details:
//    if appErr, ok := errors.AsAppError(err); ok {
//        // Use appErr.StatusCode, appErr.Code, etc.
//    }
