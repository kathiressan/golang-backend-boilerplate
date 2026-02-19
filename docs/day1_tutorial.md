# Day 1 Tutorial: Go & PostgreSQL Fundamentals

Welcome! This tutorial will teach you everything you need to know to work on this repository, assuming zero prior experience with Go or PostgreSQL.

---

## Part 1: Understanding Backend Development

### What is a Backend?

A backend is a server that:

1. **Receives requests** from clients (web browsers, mobile apps)
2. **Processes data** according to business rules
3. **Stores and retrieves data** from a database
4. **Sends responses** back to clients

```
┌─────────┐      Request       ┌─────────┐      Query       ┌─────────┐
│  Client │ ─────────────────> │ Backend │ ────────────────> │ Database│
│ (Browser│                     │  (API)  │                   │(Postgres│
│   App)  │ <───────────────── │         │ <──────────────── │   )    │
└─────────┘      Response      └─────────┘      Results      └─────────┘
```

### What is an API?

API (Application Programming Interface) is how clients talk to our backend. We use **HTTP APIs**:

| Method | Purpose | Example |
|--------|---------|---------|
| GET | Retrieve data | `GET /users/123` - Get user with ID 123 |
| POST | Create new data | `POST /users` - Create a new user |
| PUT | Update data | `PUT /users/123` - Update user 123 |
| DELETE | Remove data | `DELETE /users/123` - Delete user 123 |

### What is JSON?

JSON (JavaScript Object Notation) is the format we use to send data:

```json
{
  "id": "01HQX4Y5Z6K7M8N9P0Q1R2S3T4",
  "name": "John Doe",
  "email": "john@example.com",
  "age": 30,
  "isActive": true,
  "tags": ["developer", "golang"]
}
```

JSON supports:
- **Strings**: `"hello"` (in double quotes)
- **Numbers**: `42`, `3.14`
- **Booleans**: `true`, `false`
- **Null**: `null`
- **Arrays**: `[1, 2, 3]`
- **Objects**: `{"key": "value"}`

---

## Part 2: Go Language Fundamentals

### Why Go?

Go (Golang) is a programming language created by Google. It's designed for:
- **Simplicity**: Only 25 keywords
- **Speed**: Compiled to machine code
- **Concurrency**: Built-in support for parallel processing
- **Reliability**: Strong typing catches errors at compile time

### Installing Go

1. **Download**: Go to https://go.dev/dl/
2. **Install**: Run the installer for your OS
3. **Verify**: Open terminal and run:
   ```bash
   go version
   # Output: go version go1.21.x windows/amd64
   ```

### Go Basics

#### 1. Variables

```go
// Declare and assign
var name string = "John"
var age int = 30

// Short declaration (most common)
email := "john@example.com"
count := 42

// Multiple variables
x, y := 10, 20

// Constants (cannot be changed)
const Pi = 3.14159
```

#### 2. Data Types

```go
// Basic types
var text string = "hello"
var number int = 42
var decimal float64 = 3.14
var isTrue bool = true

// Collections
var numbers []int = []int{1, 2, 3}        // Slice (dynamic array)
var scores map[string]int = map[string]int{  // Map (key-value)
    "alice": 95,
    "bob":   87,
}
```

#### 3. Functions

```go
// Basic function
func sayHello() {
    fmt.Println("Hello!")
}

// Function with parameters
func greet(name string) {
    fmt.Println("Hello, " + name)
}

// Function with return value
func add(a int, b int) int {
    return a + b
}

// Function with multiple returns (very common in Go)
func divide(a, b float64) (float64, error) {
    if b == 0 {
        return 0, errors.New("cannot divide by zero")
    }
    return a / b, nil
}

// Using multiple returns
result, err := divide(10, 2)
if err != nil {
    // Handle error
}
```

#### 4. Structs (Go's "Classes")

Go doesn't have classes. Instead, we use **structs**:

```go
// Define a struct
type User struct {
    ID    string
    Name  string
    Email string
    Age   int
}

// Create an instance
user := User{
    ID:    "user_123",
    Name:  "John Doe",
    Email: "john@example.com",
    Age:   30,
}

// Access fields
fmt.Println(user.Name)  // "John Doe"

// Update fields
user.Age = 31
```

