# Logging System

This document explains the structured logging system, including log levels, formatting, environment-specific behavior, and best practices.

## Overview

The application uses **Zap** as its logging library, wrapped with custom formatting and convenience methods. Zap provides:
- **High performance** - Zero-allocation logging
- **Structured output** - Key-value pairs for easy parsing
- **Flexible formatting** - JSON for production, colored console for development

---

## Log Levels

| Level | Color | Usage |
|-------|-------|-------|
| **DEBUG** | Blue | Detailed information for debugging (dev only) |
| **INFO** | Green | General operational information |
| **WARN** | Yellow | Warning conditions that might indicate problems |
| **ERROR** | Red | Error conditions that need attention |
| **FATAL** | Purple | Severe errors that require application shutdown |

### Log Level by Environment

| Environment | Minimum Level |
|-------------|---------------|
| Development | DEBUG |
| Production | INFO |

---

## Basic Usage

**Location**: `pkg/logger/index.go`

### Simple Logging

```go
import "ovmsa-be/pkg/logger"

logger.Info("Server started", "port", 8080)
logger.Debug("Processing request", "userID", "user_123")
logger.Warn("High memory usage", "usageMB", 1024)
logger.Error("Database connection failed", "error", err)
logger.Fatal("Critical startup failure", "error", err)  // Exits application
```

### Structured Fields

Pass key-value pairs for structured logging:

```go
logger.Info("User registered",
    "userID", "user_123",
    "email", "user@example.com",
    "orgID", "org_456",
)
```

### Formatted Messages

```go
logger.Infof("Processing %d items for user %s", count, userID)
```

---

## Development vs Production

### Development Output

Colored console output with human-readable format:

```
2024-01-15 10:30:45.123 INFO  Request completed  requestID=abc-123  clientIP=192.168.1.1  method=POST  path=/api/users  status=201 Created (Success)  latency=45ms
```

**Features**:
- ANSI colors for log levels
- Timestamp in dim gray
- Caller location in cyan
- Status codes formatted with HTTP status text
- Code line shown for errors

### Production Output

JSON format for log aggregation systems:

```json
{
    "level": "info",
    "timestamp": "2024-01-15T10:30:45.123Z",
    "caller": "api/index.go:45",
    "message": "Request completed",
    "requestID": "abc-123",
    "clientIP": "192.168.1.1",
    "method": "POST",
    "path": "/api/users",
    "status": 201,
    "latency": "45ms"
}
```

---

## Contextual Logging

### Creating Context Loggers

Create loggers with preset fields for consistent context:

```go
// Create a logger with request context
requestLogger := logger.With("requestID", requestID, "userID", userID)

// All logs from this logger include the context
requestLogger.Info("Processing payment")
requestLogger.Info("Payment successful", "amount", 99.99)
requestLogger.Error("Payment failed", "error", err)
```

### In Services

```go
type UserService struct {
    logger *zap.SugaredLogger
}

func NewUserService() *UserService {
    return &UserService{
        logger: logger.With("service", "UserService"),
    }
}

func (s *UserService) CreateUser(ctx context.Context, req *CreateUserRequest) error {
    s.logger.Info("Creating user", "email", req.Email)
    // ...
}
```

---

## Error Logging

### Basic Error Logging

```go
if err != nil {
    logger.Error("Failed to create user", "error", err)
}
```

### Enhanced Error Output

The logger automatically:
1. Appends error message to log message
2. Shows the source code line that caused the error
3. Adds a visual separator for errors

**Example Output**:
```
==================== 🚨 ERROR 🚨 ====================
2024-01-15 10:30:45.123 ERROR user_service.go:45  Failed to create user: connection refused | code: if err := db.Create(&user).Error; err != nil {
```

### Stack Traces

Stack traces are automatically included for ERROR and FATAL levels in all environments:

```go
logger.Error("Database error", "error", err)
// Includes full stack trace after the log entry
```

---

## HTTP Request Logging

The middleware automatically logs all HTTP requests:

```go
// In middleware/logger.go
logger.Info("Request completed",
    "requestID", requestID,
    "clientIP", clientIP,
    "method", method,
    "path", path,
    "status", statusCode,
    "latency", latency,
    "userAgent", userAgent,
    "errorCount", len(c.Errors),
)
```

### Status Code Formatting

HTTP status codes are automatically formatted:

| Status | Output |
|--------|--------|
| 200 | `200 Success (OK)` |
| 201 | `201 Success (Created)` |
| 400 | `400 Client Error (Bad Request)` |
| 404 | `404 Client Error (Not Found)` |
| 500 | `500 Server Error (Internal Server Error)` |

Error status codes (≥400) are logged at ERROR level with colored output.

---

## Log Fields Reference

### Standard Fields

| Field | Description | Example |
|-------|-------------|---------|
| `requestID` | Unique request identifier | `abc-123-def` |
| `userID` | Authenticated user ID | `user_123` |
| `orgID` | Organization ID | `org_456` |
| `clientIP` | Client IP address | `192.168.1.1` |
| `method` | HTTP method | `POST` |
| `path` | Request path | `/api/users` |
| `status` | HTTP status code | `201` |
| `latency` | Request duration | `45ms` |
| `error` | Error object | `connection refused` |

