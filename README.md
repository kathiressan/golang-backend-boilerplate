# Go Backend Boilerplate

A robust, enterprise-ready Go backend boilerplate featuring a multi-tenant hierarchical architecture, Postgres Row-Level Security (RLS), and a generic repository pattern.

## 🚀 Key Features

- **Multi-Tenancy**: Built-in support for hierarchical organizations with strictly isolated data.
- **Postgres RLS**: Security enforcement at the database level using Row-Level Security.
- **Generic Repository**: Abstracted data access layer with built-in sanitization and pagination.
- **Modern Tech Stack**: Gin (HTTP), GORM (ORM), Zap (Logging), JWT (Auth).
- **Graceful Shutdown**: Handles OS signals for clean server termination.

## 📂 Project Structure

```text
├── cmd/                # Entry points
│   └── app/            # Main application
├── configs/            # Configuration and static data
├── docs/               # Technical documentation
├── internal/           # Private application code
│   ├── api/            # API handlers and routing
│   ├── entities/       # Domain models (GORM)
│   ├── middleware/     # Gin middleware
│   ├── repository/     # Data access layer
│   └── services/       # Business logic
├── pkg/                # Public shared libraries
└── tests/              # Integration and unit tests
```

## 🛠 Prerequisites

- **Go**: 1.25+
- **PostgreSQL**: 14+
- **Tools**: `npx` (for optional frontend work)

## 🚦 Quick Start

1. **Clone the repository**:
   ```bash
   git clone <repo-url>
   cd golang-backend-boilerplate
   ```

2. **Configure environment**:
   Copy `.env.example` to `.env` and fill in your database credentials.

3. **Install dependencies**:
   ```bash
   go mod download
   ```

4. **Run the application**:
   ```bash
   go run cmd/app/main.go
   ```

## 📚 Documentation

### Getting Started

> **New to Go or PostgreSQL?** Start with **[Day 1 Tutorial](docs/day1_tutorial.md)** - a complete guide that teaches Go and PostgreSQL from scratch with hands-on exercises.

- **[Day 1 Tutorial](docs/day1_tutorial.md)**: **Complete beginners start here!** Go & PostgreSQL fundamentals with hands-on exercises.
- **[Getting Started Guide](docs/getting_started_guide.md)**: Project overview and quick setup guide.
- **[Cheatsheet](docs/cheatsheet.md)**: Common commands for running, testing, and debugging.
- **[Root User Credentials](docs/root-user.md)**: Default admin account for testing.

### Architecture & Design
- **[Architecture Overview](docs/architecture.md)**: Deep dive into system design and request lifecycle.
- **[Design Patterns Catalog](docs/design_patterns.md)**: Patterns used (Factory, Builder, Decorator, etc.).
- **[Multi-Tenant Architecture](docs/multi_tenant_architecture.md)**: RLS and hierarchical organization logic.
- **[Database Schema](docs/database_schema.md)**: Visual guide to tables and relationships.
- **[Database Migrations](docs/database_migrations.md)**: Versioned migration system.

### Data Layer
- **[Repository Layer](docs/repository_layer.md)**: Guide to working with the data access layer.
- **[Advanced Repository Patterns](docs/advanced_repository_patterns.md)**: Decorators, GORM plugins, RLS enforcement.

### Security
- **[Security & Identity](docs/security_and_identity.md)**: Authentication flows, session management, replay protection.
- **[Security Hardening](docs/security_hardening.md)**: XSS, CORS, rate limiting, security headers.
- **[Permissions System](docs/permissions.md)**: RBAC, ABAC, and combined authorization strategies.

### Core Systems
- **[Middleware System](docs/middleware.md)**: Complete middleware chain documentation.
- **[Error Handling](docs/error_handling.md)**: AppError system, error codes, response formats.
- **[Request Validation](docs/validation.md)**: Two-stage validation, tags, formatters.
- **[Logging System](docs/logging.md)**: Structured logging, log levels, context.

### Development
- **[Development Guide](docs/development.md)**: How to add new features, entities, and services.
- **[Utilities Reference](docs/utilities.md)**: ULID, Slug, Password, Crypto utilities.
- **[Testing Guide](docs/testing.md)**: Unit tests, integration tests, test database setup.
- **[Configuration Reference](docs/configuration.md)**: Full list of environment variables.

### API Documentation
- **[Swagger API Docs](docs/swagger_documentation.md)**: How to view and test APIs interactively.

### Troubleshooting
- **[Troubleshooting Guide](docs/troubleshooting.md)**: Common pitfalls, RLS debugging, and more.

## ⚖ License

MIT License.