#### 5. Methods (Functions on Structs)

```go
// Method with value receiver
func (u User) GetDisplayName() string {
    return u.Name + " (" + u.Email + ")"
}

// Method with pointer receiver (can modify the struct)
func (u *User) UpdateEmail(newEmail string) {
    u.Email = newEmail
}

// Usage
user := User{Name: "John", Email: "old@example.com"}
user.UpdateEmail("new@example.com")
fmt.Println(user.Email)  // "new@example.com"
```

#### 6. Interfaces

An interface defines what methods a type must have:

```go
// Define an interface
type Writer interface {
    Write(data []byte) error
}

// Any type with a Write method implements Writer
type FileLogger struct{}

func (f FileLogger) Write(data []byte) error {
    // Write to file...
    return nil
}

// Function that accepts any Writer
func Log(w Writer, message string) {
    w.Write([]byte(message))
}
```

#### 7. Pointers

Go passes values by copy. To modify the original, use pointers:

```go
// Without pointer (original not modified)
func incrementWrong(n int) {
    n = n + 1  // Only modifies local copy
}

// With pointer (original is modified)
func increment(n *int) {
    *n = *n + 1  // Dereference and modify
}

num := 5
increment(&num)  // Pass address of num
fmt.Println(num)  // 6
```

Pointer syntax:
- `&variable` - Get the address (pointer)
- `*pointer` - Get the value at the address

#### 8. Error Handling

Go doesn't have exceptions. Errors are returned as values:

```go
// Function that can fail
func readFile(filename string) (string, error) {
    data, err := os.ReadFile(filename)
    if err != nil {
        return "", err  // Return the error
    }
    return string(data), nil  // Return data, no error
}

// Calling code must check for errors
content, err := readFile("config.txt")
if err != nil {
    log.Fatal("Failed to read file:", err)
}
// Use content...
```

**Pattern you'll see everywhere:**
```go
result, err := someFunction()
if err != nil {
    return err  // Propagate error up
}
// Use result...
```

#### 9. Packages and Imports

Go code is organized into packages:

```go
package main  // This file belongs to package "main"

import (
    "fmt"           // Standard library
    "errors"        // Standard library
    "myapp/models"  // Local package
)

func main() {
    fmt.Println("Hello!")
}
```

Package rules:
- Files in the same folder = same package
- Capitalized names are **exported** (public)
- Lowercase names are **unexported** (private)

```go
package models

type User struct {        // Exported (public)
    Name  string         // Exported
    email string         // Unexported (private)
}

func CreateUser() {}      // Exported
func validateEmail() {}   // Unexported
```

#### 10. Control Flow

```go
// If-else
if age >= 18 {
    fmt.Println("Adult")
} else if age >= 13 {
    fmt.Println("Teenager")
} else {
    fmt.Println("Child")
}

// If with initialization
if err := doSomething(); err != nil {
    return err
}

// For loop (Go only has "for")
for i := 0; i < 5; i++ {
    fmt.Println(i)
}

// While-like loop
count := 0
for count < 5 {
    fmt.Println(count)
    count++
}

// Infinite loop
for {
    // Break with: break
    // Continue with: continue
}

// Iterate over slice
names := []string{"Alice", "Bob", "Charlie"}
for index, name := range names {
    fmt.Println(index, name)
}

// Iterate over map
scores := map[string]int{"Alice": 95, "Bob": 87}
for name, score := range scores {
    fmt.Println(name, score)
}

// Switch
switch day {
case "Monday":
    fmt.Println("Start of week")
case "Friday":
    fmt.Println("End of week")
default:
    fmt.Println("Midweek")
}
```

#### 11. Goroutines and Channels (Concurrency)

Go's superpower - easy concurrency:

```go
// Goroutine: Run function in background
go func() {
    fmt.Println("Running in background")
}()

// Channel: Communication between goroutines
ch := make(chan string)

// Send data
go func() {
    ch <- "Hello from goroutine"  // Send
}()

// Receive data
message := <-ch  // Block until received
fmt.Println(message)
```

### Go Project Structure

```
myapp/
├── go.mod          # Dependencies file
├── go.sum          # Dependency checksums
├── cmd/            # Application entry points
│   └── app/
│       └── main.go # Main executable
├── internal/       # Private application code
│   ├── models/     # Data structures
│   ├── handlers/   # HTTP handlers
│   └── services/   # Business logic
└── pkg/            # Public libraries
```

