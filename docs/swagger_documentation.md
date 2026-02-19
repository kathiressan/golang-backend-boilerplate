# Swagger API Documentation Guide

This project includes an **Automatic Swagger (OpenAPI 3.0) Generator** that visualizes your API routes without requiring manual YAML files or extensive comments.

## 🚀 Accessing Swagger UI

Once the application is running, you can access the interactive API documentation at:

**[http://localhost:8080/swagger/](http://localhost:8080/swagger/)**

> **Note**: The trailing slash `/` is important!

The raw OpenAPI JSON spec is available at:
`http://localhost:8080/swagger/doc.json`

---

## 🛠 How It Works

The system inspects your `RouteRegistry` at runtime. It uses Go reflection to analyze:
1.  **Request Schemas**: The struct passed to `WithSchema()`.
2.  **Response Schemas**: The struct passed to `WithResponseSchema()`.
3.  **Path Parameters**: Extracted from your route path (e.g., `/users/:id`).
4.  **Security**: Based on `ProtectedByJWT()`, `ProtectedByRBAC()`, etc.

### Adding Documentation to a Route

To ensure your endpoint is fully documented, simply add schema information in your route definition:

```go
// internal/api/v1.0/ovmsa/auth/auth.route.go

helpers.POST("/login").
    Unprotected().
    WithSchema(&LoginRequest{}).              // <--- Documents Request Body
    WithResponseSchema(&auth.LoginResult{}).  // <--- Documents Response Body (NEW)
    WithHandler(LoginHandler).
    Build(),
```

*   **`WithSchema(&Struct{})`**: Tells Swagger what JSON body the endpoint expects.
*   **`WithResponseSchema(&Struct{})`**: Tells Swagger what JSON body the endpoint returns on success (200 OK).

---

## 🔐 How to Test Authenticated APIs

Many endpoints are protected (e.g., `ProtectedByJWT`). Swagger UI provides a built-in way to authenticate.

### Step 1: Login
1.  Go to the **Auth** section in Swagger UI.
2.  Expand `POST /ovmsa/v1.0/auth/login`.
3.  Click **Try it out**.
4.  Enter valid credentials in the Request Body:
    ```json
    {
      "email": "admin@example.com",
      "password": "password123"
    }
    ```
5.  Click **Execute**.
6.  Copy the `access_token` from the Response body (without quotes).

### Step 2: Authorize
1.  Scroll to the top of the Swagger page.
2.  Click the **Authorize** button (lock icon).
3.  In the "BearerAuth" box, paste your **access_token**.
    *   *Do not* add the word "Bearer " prefix; the system does it for you.
4.  Click **Authorize** and then **Close**.

### Step 3: Test Protected Routes
Now that you are authorized, you can test protected endpoints:
1.  Go to `GET /ovmsa/v1.0/auth/me`.
2.  Click **Try it out** -> **Execute**.
3.  You should see your user profile instead of a 401 Unauthorized error.

---

## 🧩 Advanced Features

### 1. Simple Types and Arrays
The generator handles standard Go types automatically:
*   `string`, `int`, `bool`, `float64`
*   `time.Time` -> `string (format: date-time)`
*   `uuid.UUID` -> `string (format: uuid)`
*   Slices `[]string` -> `Array of strings`

### 2. Nested Structs
If your request/response struct contains other structs, they will be recursively generated and added to the Swagger `components/schemas` section.

### 3. Validation Tags
The generator respects `json` tags for field names and `binding:"required"` / `validate:"required"` tags to mark fields as **Required** in the documentation.

```go
type CreateUserRequest struct {
    Email string `json:"email" binding:"required"` // Will be marked Required*
    Age   int    `json:"age,omitempty"`            // Will be optional
}
```
