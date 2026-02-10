# Design Patterns Catalog

This document catalogs the various design patterns implemented across the codebase to ensure consistency and modularity.

## 🏗 Behavioral Patterns

### 1. Chain of Responsibility (Error Handling)
Found in `pkg/helpers/errors.go`.
- **Purpose**: Decouples the service error from the API response logic.
- **Implementation**: `ErrorHandlerChain` allows adding multiple `SpecificErrorHandler` instances. The first handler that matches the error type processes the request and aborts the chain.
- **Usage**: Configured in `ControllerFactory` for each controller.

### 2. Strategy Pattern (Authorization)
Found in `internal/api/index.go` and `internal/middleware/permissions.go`.
- **Purpose**: Selects the authorization logic dynamically based on the route configuration.
- **Implementation**: The `ProtectedBy` field (Strategy) determines which check (JWT, RBAC, ABAC, etc.) is executed by the `EnforcePermissions` middleware.

### 3. Observer Pattern (Hooks)
Found in `internal/entities/base.go` and `internal/repository/base_repository.go`.
- **Purpose**: Automatically triggers logic before/after database operations.
- **Implementation**: Utilizes GORM's `BeforeCreate`, `BeforeUpdate`, etc., to enforce `OrgPath` formatting and ULID generation.

---

## 🛠 Creational Patterns

### 4. Builder Pattern (Fluent API)
Found in `pkg/helpers/route_builder.go`.
- **Purpose**: Simplifies the construction of complex `TRouteMatrix` objects.
- **Implementation**: `RouteBuilder` provides fluent methods like `ProtectedByJWT()`, `WithSchema()`, and `WithHandler()`.

### 5. Factory Pattern (Dependency Injection)
Found in `internal/api/controller_factory.go`.
- **Purpose**: Centralizes the creation and wiring of controllers with their services and error chains.
- **Implementation**: `ControllerFactory` encapsulates the logic for initializing specific controllers (e.g., `NewAuthController`).

### 6. Singleton Pattern (Configuration)
Found in `configs/config.go`.
- **Purpose**: Ensures that the application configuration is loaded once and shared globally.
- **Implementation**: Uses `sync.Once` to safe-guard the initialization of the `Config` struct.

---

## 🏗 Structural Patterns

### 7. Decorator Pattern (Cross-cutting Concerns)
Found in `internal/repository/decorators.go`.
- **Purpose**: Adds functionality like logging or auditing to repositories without modifying their core logic.
- **Implementation**: `LoggingRepositoryDecorator` wraps a `BaseRepository` and implements the same interface.
- **See also**: [Advanced Repository Patterns](file:///c:/Users/User/Desktop/golang-backend-boilerplate/docs/advanced_repository_patterns.md).