### Go Commands Reference

| Command | Purpose |
|---------|---------|
| `go run main.go` | Compile and run |
| `go build` | Compile to executable |
| `go test ./...` | Run all tests |
| `go mod download` | Download dependencies |
| `go mod tidy` | Clean up dependencies |
| `go get package` | Add a dependency |
| `go fmt ./...` | Format code |

---

## Part 3: PostgreSQL Fundamentals

### What is a Database?

A database is an organized collection of data. PostgreSQL is a **relational database** - data is stored in tables with rows and columns.

```
┌─────────────────────────────────────────────┐
│                  users                      │
├────────────┬──────────┬───────────┬────────┤
│ id         │ name     │ email     │ age    │
├────────────┼──────────┼───────────┼────────┤
│ user_001   │ Alice    │ a@ex.com  │ 25     │ ← Row
│ user_002   │ Bob      │ b@ex.com  │ 30     │
│ user_003   │ Charlie  │ c@ex.com  │ 35     │
└────────────┴──────────┴───────────┴────────┘
     ↑            ↑          ↑         ↑
   Column      Column     Column    Column
```

### Installing PostgreSQL

#### Windows

1. Download from https://www.postgresql.org/download/windows/
2. Run installer, set password for `postgres` user
3. Default port: `5432`
4. Add to PATH (installer usually does this)

#### macOS

```bash
brew install postgresql@14
brew services start postgresql@14
```

#### Linux (Ubuntu)

```bash
sudo apt update
sudo apt install postgresql postgresql-contrib
sudo systemctl start postgresql
```

### Verifying Installation

```bash
# Check if PostgreSQL is running
psql --version
# Output: psql (PostgreSQL) 14.x

# Connect to PostgreSQL
psql -U postgres
# Enter password when prompted
```

### PostgreSQL Basics

#### 1. Databases

A PostgreSQL server can hold multiple databases:

```sql
-- List all databases
\l

-- Create a database
CREATE DATABASE myapp;

-- Connect to a database
\c myapp

-- Drop a database
DROP DATABASE myapp;
```

#### 2. Tables

```sql
-- Create a table
CREATE TABLE users (
    id          VARCHAR(26) PRIMARY KEY,
    name        VARCHAR(100) NOT NULL,
    email       VARCHAR(255) UNIQUE NOT NULL,
    age         INTEGER,
    is_active   BOOLEAN DEFAULT true,
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- List all tables
\dt

-- Describe table structure
\d users

-- Drop a table
DROP TABLE users;
```

#### 3. CRUD Operations

```sql
-- CREATE: Insert data
INSERT INTO users (id, name, email, age)
VALUES ('user_001', 'Alice', 'alice@example.com', 25);

-- READ: Query data
SELECT * FROM users;                          -- All columns
SELECT name, email FROM users;                -- Specific columns
SELECT * FROM users WHERE age > 20;           -- With condition
SELECT * FROM users WHERE name = 'Alice';     -- Exact match
SELECT * FROM users WHERE name LIKE 'A%';     -- Pattern matching
SELECT * FROM users ORDER BY created_at DESC; -- Sorting
SELECT * FROM users LIMIT 10 OFFSET 0;        -- Pagination

-- UPDATE: Modify data
UPDATE users SET age = 26 WHERE id = 'user_001';
UPDATE users SET is_active = false WHERE age < 18;

-- DELETE: Remove data
DELETE FROM users WHERE id = 'user_001';
DELETE FROM users WHERE is_active = false;
```

#### 4. Relationships

```sql
-- One-to-Many: User has many posts
CREATE TABLE posts (
    id          VARCHAR(26) PRIMARY KEY,
    title       VARCHAR(255) NOT NULL,
    content     TEXT,
    user_id     VARCHAR(26) REFERENCES users(id),
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Query with JOIN
SELECT users.name, posts.title
FROM users
JOIN posts ON users.id = posts.user_id;

-- Many-to-Many: Users belong to organizations
CREATE TABLE memberships (
    id          VARCHAR(26) PRIMARY KEY,
    user_id     VARCHAR(26) REFERENCES users(id),
    org_id      VARCHAR(26) REFERENCES organizations(id),
    role        VARCHAR(50) DEFAULT 'member'
);
```

