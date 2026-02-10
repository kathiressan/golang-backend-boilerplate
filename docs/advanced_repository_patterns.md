# Advanced Repository Patterns

This document explains the advanced design patterns used in the data access layer, including decorators and custom GORM integrations.

## 🏗 Repository Decorator Pattern

The repository layer utilizes the **Decorator Pattern** to add cross-cutting concerns (like logging, performance tracking, or auditing) without modifying the core repository logic.

### Design Details
- **Interface**: `RepositoryDecorator[T]` mirrors the `BaseRepository` methods.
- **Wrappers**: Decorators wrap a `*BaseRepository[T]` and implement the interface.
- **Composition**: Methods in the decorator call the underlying repository after/before performing their specific logic.

### Example: Logging Decorator
The `LoggingRepositoryDecorator` (in `internal/repository/decorators.go`) automatically logs every operation, its duration, and any errors.

```go
// Usage
userRepo := repository.NewUserRepository(db)
loggingRepo := repository.WithLogging(userRepo, "User")
```

---

## 🛠 Custom GORM Plugins

The system integrates with GORM using custom plugins to handle enterprise requirements like multi-tenancy.

### RLS Plugin (`pkg/database/rls_plugin.go`)
The `RLSPlugin` is registered during database initialization. It intercepts all database operations to set the PostgreSQL session variables required for Row-Level Security.

1.  **Extraction**: It extracts the `*entities.Identity` from the `context.Context`.
2.  **Injection**: It executes `SET LOCAL app.current_org_id = ...` and `SET LOCAL app.current_org_path = ...` within the current transaction.
3.  **Automatic Enforcement**: This ensures that even raw SQL queries or manual GORM queries are scoped to the user's organization.

---

## 🔄 Materialized Path Enforcement

The repository layer and `BaseEntity` work together to ensure that `OrgPath` is always correctly formatted with trailing slashes to prevent prefix collisions (e.g., `/org1` vs `/org10`).

### Enforcement Points:
1.  **BeforeCreate Hook**: In `BaseEntity`, a hook ensures the `OrgPath` ends with a `/`.
2.  **Manual Scopes**: Functions like `ScopeByPath` automatically sanitize the input path.
3.  **RLS Logic**: The RLS policies in Postgres use the `LIKE '/path/%'` operator, which relies on the trailing slash for perfect isolation.

---

## 🧩 Extending the Base Repository

The `BaseRepository` is generic, but can be extended in two ways:

1.  **Specialized Methods**: Define a new struct that embeds `*BaseRepository[T]` and adds domain-specific methods.
2.  **Overriding**: Override base methods if specific behavior is needed for an entity (e.g., soft deletes or custom update logic).

```go
type DocumentRepository struct {
    *BaseRepository[entities.Document]
}

// Custom Extension
func (r *DocumentRepository) FindByTitle(ctx context.Context, title string) (*entities.Document, error) {
    return r.FindOne(ctx, map[string]any{"title": title}, nil)
}
```
