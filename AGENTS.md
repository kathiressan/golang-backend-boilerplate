# AGENTS.md - Quick Reference for AI Agents

A concise guide for AI agents to work with this Go backend boilerplate.

## Essential Commands

```bash
go run cmd/app/main.go          # Start server
go test ./...                   # Run all tests
go test -v ./path/to/package    # Verbose tests
go fmt ./...                    # Format code
go mod tidy                     # Tidy dependencies
```

## Architecture

```
cmd/app/main.go       # Entry point
internal/
  api/                # HTTP handlers, routes (Gin)
  services/           # Business logic
  repository/         # Data access (GORM)
  entities/           # Domain models
  middleware/         # Auth, logging, CORS, etc.
pkg/                  # Shared utilities (logger, errors, validator, etc.)
configs/              # Configuration, static data
```

**Request Flow**: Middleware → Controller → Service → Repository → Database

## Key Patterns

### Multi-Tenancy (RLS)
- All tenanted entities embed `entities.BaseEntity` (provides `OrgID`, `OrgPath`)
- PostgreSQL Row-Level Security enforces isolation at DB level
- Identity travels via `context.Context`
- RLS Plugin auto-injects session variables before queries

### Authentication
- **JWT**: `Authorization: Bearer <token>` - validates signature, checks session
- **Service Token**: `Authorization: Bearer requesterID:timestamp:nonce:signature`
- Identity stored in context: `entities.GetIdentity(ctx)`

### Authorization Strategies
- `UNPROTECTED` - No auth required
- `JWT` - Authenticated only
- `RBAC_AUTH` - Role-based (`AllowedRoles: []string{"admin"}`)
- `ABAC_AUTH` - Attribute-based (`RequiredAttributes: map[string]any{}`)
- `COMBINED_AUTH` - Both RBAC + ABAC

### Error Handling
```go
import "ovmsa-be/pkg/errors"

errors.NotFound(err, "User not found")
errors.BadRequest(err, "Invalid input")
errors.Unauthorized(err, "Invalid token")
errors.InternalServerError(err, "Database error")
```

### Logging
```go
import "ovmsa-be/pkg/logger"

logger.Info("Message", "key", value, "key2", value2)
logger.Error("Failed", "error", err)
```

## Adding a New Feature

1. **Entity** (`internal/entities/xxx.go`):
```go
type Project struct {
    entities.BaseEntity  // Embed for multi-tenancy
    Title string `gorm:"not null"`
}
```

2. **Repository** (`internal/repository/xxx_repository.go`):
```go
type ProjectRepository struct {
    *BaseRepository[entities.Project]
}
```
Register in `internal/repository/repository.go`.

3. **Service** (`internal/services/xxx/xxx_service.go`):
Business logic here. Register in `internal/services/index.go`.

4. **API** (`internal/api/v1.0/ovmsa/xxx/`):
- `xxx.validate.go` - Request/response structs with `validate` tags
- `xxx.controller.go` - Handler functions
- `xxx.route.go` - Route definitions using `RouteBuilder`

5. **Wire it up**:
- Add route matrix to `internal/api/routes.go`
- Add service to `controller_factory.go`

## Route Definition

```go
helpers.POST("/endpoint").
    ProtectedByJWT().
    WithSchema(&CreateRequest{}).
    WithResponseSchema(&Response{}).
    WithHandler(HandlerFunc).
    Build()
```

## Validation Tags

```go
type Request struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=8"`
    Role     string `json:"role" validate:"omitempty,oneof=admin user"`
}
```

## Common Imports

```go
import (
    "ovmsa-be/internal/entities"
    "ovmsa-be/internal/repository"
    "ovmsa-be/pkg/errors"
    "ovmsa-be/pkg/logger"
    "ovmsa-be/pkg/response"
    "ovmsa-be/pkg/utils"
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
)
```

## Root User (Testing)

- Email: `root@system.local`
- Password: `RootPass123!`
- Bypasses all RLS policies

## API Documentation

Swagger UI: `http://localhost:8080/swagger/`

## Environment Variables

Key variables (see `configs/config.go`):
- `ENVIRONMENT` - development/production
- `PORT` - Server port (default: 8000)
- `DATABASE_URL` or `DB_HOST/DB_PORT/DB_USER/DB_PASSWORD/DB_NAME`
- `JWT_SECRET` - Signing key for HS256

## Troubleshooting

- **Empty results/401**: Check RLS - verify `OrgID`/`OrgPath` match record
- **"Replay detected"**: Service token reused within 5-minute window
- **Missing columns**: Add `tx.AutoMigrate(&Entity{})` to migration

## Testing

```go
db, _ := database.InitializeTestDB()  // In-memory SQLite
repository.Initialize(db)
middleware.PurgeCaches()  // Clear caches between tests
```

## File Naming Conventions

- `xxx.controller.go` - HTTP handlers
- `xxx.service.go` - Business logic
- `xxx.repository.go` - Data access
- `xxx.validate.go` - Request/response DTOs
- `xxx.route.go` - Route definitions
- `xxx_test.go` - Tests

## Don't

- Don't use `fmt.Printf` - use `logger`
- Don't skip error wrapping - always add context
- Don't store plain text passwords - use `pkg/password.HashPassword()`
- Don't bypass RLS - always embed `BaseEntity` for tenanted entities
- Don't use UUID - use ULID via `utils.NewULID()`