#### 5. Indexes (Performance)

```sql
-- Create an index for faster lookups
CREATE INDEX idx_users_email ON users(email);

-- Create a unique index
CREATE UNIQUE INDEX idx_users_email_unique ON users(email);

-- List indexes
\di
```

#### 6. Transactions

```sql
-- Start a transaction
BEGIN;

-- Multiple operations
INSERT INTO users (id, name) VALUES ('user_002', 'Bob');
INSERT INTO memberships (id, user_id, org_id) VALUES ('mem_001', 'user_002', 'org_001');

-- Commit if successful
COMMIT;

-- Or rollback if something went wrong
ROLLBACK;
```

### psql Command Reference

| Command | Purpose |
|---------|---------|
| `\l` | List databases |
| `\c dbname` | Connect to database |
| `\dt` | List tables |
| `\d tablename` | Describe table |
| `\du` | List users |
| `\q` | Quit |
| `\?` | Help |

### Connection String Format

Applications connect to PostgreSQL using a connection string:

```
postgres://username:password@host:port/database?sslmode=disable
```

Or individual parameters:
```
host=localhost port=5432 user=postgres password=secret dbname=myapp sslmode=disable
```

### Row-Level Security (RLS)

This project uses RLS - a powerful PostgreSQL feature that automatically filters data:

```sql
-- Enable RLS on a table
ALTER TABLE users ENABLE ROW LEVEL SECURITY;

-- Create a policy
CREATE POLICY user_isolation ON users
    USING (org_id = current_setting('app.current_org_id'));

-- Now, every query automatically adds:
-- WHERE org_id = current_setting('app.current_org_id')
```

Users can only see data from their organization - enforced by the database, not the application code.

---

## Part 4: How This Project Uses Go & PostgreSQL

### The Tech Stack

| Layer | Technology | Purpose |
|-------|------------|---------|
| HTTP Server | Gin | Handle HTTP requests/responses |
| ORM | GORM | Go code instead of SQL |
| Database | PostgreSQL | Store and query data |
| Logging | Zap | Structured logging |
| Auth | JWT | Token-based authentication |

### How GORM Works

GORM translates Go code to SQL:

```go
// Go code
db.Where("age > ?", 18).Find(&users)

// Generated SQL
SELECT * FROM users WHERE age > 18;
```

**Common GORM operations:**

```go
// Create
user := User{Name: "Alice", Email: "alice@example.com"}
db.Create(&user)

// Read
var user User
db.First(&user, "id = ?", "user_001")  // SELECT * FROM users WHERE id = 'user_001'

// Read multiple
var users []User
db.Where("age > ?", 18).Find(&users)

// Update
db.Model(&user).Update("name", "Alice Smith")

// Delete
db.Delete(&user)
```

### Request Flow in This Project

```
HTTP Request
    │
    ▼
┌─────────────────────────────────────────────┐
│  Middleware Chain                           │
│  ├─ Request ID (tracking)                   │
│  ├─ Logging (request details)               │
│  ├─ Recovery (panic handling)               │
│  ├─ CORS (cross-origin)                     │
│  ├─ XSS Sanitizer (security)                │
│  ├─ Auth (JWT validation)                   │
│  └─ Permissions (RBAC/ABAC)                 │
└─────────────────────────────────────────────┘
    │
    ▼
┌─────────────────────────────────────────────┐
│  Controller (internal/api/)                 │
│  - Parse request body                       │
│  - Validate input                           │
│  - Call service                             │
│  - Return response                          │
└─────────────────────────────────────────────┘
    │
    ▼
┌─────────────────────────────────────────────┐
│  Service (internal/services/)               │
│  - Business logic                           │
│  - Validation rules                         │
│  - Coordinate repositories                  │
└─────────────────────────────────────────────┘
    │
    ▼
┌─────────────────────────────────────────────┐
│  Repository (internal/repository/)          │
│  - GORM operations                          │
│  - RLS context injection                    │
│  - Data access                              │
└─────────────────────────────────────────────┘
    │
    ▼
┌─────────────────────────────────────────────┐
│  PostgreSQL                                 │
│  - Execute queries                          │
│  - Apply RLS policies                       │
│  - Return results                           │
└─────────────────────────────────────────────┘
```

