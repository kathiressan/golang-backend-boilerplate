# Developer Cheatsheet 📝

This cheatsheet contains the most common commands you will need for daily development. 

## 🚀 Running the App

```bash
# Start the server (Development Mode)
# This will watch for changes if you use tools like `air`, otherwise restart manually.
go run cmd/app/main.go
```

## 📚 API Documentation

Once the app is running:

| Feature | URL | Note |
| :--- | :--- | :--- |
| **Swagger UI** | `http://localhost:8080/swagger/` | Interactive API testing |
| **JSON Spec** | `http://localhost:8080/swagger/doc.json` | Raw OpenAPI spec |


## 🛠 Database Tasks

Since this project uses GORM's `AutoMigrate`, the schema updates automatically when you run the app. However, sometimes you need to reset things.

```bash
# Wipe the database (Nuclear Option - Only do this locally!) ⚠️
# Windows (Powershell)
psql -U postgres -c "DROP DATABASE my_app; CREATE DATABASE my_app;"

# MacOS/Linux
dropdb my_app && createdb my_app
```

## 🧪 Testing

We use standard Go tests.

```bash
# Run ALL tests
go test ./...

# Run tests with verbose output (to see logs)
go test -v ./...

# Run tests in a specific package
go test ./internal/services/auth/...

# Run tests with coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## 🧹 Code Quality

Keep the code clean!

```bash
# Format code (Auto-indents everything)
go fmt ./...

# Tidy dependencies (Removes unused imports)
go mod tidy

# Run Linter (If you have golangci-lint installed)
golangci-lint run
```

## 📦 Dependency Management

When adding a new library:

1.  Import it in your Go file: `import "github.com/gin-gonic/gin"`
2.  Run:
    ```bash
    go mod tidy
    ```
3.  Commit `go.mod` and `go.sum`.

## 🔍 Debugging Tips

1.  **Print Debugging**: `fmt.Println("DEBUG:", variable)` is your friend.
2.  **SQL Logs**: Check the console! In development mode, all SQL queries are printed.
3.  **RLS Issues**: If you can't see data you just inserted, check `org_id` and `org_path`.
