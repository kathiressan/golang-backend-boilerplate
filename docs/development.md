# Development Guide

This guide provides practical instructions for extending the codebase and following the established patterns.

## 🛠 Adding a New Feature

To add a new entity (e.g., "Project") to the system, follow these steps:

### 1. Create the Entity
Define the GORM model in `internal/entities/project.go`.
Always embed `BaseEntity` if the model requires organization isolation.

```go
type Project struct {
    entities.BaseEntity
    Title       string `gorm:"not null"`
    Description string
}
```

### 2. Create the Repository
Add a specialized repository in `internal/repository/project_repository.go`.

```go
type ProjectRepository struct {
    *BaseRepository[entities.Project]
}

func (r *ProjectRepository) GetActiveProjects(ctx context.Context) ([]entities.Project, error) {
    // Custom logic here using r.FindAll or r.db
}
```

Register it in `internal/repository/repository.go`.

### 3. Create the Service
Implement business logic in `internal/services/project/project_service.go`.

```go
type ProjectService struct {
    db *gorm.DB
}

func (s *ProjectService) CreateProject(ctx context.Context, title string) (*entities.Project, error) {
    // Business logic, validation, etc.
}
```

Register it in `internal/services/init.go`.

### 4. Create the Controller, Route & Validation
- Define request/response structs in `internal/api/v1.0/ovmsa/project/project.validate.go`.
- Implement API handlers in `internal/api/v1.0/ovmsa/project/project.controller.go`.
- Define routes using `RouteBuilder` in `internal/api/v1.0/ovmsa/project/project.route.go`.

### 5. Wiring it up
1.  **Register Routes**: Add your `project.RouteMatrices` to `internal/api/routes.go`.
2.  **Dependency Injection**: Add the service and a `PrepareProjectChain` method to `internal/api/controller_factory.go`.
3.  **Initialization**: Update `InitializeControllersWithFactory` in `controller_factory.go` to inject the service and error chain for your module.
4.  **Graceful Shutdown**: If your service requires cleanup, add it to the `stopSignal` channel in `cmd/app/main.go`.

---

## 📏 Coding Standards

### Error Handling
Always wrap errors from lower layers with context. Use standard application errors from `pkg/errors` for consistent API responses.

```go
if err != nil {
    return nil, fmt.Errorf("failed to fetch user: %w", err)
}
```

### Logging
Use the `logger` package (`pkg/logger`). Avoid `fmt.Printf` or `log.Print`.

```go
logger.Info("User created", "userID", user.ID, "orgID", user.OrgID)
```

### Validation
Use the `validate` tags in your request structs. The `pkg/validator` helper will automatically handle these in the API layer.

```go
type CreateUserRequest struct {
    Email string `json:"email" validate:"required,email"`
    Name  string `json:"name" validate:"required"`
}
```

---

## 🧪 Testing

### Running Tests
Use the standard Go test command:

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...
```

### Writing Tests
- **Unit Tests**: Place them alongside the code being tested (e.g., `service_test.go`).
- **Integration Tests**: Place them in the `tests/` directory. Use the `test_helper.go` in `pkg/database` for setting up test databases.