---

## Part 5: Hands-On Exercise

Let's create a simple endpoint from scratch.

### Step 1: Setup Your Environment

1. **Install Go** (if not done):
   ```bash
   go version
   ```

2. **Install PostgreSQL** (if not done):
   ```bash
   psql --version
   ```

3. **Create a database**:
   ```bash
   # Connect to PostgreSQL
   psql -U postgres
   
   # Create database
   CREATE DATABASE myapp_dev;
   
   # Exit
   \q
   ```

4. **Configure the project**:
   ```bash
   # Navigate to project
   cd golang-backend-boilerplate
   
   # Copy environment file
   cp .env.example .env
   
   # Edit .env and set your database credentials
   # DB_HOST=localhost
   # DB_PORT=5432
   # DB_USER=postgres
   # DB_PASSWORD=your_password
   # DB_NAME=myapp_dev
   ```

5. **Run the project**:
   ```bash
   go mod download
   go run cmd/app/main.go
   ```

   You should see:
   ```
   INFO  Server starting...
   INFO  Connected to database
   INFO  Server running on :8000
   ```

### Step 2: Explore the Codebase

Open these files in your editor:

1. **`cmd/app/main.go`** - Where the app starts
2. **`internal/api/routes.go`** - Where routes are defined
3. **`internal/api/v1.0/ovmsa/auth/auth.controller.go`** - Example controller

### Step 3: Add a Simple "Hello" Endpoint

Create a new file `internal/api/v1.0/ovmsa/hello/hello.go`:

```go
package hello

import (
    "github.com/gin-gonic/gin"
    "ovmsa-be/pkg/response"
)

// HelloHandler returns a greeting
func HelloHandler(c *gin.Context, payload any, identity any, params map[string]string) (any, error, error) {
    return gin.H{
        "message": "Hello, World!",
        "time":    "2024-01-15T10:00:00Z",
    }, nil, nil
}
```

### Step 4: Register the Route

Add to `internal/api/routes.go`:

```go
import "ovmsa-be/internal/api/v1.0/ovmsa/hello"

// In your route registration:
{
    Path:        "/hello",
    Method:      "GET",
    ProtectedBy: entities.UNPROTECTED,
    Controller: entities.TController{
        Handler: hello.HelloHandler,
    },
}
```

### Step 5: Test It

```bash
# Restart the server
go run cmd/app/main.go

# In another terminal, test the endpoint
curl http://localhost:8000/ovmsa/v1.0/hello

# Expected response:
{
    "success": true,
    "message": "Success",
    "data": {
        "message": "Hello, World!",
        "time": "2024-01-15T10:00:00Z"
    }
}
```

### Step 6: Add a Dynamic Endpoint

Let's add an endpoint that takes a parameter:

```go
// HelloNameHandler greets a specific user
func HelloNameHandler(c *gin.Context, payload any, identity any, params map[string]string) (any, error, error) {
    name := c.Param("name")  // Get URL parameter
    return gin.H{
        "message": "Hello, " + name + "!",
    }, nil, nil
}
```

Register with parameter:
```go
{
    Path:        "/hello/:name",  // :name is a parameter
    Method:      "GET",
    ProtectedBy: entities.UNPROTECTED,
    Controller: entities.TController{
        Handler: hello.HelloNameHandler,
    },
}
```

Test:
```bash
curl http://localhost:8000/ovmsa/v1.0/hello/Alice
# {"success":true,"data":{"message":"Hello, Alice!"}}
```

### Step 7: Add Request Validation

Create `internal/api/v1.0/ovmsa/hello/hello.validate.go`:

```go
package hello

// GreetRequest defines the request body
type GreetRequest struct {
    Name string `json:"name" validate:"required,min=2,max=50"`
    Age  int    `json:"age" validate:"omitempty,min=0,max=150"`
}
```

Update the handler:

```go
func GreetHandler(c *gin.Context, payload any, identity any, params map[string]string) (any, error, error) {
    req := payload.(*GreetRequest)  // Type assertion
    
    greeting := "Hello, " + req.Name + "!"
    if req.Age > 0 {
        greeting += fmt.Sprintf(" You are %d years old.", req.Age)
    }
    
    return gin.H{
        "greeting": greeting,
    }, nil, nil
}
```

