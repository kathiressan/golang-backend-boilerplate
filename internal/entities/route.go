package entities

import "github.com/gin-gonic/gin"

type ProtectionStrategy string

const (
	JWT           ProtectionStrategy = "JSON_WEB_TOKEN"
	OTP           ProtectionStrategy = "ONE_TIME_TOKEN"
	STATIC_TOKEN  ProtectionStrategy = "STATIC_TOKEN"
	UNPROTECTED   ProtectionStrategy = "UNPROTECTED"
	RBAC_AUTH     ProtectionStrategy = "ROLE_BASED_ACCESS_CONTROL"
	ABAC_AUTH     ProtectionStrategy = "ATTRIBUTE_BASED_ACCESS_CONTROL"
	COMBINED_AUTH ProtectionStrategy = "COMBINED_RBAC_ABAC_CONTROL"
)

type TValidatedPayload = any

type TJwtData struct {
	UserID string
	Role   string
}

type TParams map[string]any

type RBACConfig struct {
	AllowedRoles []string
}

type ABACConfig struct {
	RequiredAttributes map[string]any
}

type TController struct {
	Handler func(*gin.Context, TValidatedPayload, *TJwtData, TParams) (any, error, error)
	Params  TParams
}

type TRouteMatrix struct {
	Path        string `validate:"required"`
	Method      string `validate:"required"`
	ProtectedBy ProtectionStrategy
	Permissions *RBACConfig
	Attributes  *ABACConfig
	Schema      any
	Controller  TController `validate:"required"`
	Params      TParams
}

type TGroup struct {
	Group         string
	RouteMatrices []TRouteMatrix
}
