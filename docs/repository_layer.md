# Repository Layer Architecture

This document explains the design, security, and usage of the repository layer in this boilerplate. The repository layer acts as an abstraction over **GORM**, providing a consistent interface for data access while enforcing security and multi-tenancy rules.

---

## 1. Core Architecture

### `BaseRepository[T]`
The foundation of the repository layer is the `BaseRepository`, found in `internal/repository/base_repository.go`. This generic struct provides standard CRUD operations for any entity type.

**Key Features:**
- **Generic CRUD**: `Create`, `FindByID`, `FindOne`, `FindAll`, `List` (paginated), `Update`, `Delete`.
- **Preloading**: Helper method `applyPreloads` to easily include related data.
- **Transaction Support**: All methods accept an optional `*gorm.DB` transaction.

### Repository Aggregation
To simplify dependency injection, all specialized repositories are aggregated into a `Repositories` struct in `internal/repository/repository.go`. The global `Repo` variable provides easy access throughout the application after initialization.

---

## 2. Security Patterns

### SQL Injection Prevention
The system uses multiple layers to prevent SQL injection, especially in dynamic queries:
1. **Parameterized Queries**: Standard GORM methods are used with parameters (e.g., `.Where("id = ?", id)`).
2. **Identifier Sanitization**: The `sanitizeIdentifier` method uses regex to validate dynamic field names or sort parameters (e.g., in `List` or `UpdateFields`).
3. **Condition Validation**: The `validateConditions` method ensures that keys in maps (used for queries) are safe identifiers.

### Safe Sorting and Filtering
When exposing sorting or filtering to API consumers, identifiers are always passed through the sanitizer:
```go
// Example from List method
if sort != "" {
    if err := r.sanitizeIdentifier(sort); err != nil {
        return nil, 0, appErrors.BadRequest(err, "Invalid sort parameter")
    }
    query = query.Order(sort)
}
```

---

## 3. Row-Level Security (RLS) Integration

The repository layer is the bridge between the application context and the database's RLS engine.

### Passing Identity Context
The `getDB` helper method automatically attaches the request `context.Context` to the GORM instance. This ensures that the [RLS Plugin](file:///c:/Users/User/Desktop/golang-backend-boilerplate/pkg/database/rls_plugin.go) can extract the `Identity` from the context and set the appropriate session variables in PostgreSQL.

---

## 4. Error Handling

Individual repositories do not return raw GORM errors. Instead, they use the `wrapError` helper to map them to standardized application errors defined in `pkg/errors`:
- `gorm.ErrRecordNotFound` -> `appErrors.NotFound`
- Other errors -> `appErrors.InternalServerError`

---

## 5. Usage Example

### Creating a Specialized Repository
Specialized repositories embed the `BaseRepository` and add domain-specific logic.

```go
type UserRepository struct {
    *BaseRepository[entities.User]
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*entities.User, error) {
    return r.FindOne(ctx, map[string]any{"email": email}, nil)
}
```

### Accessing Repositories
Repositories should be accessed via the global `Repo` instance after `repository.Initialize(db)` has been called in `main.go`.

```go
user, err := repository.Repo.User.FindByID(ctx, userID, nil)
```