### Custom Fields

Add any field relevant to your context:

```go
logger.Info("Order processed",
    "orderID", "order_789",
    "items", 5,
    "total", 199.99,
    "paymentMethod", "credit_card",
)
```

---

## Initialization

The logger must be initialized at application startup:

```go
// cmd/app/main.go
func main() {
    cfg := config.GetConfig()
    logger.Initialize(cfg.Environment)
    defer logger.Sync()  // Flush logs before exit
    
    // ... rest of application
}
```

### Configuration

```go
func Initialize(environment config.Environment) {
    once.Do(func() {
        // Configure based on environment
        if env == string(config.EnvProduction) {
            encoder = zapcore.NewJSONEncoder(encoderConfig)
            logLevel = zap.InfoLevel
        } else {
            encoder = zapcore.NewConsoleEncoder(encoderConfig)
            logLevel = zap.DebugLevel
        }
        // ...
    })
}
```

---

## Best Practices

### Do

```go
// Use structured fields
logger.Info("User logged in", "userID", user.ID, "orgID", user.OrgID)

// Log at appropriate levels
logger.Debug("Cache hit", "key", cacheKey)
logger.Info("User created", "userID", user.ID)
logger.Warn("Rate limit approaching", "requests", 950, "limit", 1000)
logger.Error("Database error", "error", err)

// Include context
logger.Error("Failed to process order",
    "error", err,
    "orderID", orderID,
    "userID", userID,
)

// Use With for repeated context
orderLogger := logger.With("orderID", orderID, "userID", userID)
orderLogger.Info("Processing payment")
orderLogger.Info("Payment successful")
```

### Don't

```go
// Don't use fmt.Printf or log.Print
fmt.Printf("User %s logged in\n", userID)  // Wrong!

// Don't log sensitive data
logger.Info("User login", "password", password)  // NEVER!

// Don't log at wrong levels
logger.Error("User logged in", "userID", userID)  // Should be Info

// Don't create unstructured messages
logger.Info(fmt.Sprintf("User %s logged in", userID))  // Use fields instead

// Don't log excessively in hot paths
for _, item := range items {
    logger.Debug("Processing item", "itemID", item.ID)  // Can be expensive
}
```

---

## Performance Considerations

### Zero-Allocation Logging

Zap's structured logging is designed for zero-allocation in the hot path:

```go
// Efficient: key-value pairs
logger.Info("Processing request", "requestID", id, "count", count)

// Less efficient: fmt.Sprintf
logger.Infof("Processing request %s with count %d", id, count)
```

### Debug Logs in Production

Debug logs are automatically disabled in production, so you can safely leave them in code:

```go
// This is a no-op in production (log level < INFO)
logger.Debug("Detailed trace", "step", step, "state", state)
```

### Synchronous vs Asynchronous

Logs are written synchronously. For high-throughput applications, consider:
- Using a log aggregation service (ELK, Datadog, etc.)
- Batching logs in the logging infrastructure

---

## Integration with Error Handling

The logging system integrates with the error handling system:

```go
// In pkg/response/index.go
if appErr, ok := err.(*appErrors.AppError); ok {
    logger.Error("Request error",
        "requestID", requestID,
        "statusCode", appErr.StatusCode,
        "errorCode", appErr.Code,
        "errorType", appErr.Type,
        "message", appErr.Message,
    )
}
```

---

## Testing with Logs

### Capturing Logs in Tests

For testing, you can check log output:

```go
func TestUserCreation_LogsUserID(t *testing.T) {
    // Setup
    var buf bytes.Buffer
    // Redirect log output for testing (implementation depends on your test setup)
    
    // Execute
    service.CreateUser(ctx, req)
    
    // Verify log contains expected fields
    // Note: In practice, use a log buffer or mock
}
```

### Log Assertion Libraries

Consider using libraries like `zaptest` for testing:

```go
import "go.uber.org/zap/zaptest"

func TestWithLogCapture(t *testing.T) {
    logger := zaptest.NewLogger(t)
    // Use this logger in your test
}
```

---

## Common Patterns

### Request Tracing

```go
func (s *Service) ProcessOrder(ctx context.Context, orderID string) error {
    log := logger.With("orderID", orderID)
    
    log.Info("Processing order")
    
    if err := s.validateOrder(orderID); err != nil {
        log.Error("Validation failed", "error", err)
        return err
    }
    
    log.Info("Order processed successfully")
    return nil
}
```

### Error Context Chain

```go
func (s *Service) CreateUser(ctx context.Context, req *CreateUserRequest) error {
    log := logger.With("email", req.Email)
    
    user, err := s.repo.FindByEmail(ctx, req.Email)
    if err != nil {
        log.Error("Failed to find existing user", "error", err)
        return err
    }
    
    if user != nil {
        log.Warn("User already exists")
        return errors.Conflict(nil, "User already exists")
    }
    
    log.Info("Creating new user")
    // ...
}
```

### Graceful Shutdown Logging

```go
func main() {
    logger.Initialize(cfg.Environment)
    defer logger.Sync()
    
    // ... setup
    
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    
    <-quit
    logger.Info("Shutting down server", "signal", sig)
    
    // ... cleanup
    
    logger.Info("Server stopped")
}
```
