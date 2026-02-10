package helpers

import (
	"errors"
	"ovmsa-be/internal/entities"

	"github.com/gin-gonic/gin"
)

var (
	// ErrIdentityNotFound is returned when identity is not found in context
	ErrIdentityNotFound = errors.New("identity not found in context")
	// ErrInvalidIdentityType is returned when identity has invalid type
	ErrInvalidIdentityType = errors.New("invalid identity type")
)

// MustGetIdentity extracts the identity from the Gin context.
// Returns an error if identity is not found or has invalid type.
func MustGetIdentity(ctx *gin.Context) (*entities.Identity, error) {
	identity, exists := ctx.Get("identity")
	if !exists {
		return nil, ErrIdentityNotFound
	}

	id, ok := identity.(*entities.Identity)
	if !ok {
		return nil, ErrInvalidIdentityType
	}

	return id, nil
}

// ExtractPayload extracts and type-asserts the payload from the validated payload.
// Uses Go generics to provide type-safe payload extraction.
func ExtractPayload[T any](payload entities.TValidatedPayload) (*T, error) {
	req, ok := payload.(*T)
	if !ok {
		return nil, errors.New("invalid payload type")
	}
	return req, nil
}
