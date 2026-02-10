package middleware

import (
	"ovmsa-be/internal/entities"
	"ovmsa-be/pkg/response"
	"slices"

	"github.com/gin-gonic/gin"
)

// EnforceRBAC checks if the user's role is in the allowed roles list
func EnforceRBAC(c *gin.Context, identity *entities.Identity, config *entities.RBACConfig) bool {
	if config == nil || len(config.AllowedRoles) == 0 {
		return true // No restrictions if config is empty
	}

	// Check if user's role matches any allowed role
	return slices.Contains(config.AllowedRoles, identity.Role)
}

// EnforceABAC checks if the user's attributes match the required attributes
func EnforceABAC(c *gin.Context, identity *entities.Identity, config *entities.ABACConfig) bool {
	if config == nil || len(config.RequiredAttributes) == 0 {
		return true // No restrictions if config is empty
	}

	// Check each required attribute
	for key, requiredValue := range config.RequiredAttributes {
		actualValue, exists := identity.Attributes[key]
		if !exists || actualValue != requiredValue {
			return false
		}
	}

	return true
}

// EnforcePermissions is a generic middleware that enforces RBAC, ABAC, or COMBINED authorization
func EnforcePermissions(protectionStrategy entities.ProtectionStrategy, rbacConfig *entities.RBACConfig, abacConfig *entities.ABACConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip if no protection required
		if protectionStrategy == entities.UNPROTECTED {
			c.Next()
			return
		}

		// Get identity from context (should be set by AuthMiddleware)
		identityVal, exists := c.Get("identity")
		if !exists {
			response.UnauthorizedResponse(c, nil, "Authentication required")
			c.Abort()
			return
		}

		identity, ok := identityVal.(*entities.Identity)
		if !ok {
			response.UnauthorizedResponse(c, nil, "Invalid authentication")
			c.Abort()
			return
		}

		// Handle different protection strategies
		switch protectionStrategy {
		case entities.JWT:
			// JWT only requires authentication (already validated above)
			c.Next()
			return

		case entities.RBAC_AUTH:
			// Enforce RBAC only
			if !EnforceRBAC(c, identity, rbacConfig) {
				response.ForbiddenResponse(c, nil, "Insufficient permissions: role not authorized")
				c.Abort()
				return
			}

		case entities.ABAC_AUTH:
			// Enforce ABAC only
			if !EnforceABAC(c, identity, abacConfig) {
				response.ForbiddenResponse(c, nil, "Insufficient permissions: required attributes not met")
				c.Abort()
				return
			}

		case entities.COMBINED_AUTH:
			// Enforce both RBAC and ABAC (both must pass)
			rbacPassed := EnforceRBAC(c, identity, rbacConfig)
			abacPassed := EnforceABAC(c, identity, abacConfig)

			if !rbacPassed || !abacPassed {
				response.ForbiddenResponse(c, nil, "Insufficient permissions: authorization requirements not met")
				c.Abort()
				return
			}

		default:
			// Unknown protection strategy
			response.InternalServerErrorResponse(c, nil, "Invalid protection strategy configured")
			c.Abort()
			return
		}

		c.Next()
	}
}
