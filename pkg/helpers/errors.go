package helpers

import (
	"errors"
	"ovmsa-be/pkg/response"

	"github.com/gin-gonic/gin"
)

// ErrorHandler defines the interface for error handling in a chain
type ErrorHandler interface {
	Handle(ctx *gin.Context, err error) bool
	SetNext(handler ErrorHandler) ErrorHandler
}

// BaseErrorHandler provides the base implementation for chaining
type BaseErrorHandler struct {
	next ErrorHandler
}

// SetNext sets the next handler in the chain
func (h *BaseErrorHandler) SetNext(handler ErrorHandler) ErrorHandler {
	h.next = handler
	return handler
}

// CallNext calls the next handler in the chain if it exists
func (h *BaseErrorHandler) CallNext(ctx *gin.Context, err error) bool {
	if h.next != nil {
		return h.next.Handle(ctx, err)
	}
	return false
}

// SpecificErrorHandler handles specific error types with custom responses
type SpecificErrorHandler struct {
	BaseErrorHandler
	targetError error
	statusCode  int
	message     string
}

// NewSpecificErrorHandler creates a new handler for a specific error type
func NewSpecificErrorHandler(targetError error, statusCode int, message string) *SpecificErrorHandler {
	return &SpecificErrorHandler{
		targetError: targetError,
		statusCode:  statusCode,
		message:     message,
	}
}

// Handle checks if the error matches and responds accordingly
func (h *SpecificErrorHandler) Handle(ctx *gin.Context, err error) bool {
	if errors.Is(err, h.targetError) {
		switch h.statusCode {
		case 401:
			response.UnauthorizedResponse(ctx, err, h.message)
		case 403:
			response.ForbiddenResponse(ctx, err, h.message)
		case 404:
			response.NotFoundResponse(ctx, err, h.message)
		case 409:
			response.ConflictResponse(ctx, err, h.message)
		default:
			response.InternalServerErrorResponse(ctx, err, h.message)
		}
		ctx.Abort()
		return true
	}
	return h.CallNext(ctx, err)
}

// ErrorHandlerChain manages a chain of error handlers
type ErrorHandlerChain struct {
	first ErrorHandler
	last  ErrorHandler
}

// NewErrorHandlerChain creates a new error handler chain
func NewErrorHandlerChain() *ErrorHandlerChain {
	return &ErrorHandlerChain{}
}

// Add adds a handler to the chain
func (c *ErrorHandlerChain) Add(handler ErrorHandler) *ErrorHandlerChain {
	if c.first == nil {
		c.first = handler
		c.last = handler
	} else {
		c.last.SetNext(handler)
		c.last = handler
	}
	return c
}

// Handle processes the error through the chain
func (c *ErrorHandlerChain) Handle(ctx *gin.Context, err error) bool {
	if c.first != nil {
		return c.first.Handle(ctx, err)
	}
	return false
}

// HandleServiceError is a convenience function that handles service errors
// Returns (handled bool, error) where handled indicates if the error was handled by the chain
func HandleServiceError(ctx *gin.Context, err error, chain *ErrorHandlerChain) (bool, error) {
	if err == nil {
		return false, nil
	}
	
	if chain != nil && chain.Handle(ctx, err) {
		return true, nil // Error was handled
	}
	
	return false, err // Unhandled error, should be returned to caller
}
