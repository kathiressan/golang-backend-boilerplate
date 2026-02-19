# Getting Started Guide

Welcome to the team! 🎉

This guide is designed to help you understand the codebase and get up and running, especially if you are new to Backend development, Go (Golang), or PostgreSQL.

## 1. The Big Picture 🌍

We are building a **Backend API**.
*   **The Client** (Frontend/Mobile App) sends requests (like "Create User" or "Get Products") to us.
*   **We (The API)** process those requests, apply business logic, and talk to the database.
*   **Postgres** is our database where we store all the data safely.

### Why Go?
Go is a simple, compiled language from Google. It is strictly typed (compiler catches errors early) and excellent for building fast, concurrent servers.
*   **Key traits**: No classes (we use `structs`), explicit error handling (`if err != nil`), and simple concurrency.

### Why Postgres?
PostgreSQL is the world's most advanced open-source relational database. It ensures our data is structured and consistent.

---

## 2. Setup Prerequisites 🛠️

Before running the code, ensure you have:

1.  **Go**: [Download & Install](https://go.dev/dl/). (Version 1.25+)
2.  **PostgreSQL**: [Download & Install](https://www.postgresql.org/download/). (Version 14+)
    *   **Important**: Remember the password you set during installation!
3.  **VS Code**: Recommended editor.
    *   Install the **Go** extension by the Go Team.

---

## 3. Running the Project Locally 🏃‍♂️

1.  **Clone/Open the Repo**: You are already here!
2.  **Environment Variables**:
    *   Information like database passwords should never be in the code. We use `.env` files.
    *   Copy the example file: `cp .env.example .env` (or just copy-paste in file explorer).
    *   Open `.env` and check the `DB_PASSWORD`. Set it to what you chose during Postgres installation.
3.  **Download Dependencies**:
    *   Open the terminal in VS Code (`Ctrl + ~`).
    *   Run: `go mod download` (This downloads all the libraries we use).
4.  **Run the Server**:
    *   Run: `go run cmd/app/main.go`
    *   You should see logs saying the server started (usually on port 8080).

---

## 4. Understanding the Folder Structure 📂

This project follows the **Standard Go Project Layout**:

*   **`cmd/app/main.go`**: The entry point. The program starts here.
*   **`internal/`**: Private application code. High-level modules live here.
    *   **`api/`**: **The Front Door**. Handles HTTP requests (Controllers, Routes).
    *   **`services/`**: **The Brain**. Business logic (calculations, rules).
    *   **`repository/`**: **The Librarian**. Talks to the Database (SQL queries).
    *   **`entities/`**: **The Data Shapes**. Structs that match database tables.
*   **`pkg/`**: Public libraries that could be used by other projects (helpers, tools).
*   **`configs/`**: Configuration files.

---

## 5. Key Concepts & Dictionary 📖

You will see these terms often:

*   **Struct**: Custom data types. Like "Classes" in other languages, but simpler.
    ```go
    type User struct {
        Name string
        Age  int
    }
    ```
*   **Interface**: A contract. "If you promise to have a `Save()` method, you are a `Saver`."
*   **Context (`ctx`)**: Passed to almost every function. It holds request info (like "Who is this user?") and controls timeouts. **Always pass it down.**
*   **GORM**: An "ORM" (Object Relational Mapper). It lets us write Go code (`db.Create(&user)`) instead of raw SQL (`INSERT INTO...`).
*   **Middleware**: Code that runs *before* or *after* the main handler. Examples: Logging "Request received", checking "Is user logged in?".
*   **RLS (Row Level Security)**: A security feature in Postgres. It automatically filters data so User A cannot see User B's data, even if the code forgets to check.

---

## 6. How a Request Flows 🌊

The lifecycle of an API call (e.g., `GET /users/1`):

1.  **Route**: The router (`internal/api/routes.go`) sees `GET /users/1` and sends it to the `UserController`.
2.  **Controller**: The `UserController` reads the input (ID: 1) and calls the `UserService`.
3.  **Service**: The `UserService` checks if you are allowed to see User 1. If yes, it asks the `UserRepository`.
4.  **Repository**: The `UserRepository` uses GORM to run a SQL query on the database.
5.  **Database**: Postgres returns the row.
6.  **Response**: The data bubbles back up: Repo -> Service -> Controller -> JSON sent to Client.

---

## 7. Learning Resources 📚

Don't panic! Here are the best places to learn:

*   **[A Tour of Go](https://go.dev/tour/welcome/1)**: (Highly Recommended) An interactive tutorial in your browser.
*   **[Go by Example](https://gobyexample.com/)**: Quick snippets for "How do I do X?".
*   **[Effective Go](https://go.dev/doc/effective_go)**: Tips on writing "idiomatic" Go code.
*   **[GORM Guides](https://gorm.io/docs/index.html)**: Documentation for our database tool.

---

## 8. Common Tasks

> **Tip**: Check out the **[Cheatsheet](../docs/cheatsheet.md)** for a quick list of commands!

### How to add a new route?
Check `internal/api/routes.go`. You'll see how groups and endpoints are defined.

### How to debug?
Use `fmt.Println("I am here")` or the VS Code Debugger. Tests are also a great way to debug specific logic without running the whole server.

### Where are the tests?
Unit tests are usually next to the file they test (e.g., `service_test.go`). Run them with `go test ./...`.

Good luck! If you get stuck, check the specific documentation in the `docs/` folder.
