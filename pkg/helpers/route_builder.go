package helpers

import (
	"ovmsa-be/internal/entities"

	"github.com/gin-gonic/gin"
)

// RouteBuilder provides a fluent API for building route configurations
type RouteBuilder struct {
	route entities.TRouteMatrix
}

// NewRoute creates a new route builder with path and method
func NewRoute(path, method string) *RouteBuilder {
	return &RouteBuilder{
		route: entities.TRouteMatrix{
			Path:   path,
			Method: method,
			// Default to unprotected
			ProtectedBy: entities.UNPROTECTED,
		},
	}
}

// Unprotected marks the route as unprotected (public)
func (b *RouteBuilder) Unprotected() *RouteBuilder {
	b.route.ProtectedBy = entities.UNPROTECTED
	return b
}

// ProtectedByJWT marks the route as requiring JWT authentication
func (b *RouteBuilder) ProtectedByJWT() *RouteBuilder {
	b.route.ProtectedBy = entities.JWT
	return b
}

// ProtectedByRBAC marks the route as requiring RBAC authorization
func (b *RouteBuilder) ProtectedByRBAC(roles ...string) *RouteBuilder {
	b.route.ProtectedBy = entities.RBAC_AUTH
	b.route.Permissions = &entities.RBACConfig{
		AllowedRoles: roles,
	}
	return b
}

// ProtectedByABAC marks the route as requiring ABAC authorization
func (b *RouteBuilder) ProtectedByABAC(attributes map[string]any) *RouteBuilder {
	b.route.ProtectedBy = entities.ABAC_AUTH
	b.route.Attributes = &entities.ABACConfig{
		RequiredAttributes: attributes,
	}
	return b
}

// ProtectedByCombined marks the route as requiring both RBAC and ABAC
func (b *RouteBuilder) ProtectedByCombined(roles []string, attributes map[string]any) *RouteBuilder {
	b.route.ProtectedBy = entities.COMBINED_AUTH
	b.route.Permissions = &entities.RBACConfig{
		AllowedRoles: roles,
	}
	b.route.Attributes = &entities.ABACConfig{
		RequiredAttributes: attributes,
	}
	return b
}

// WithSchema sets the request schema for validation
func (b *RouteBuilder) WithSchema(schema any) *RouteBuilder {
	b.route.Schema = schema
	return b
}

// WithHandler sets the handler function
func (b *RouteBuilder) WithHandler(handler func(*gin.Context, entities.TValidatedPayload, *entities.TJwtData, entities.TParams) (any, error, error)) *RouteBuilder {
	b.route.Controller = entities.TController{
		Handler: handler,
	}
	return b
}

// WithParams sets route parameters
func (b *RouteBuilder) WithParams(params entities.TParams) *RouteBuilder {
	b.route.Params = params
	return b
}

// WithSuccessCode sets the HTTP status code returned on success.
// Defaults to 200 when not called. Use 201 for POST endpoints that create resources.
func (b *RouteBuilder) WithSuccessCode(code int) *RouteBuilder {
	b.route.SuccessCode = code
	return b
}

// Build returns the configured route matrix
func (b *RouteBuilder) Build() entities.TRouteMatrix {
	return b.route
}

// Convenience methods for common HTTP methods

// GET creates a GET route builder
func GET(path string) *RouteBuilder {
	return NewRoute(path, "GET")
}

// POST creates a POST route builder
func POST(path string) *RouteBuilder {
	return NewRoute(path, "POST")
}

// PUT creates a PUT route builder
func PUT(path string) *RouteBuilder {
	return NewRoute(path, "PUT")
}

// DELETE creates a DELETE route builder
func DELETE(path string) *RouteBuilder {
	return NewRoute(path, "DELETE")
}

// PATCH creates a PATCH route builder
func PATCH(path string) *RouteBuilder {
	return NewRoute(path, "PATCH")
}
