# Permission System Guide

## Overview

The application supports three authorization strategies that can be configured per route:

1. **RBAC (Role-Based Access Control)** - Restrict access based on user roles
2. **ABAC (Attribute-Based Access Control)** - Restrict access based on user attributes
3. **COMBINED** - Require both RBAC and ABAC conditions to pass

## Protection Strategies

### 1. UNPROTECTED
No authentication or authorization required.

```go
{
    Path:        "/public",
    Method:      "GET",
    ProtectedBy: entities.UNPROTECTED,
    Controller: entities.TController{
        Handler: PublicHandler,
    },
}
```

---

### 2. JWT
Requires valid JWT authentication only (no role/attribute checks).

```go
{
    Path:        "/profile",
    Method:      "GET",
    ProtectedBy: entities.JWT,
    Controller: entities.TController{
        Handler: ProfileHandler,
    },
}
```

---

### 3. RBAC_AUTH
Requires authentication AND user role must be in allowed roles list.

```go
{
    Path:        "/admin/users",
    Method:      "GET",
    ProtectedBy: entities.RBAC_AUTH,
    Permissions: &entities.RBACConfig{
        AllowedRoles: []string{"admin", "superadmin"},
    },
    Controller: entities.TController{
        Handler: AdminUsersHandler,
    },
}
```

**Example**: Only users with role "admin" or "superadmin" can access this endpoint.

---

### 4. ABAC_AUTH
Requires authentication AND user attributes must match required attributes.

```go
{
    Path:        "/department/reports",
    Method:      "GET",
    ProtectedBy: entities.ABAC_AUTH,
    Attributes: &entities.ABACConfig{
        RequiredAttributes: map[string]any{
            "department": "finance",
            "clearance_level": "high",
        },
    },
    Controller: entities.TController{
        Handler: DepartmentReportsHandler,
    },
}
```

**Example**: Only users with `department: "finance"` AND `clearance_level: "high"` can access this endpoint.

---

### 5. COMBINED_AUTH
Requires authentication AND both RBAC and ABAC conditions must pass.

```go
{
    Path:        "/sensitive/data",
    Method:      "POST",
    ProtectedBy: entities.COMBINED_AUTH,
    Permissions: &entities.RBACConfig{
        AllowedRoles: []string{"admin", "auditor"},
    },
    Attributes: &entities.ABACConfig{
        RequiredAttributes: map[string]any{
            "security_clearance": "top_secret",
            "background_check": "passed",
        },
    },
    Controller: entities.TController{
        Handler: SensitiveDataHandler,
    },
}
```

**Example**: User must have role "admin" or "auditor" AND have `security_clearance: "top_secret"` AND `background_check: "passed"`.

---

## Real-World Examples

### Example 1: Admin-Only Endpoint (RBAC)

```go
{
    Path:        "/auth/logout-all",
    Method:      "POST",
    ProtectedBy: entities.RBAC_AUTH,
    Permissions: &entities.RBACConfig{
        AllowedRoles: []string{"root"},
    },
    Controller: entities.TController{
        Handler: LogoutAllHandler,
    },
}
```

Only root users can logout all sessions.

---

### Example 2: Department-Specific Access (ABAC)

```go
{
    Path:        "/hr/payroll",
    Method:      "GET",
    ProtectedBy: entities.ABAC_AUTH,
    Attributes: &entities.ABACConfig{
        RequiredAttributes: map[string]any{
            "department": "hr",
            "access_payroll": true,
        },
    },
    Controller: entities.TController{
        Handler: PayrollHandler,
    },
}
```

Only users in HR department with payroll access can view payroll data.

---

### Example 3: Multi-Factor Authorization (COMBINED)

```go
{
    Path:        "/financial/transactions",
    Method:      "POST",
    ProtectedBy: entities.COMBINED_AUTH,
    Permissions: &entities.RBACConfig{
        AllowedRoles: []string{"finance_manager", "cfo"},
    },
    Attributes: &entities.ABACConfig{
        RequiredAttributes: map[string]any{
            "mfa_enabled": true,
            "ip_whitelisted": true,
        },
    },
    Controller: entities.TController{
        Handler: TransactionHandler,
    },
}
```

User must be finance_manager or CFO AND have MFA enabled AND be on whitelisted IP.

---

## How It Works

### 1. Authentication (AuthMiddleware)
Applied globally to all routes. Extracts JWT and sets identity in context.