Register with schema:
```go
helpers.POST("/greet").
    Unprotected().
    WithSchema(&hello.GreetRequest{}).
    WithHandler(hello.GreetHandler).
    Build()
```

Test:
```bash
curl -X POST http://localhost:8000/ovmsa/v1.0/hello/greet \
  -H "Content-Type: application/json" \
  -d '{"name": "Alice", "age": 25}'
  
# Response:
{"success":true,"data":{"greeting":"Hello, Alice! You are 25 years old."}}
```

---

## Part 6: Troubleshooting

### Common Issues

#### "go: command not found"

**Problem**: Go is not in your PATH.

**Solution**:
- Windows: Restart your terminal after installation
- macOS/Linux: Add to `~/.bashrc` or `~/.zshrc`:
  ```bash
  export PATH=$PATH:/usr/local/go/bin
  ```

#### "pq: connection refused"

**Problem**: PostgreSQL is not running.

**Solution**:
```bash
# Windows: Start the PostgreSQL service
net start postgresql-x64-14

# macOS
brew services start postgresql@14

# Linux
sudo systemctl start postgresql
```

#### "pq: database does not exist"

**Problem**: The database hasn't been created.

**Solution**:
```bash
psql -U postgres
CREATE DATABASE myapp_dev;
\q
```

#### "pq: password authentication failed"

**Problem**: Wrong password in `.env`.

**Solution**: Check your `.env` file matches the password you set during PostgreSQL installation.

#### "port 8000 already in use"

**Problem**: Another process is using the port.

**Solution**:
```bash
# Find what's using the port
# Windows
netstat -ano | findstr :8000

# macOS/Linux
lsof -i :8000

# Kill the process or change PORT in .env
```

#### "panic: runtime error"

**Problem**: Unexpected error in code.

**Solution**: Read the stack trace to find the file and line number. Common causes:
- Nil pointer dereference
- Array index out of bounds
- Type assertion on wrong type

### Debugging Tips

1. **Use print statements**:
   ```go
   fmt.Printf("DEBUG: user = %+v\n", user)
   ```

2. **Check the logs**: The server outputs detailed logs including SQL queries.

3. **Use the VS Code debugger**:
   - Set a breakpoint (click left of line number)
   - Press F5 to start debugging

4. **Test in isolation**:
   ```bash
   go test -v -run TestSpecificFunction ./path/to/package
   ```

---

## Part 7: Quick Reference

### Go Cheat Sheet

```go
// Variables
name := "Alice"
age := 30
pi := 3.14
isTrue := true

// Struct
type User struct {
    Name string
    Age  int
}

// Function
func add(a, b int) int {
    return a + b
}

// Error handling
result, err := someFunc()
if err != nil {
    return err
}

// Slice
items := []string{"a", "b", "c"}
items = append(items, "d")

// Map
m := map[string]int{"a": 1, "b": 2}

// Loop
for i, item := range items {
    fmt.Println(i, item)
}

// Condition
if age >= 18 {
    fmt.Println("Adult")
}
```

### SQL Cheat Sheet

```sql
-- Create
INSERT INTO users (id, name) VALUES ('001', 'Alice');

-- Read
SELECT * FROM users;
SELECT * FROM users WHERE age > 18;
SELECT name, email FROM users ORDER BY name;

-- Update
UPDATE users SET name = 'Bob' WHERE id = '001';

-- Delete
DELETE FROM users WHERE id = '001';

-- Join
SELECT users.name, posts.title
FROM users
JOIN posts ON users.id = posts.user_id;
```

### psql Cheat Sheet

```sql
\l              -- List databases
\c dbname       -- Connect to database
\dt             -- List tables
\d tablename    -- Describe table
\du             -- List users
\q              -- Quit
```

---

## Part 8: Next Steps

1. Complete **[A Tour of Go](https://go.dev/tour/)** - Interactive Go tutorial
2. Read **[Getting Started Guide](getting_started_guide.md)** - Project-specific guide
3. Explore the **[Architecture Overview](architecture.md)** - Understand the system
4. Try **[Development Guide](development.md)** - Add a real feature

Good luck! 🚀
