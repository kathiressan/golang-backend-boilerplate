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

For more detailed information, please refer to the following guides:

- **[Getting Started Guide](docs/getting_started_guide.md)**: **Start here!** A beginner-friendly guide to understanding the project, Go, and Postgres.
- **[Swagger API Docs](docs/swagger_documentation.md)**: **New!** How to view and test APIs interactively.
- **[Cheatsheet](docs/cheatsheet.md)**: Common commands for running, testing, and debugging.
- **[Database Schema](docs/database_schema.md)**: Visual guide to the database tables and relationships.
- **[Architecture Overview](docs/architecture.md)**: Deep dive into the system design and request lifecycle.
- **[Development Guide](docs/development.md)**: How to add new features, entities, and services.
- **[Design Patterns Catalog](docs/design_patterns.md)**: Exhaustive list of patterns used (Factory, Builder, Decorator, etc.).
- **[Security & Identity](docs/security_and_identity.md)**: Authentication flows, session management, and replay protection.
- **[Advanced Repository Patterns](docs/advanced_repository_patterns.md)**: Decorators, GORM plugins, and RLS enforcement.
- **[Multi-Tenant Architecture](docs/multi_tenant_architecture.md)**: Explaining the RLS and OrgPath logic.
- **[Repository Layer](docs/repository_layer.md)**: Detailed guide on working with the data access layer.
- **[Configuration Reference](docs/configuration.md)**: Full list of supported environment variables.
- **[Troubleshooting Guide](docs/troubleshooting.md)**: Common pitfalls, RLS debugging, and more.

## ⚖ License

MIT License.