### 2. Authorization (EnforcePermissions)
Applied per-route based on `ProtectedBy` configuration:

- **UNPROTECTED**: Skip authorization
- **JWT**: Verify identity exists (authentication only)
- **RBAC_AUTH**: Check if `identity.Role` is in `AllowedRoles`
- **ABAC_AUTH**: Check if `identity.Attributes` match `RequiredAttributes`
- **COMBINED_AUTH**: Both RBAC and ABAC must pass

### 3. Error Responses

| Scenario | Status | Message |
|----------|--------|---------|
| No authentication | 401 | "Authentication required" |
| Invalid token | 401 | "Invalid authentication" |
| RBAC failed | 403 | "Insufficient permissions: role not authorized" |
| ABAC failed | 403 | "Insufficient permissions: required attributes not met" |
| COMBINED failed | 403 | "Insufficient permissions: authorization requirements not met" |

---

## Setting User Attributes

User attributes are stored in the `Identity` entity and can be set during login or token refresh:

```go
identity := &entities.Identity{
    UserID:    user.ID,
    SessionID: session.ID,
    OrgID:     orgID,
    OrgPath:   orgPath,
    Role:      role,
    IsRoot:    isRoot,
    Attributes: map[string]any{
        "department": "engineering",
        "clearance_level": "standard",
        "mfa_enabled": true,
    },
}
```

Attributes can come from:
- User profile in database
- Organization membership metadata
- Session-specific data
- External identity providers

---

## Best Practices

### 1. Use RBAC for Simple Role Checks
```go
ProtectedBy: entities.RBAC_AUTH,
Permissions: &entities.RBACConfig{
    AllowedRoles: []string{"admin", "moderator"},
}
```

### 2. Use ABAC for Context-Specific Access
```go
ProtectedBy: entities.ABAC_AUTH,
Attributes: &entities.ABACConfig{
    RequiredAttributes: map[string]any{
        "owns_resource": true,
        "resource_type": "document",
    },
}
```

### 3. Use COMBINED for High-Security Endpoints
```go
ProtectedBy: entities.COMBINED_AUTH,
Permissions: &entities.RBACConfig{
    AllowedRoles: []string{"admin"},
},
Attributes: &entities.ABACConfig{
    RequiredAttributes: map[string]any{
        "mfa_verified": true,
        "session_secure": true,
    },
}
```

### 4. Keep Attribute Names Consistent
Define attribute keys as constants to avoid typos:

```go
const (
    AttrDepartment = "department"
    AttrClearance  = "clearance_level"
    AttrMFAEnabled = "mfa_enabled"
)
```

---

## Testing Permissions

### Test RBAC
```bash
# Should succeed (admin role)
curl -X POST http://localhost:8000/ovmsa/v1.0/auth/logout-all \
  -H "Authorization: Bearer <admin_token>"

# Should fail (user role)
curl -X POST http://localhost:8000/ovmsa/v1.0/auth/logout-all \
  -H "Authorization: Bearer <user_token>"
# Expected: 403 Forbidden
```

### Test ABAC
```bash
# Should succeed (correct attributes)
curl -X GET http://localhost:8000/api/department/reports \
  -H "Authorization: Bearer <finance_user_token>"

# Should fail (wrong department)
curl -X GET http://localhost:8000/api/department/reports \
  -H "Authorization: Bearer <engineering_user_token>"
# Expected: 403 Forbidden
```

---

## Migration Guide

### Before (Hardcoded)
```go
func Handler(ctx *gin.Context, ...) {
    identity := ctx.Get("identity")
    if !identity.IsRoot {
        return errors.New("unauthorized")
    }
    // ... handler logic
}
```

### After (RBAC Configuration)
```go
// In route definition
{
    Path:        "/admin/action",
    Method:      "POST",
    ProtectedBy: entities.RBAC_AUTH,
    Permissions: &entities.RBACConfig{
        AllowedRoles: []string{"root"},
    },
    Controller: entities.TController{
        Handler: Handler,
    },
}

// Handler is now clean
func Handler(ctx *gin.Context, ...) {
    // Permission already enforced by middleware
    // ... handler logic
}
```

---

## Summary

✅ **No Hardcoding** - All permissions configured in route definitions  
✅ **Flexible** - Supports RBAC, ABAC, and COMBINED strategies  
✅ **Reusable** - Same middleware for all authorization needs  
✅ **Maintainable** - Centralized permission logic  
✅ **Testable** - Easy to test different permission scenarios
