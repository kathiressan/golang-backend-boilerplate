# Utilities Reference

This document provides a comprehensive reference for all utility packages in the application, including ID generation, string manipulation, password handling, and cryptographic functions.

## Overview

```
pkg/
├── utils/
│   ├── ulid.go      # ULID generation
│   └── slug.go      # URL-friendly string conversion
├── password/
│   └── index.go     # Password hashing and validation
└── cryptography/
    └── index.go     # HMAC and encoding utilities
```

---

## ULID Generation

**Location**: `pkg/utils/ulid.go`

ULID (Universally Unique Lexicographically Sortable Identifier) is used for all primary keys in the database.

### Why ULID?

| Feature | ULID | UUID v4 |
|---------|------|---------|
| Sortable by time | ✅ Yes | ❌ No |
| URL-safe | ✅ Yes | ❌ Contains hyphens |
| Collision-resistant | ✅ 1.21e+24 per ms | ✅ 5.3e+36 total |
| Length | 26 chars | 36 chars |

### Usage

```go
import "ovmsa-be/pkg/utils"

// Generate a new ULID
id := utils.NewULID()
// Output: "01HQX4Y5Z6K7M8N9P0Q1R2S3T4"
```

### Implementation Details

```go
func NewULID() string {
    entropy := crypto/rand.Reader  // Cryptographically secure
    id := ulid.MustNew(ulid.Timestamp(time.Now()), entropy)
    return id.String()
}
```

| Property | Value |
|----------|-------|
| Timestamp | 48 bits (milliseconds since Unix epoch) |
| Randomness | 80 bits (cryptographically secure) |
| Encoding | Crockford's Base32 |
| Length | 26 characters |

### ULID Structure

```
 01HQX4Y5Z6K7M8N9P0Q1R2S3T4
 ├──────┤├──────────────────┤
 Time    Random
 (10ch)  (16ch)
```

### Use Cases

```go
// In entity creation
user := &entities.User{
    ID:        utils.NewULID(),
    Email:     req.Email,
    CreatedAt: time.Now(),
}

// In GORM hooks (automatic)
func (e *BaseEntity) BeforeCreate(tx *gorm.DB) error {
    if e.ID == "" {
        e.ID = utils.NewULID()
    }
    return nil
}
```

---

## Slug Generation

**Location**: `pkg/utils/slug.go`

Converts strings into URL-friendly slugs for use in paths, URLs, and identifiers.

### Usage

```go
import "ovmsa-be/pkg/utils"

slug := utils.Slugify("Hello World!")
// Output: "hello-world"

slug := utils.Slugify("My Organization Name")
// Output: "my-organization-name"

slug := utils.Slugify("Test   Multiple   Spaces")
// Output: "test-multiple-spaces"
```

### Implementation Details

```go
func Slugify(s string) string {
    s = strings.ToLower(s)                      // Lowercase
    s = regexpNonAlphaNumeric.ReplaceAllString(s, "-")  // Replace non-alphanumeric with dash
    s = regexpDashes.ReplaceAllString(s, "")    // Remove leading/trailing dashes
    return s
}
```

### Transformation Rules

| Input | Output | Rule |
|-------|--------|------|
| `Hello World` | `hello-world` | Space → dash |
| `HelloWorld` | `helloworld` | No change |
| `Hello!@#$World` | `hello-world` | Special chars → dash |
| `  Hello  World  ` | `hello-world` | Trim whitespace |
| `---Hello---` | `hello` | Remove surrounding dashes |
| `Test   Multiple` | `test-multiple` | Multiple spaces → single dash |
| `UPPER CASE` | `upper-case` | Lowercase conversion |

### Use Cases

```go
// Organization path generation
func (s *OrgService) CreateOrg(ctx context.Context, name string) (*Organization, error) {
    slug := utils.Slugify(name)
    orgPath := parentPath + "/" + slug
    // ...
}

// URL-safe identifiers
func GenerateInviteCode(name string) string {
    return utils.Slugify(name) + "-" + utils.NewULID()[:8]
}
```

---

## Password Utilities

**Location**: `pkg/password/index.go`

Secure password hashing and validation using bcrypt.

### Constants

```go
const (
    MinPasswordLength = 8          // Minimum password length
    BcryptCost       = 10          // bcrypt cost factor (default)
)
```

### Functions

#### HashPassword

```go
func HashPassword(password string) (string, error)
```

Hashes a password using bcrypt. Automatically validates password strength before hashing.

```go
import "ovmsa-be/pkg/password"

hash, err := password.HashPassword("MySecure123")
if err != nil {
    // Handle validation error
}
// hash: "$2a$10$..."
```

#### VerifyPassword

```go
func VerifyPassword(password, hash string) error
```

Verifies a password against a bcrypt hash.

```go
err := password.VerifyPassword("MySecure123", storedHash)
if err != nil {
    // Password incorrect
    if errors.Is(err, password.ErrInvalidPassword) {
        // Invalid credentials
    }
}
```

#### ValidatePasswordStrength

```go
func ValidatePasswordStrength(password string) error
```

Validates password meets security requirements.

```go
err := password.ValidatePasswordStrength("weak")
if err != nil {
    // err: password.ErrPasswordTooShort
}
```

### Password Requirements

| Requirement | Error |
|-------------|-------|
| Minimum 8 characters | `ErrPasswordTooShort` |
| At least one uppercase | `ErrPasswordNoUppercase` |
| At least one lowercase | `ErrPasswordNoLowercase` |
| At least one number | `ErrPasswordNoNumber` |

### Error Types

```go
var (
    ErrPasswordTooShort    = errors.New("password must be at least 8 characters long")
    ErrPasswordNoUppercase = errors.New("password must contain at least one uppercase letter")
    ErrPasswordNoLowercase = errors.New("password must contain at least one lowercase letter")
    ErrPasswordNoNumber    = errors.New("password must contain at least one number")
    ErrInvalidPassword     = errors.New("invalid password")
)
```

### bcrypt Details

| Property | Value |
|----------|-------|
| Algorithm | bcrypt |
| Cost Factor | 10 (2^10 iterations) |
| Salt | Auto-generated per hash |
| Hash Length | 60 characters |

### Use Cases

```go
// User registration
func (s *AuthService) Register(ctx context.Context, req *RegisterRequest) error {
    hash, err := password.HashPassword(req.Password)
    if err != nil {
        return errors.BadRequest(err, "Invalid password")
    }
    
    user := &entities.User{
        ID:           utils.NewULID(),
        Email:        req.Email,
        PasswordHash: hash,
    }
    // ...
}

// User login
func (s *AuthService) Login(ctx context.Context, email, password string) (*User, error) {
    user, err := s.repo.FindByEmail(ctx, email)
    if err != nil {
        return nil, errors.Unauthorized(nil, "Invalid credentials")
    }
    
    if err := password.VerifyPassword(password, user.PasswordHash); err != nil {
        return nil, errors.Unauthorized(nil, "Invalid credentials")
    }
    // ...
}
```

---

## Cryptography Utilities

**Location**: `pkg/cryptography/index.go`

HMAC-based cryptographic functions for secure token generation and verification.

### Functions

#### Base64HMAC

```go
func Base64HMAC(data, key string) string
```

Generates a base64-encoded HMAC-SHA256 signature.

```go
import cryptographyHelper "ovmsa-be/pkg/cryptography"

signature := cryptographyHelper.Base64HMAC("message", "secret-key")
// Output: "8JiBc8...base64-encoded-hash..."
```

### Implementation Details

```go
func Base64HMAC(data, key string) string {
    h := hmac.New(sha256.New, []byte(key))
    h.Write([]byte(data))
    return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
```

| Property | Value |
|----------|-------|
| Algorithm | HMAC-SHA256 |
| Output | Base64-encoded string |
| Key Length | Variable |
| Hash Length | 32 bytes (256 bits) |

### Use Cases

#### External Service Token Verification

```go
// Token format: requesterID:timestamp:nonce:signature
// Signature = Base64HMAC(timestamp:nonce, secretKey)

expectedSignature := cryptographyHelper.Base64HMAC(
    timestamp + ":" + nonce,
    requester.SecretKey,
)

if expectedSignature != providedSignature {
    return errors.Unauthorized(nil, "Invalid token signature")
}
```

#### Token Hashing for Storage

```go
// Hash refresh tokens before storing
hashedToken := cryptographyHelper.Base64HMAC(token, config.GetConfig().JWTSecret)
session.RefreshTokenHash = hashedToken
```

#### Request Signing

```go
// Sign API requests
func SignRequest(requesterID, secretKey string) string {
    timestamp := strconv.FormatInt(time.Now().Unix(), 10)
    nonce := generateNonce()
    signature := cryptographyHelper.Base64HMAC(timestamp+":"+nonce, secretKey)
    
    return fmt.Sprintf("%s:%s:%s:%s", requesterID, timestamp, nonce, signature)
}
```

---

## Entity Integration

### BaseEntity ULID Generation

ULIDs are automatically generated in the `BeforeCreate` hook:

```go
// internal/entities/base.go
func (e *BaseEntity) BeforeCreate(tx *gorm.DB) error {
    if e.ID == "" {
        e.ID = utils.NewULID()
    }
    // ...
}
```

### Organization Path Generation

Slugs are used for organization path generation:

```go
// internal/services/org/organization_service.go
func (s *OrgService) Create(ctx context.Context, req *CreateOrgRequest) (*Organization, error) {
    slug := utils.Slugify(req.Name)
    
    // Check availability
    if !s.CheckSlugAvailability(ctx, slug, parentID) {
        return nil, errors.Conflict(nil, "Organization name not available")
    }
    
    // Build path
    orgPath := parentPath + "/" + slug
    
    org := &Organization{
        ID:      utils.NewULID(),
        Name:    req.Name,
        OrgPath: orgPath,
        // ...
    }
    // ...
}
```

---

## Best Practices

### ULID Usage

```go
// Do: Let BeforeCreate handle it
user := &entities.User{Email: email}
db.Create(user)  // ID auto-generated

// Do: Explicit generation when needed
user := &entities.User{
    ID:    utils.NewULID(),  // Explicit
    Email: email,
}

// Don't: Use UUID or auto-increment
user := &entities.User{
    ID: uuid.New().String(),  // Wrong - use ULID
}
```

### Slug Usage

```go
// Do: Use for URL-safe identifiers
slug := utils.Slugify(orgName)

// Don't: Use for primary keys
id := utils.Slugify(name)  // Wrong - use ULID for IDs

// Do: Check availability before creating
if !isSlugAvailable(slug) {
    return errors.Conflict(nil, "Name already taken")
}
```

### Password Usage

```go
// Do: Always hash before storage
hash, _ := password.HashPassword(plainPassword)
user.PasswordHash = hash

// Do: Use constant-time comparison (bcrypt does this)
err := password.VerifyPassword(inputPassword, storedHash)

// Don't: Compare hashes directly
if inputHash == storedHash {  // Wrong - timing attack
}

// Don't: Store plain text passwords
user.Password = plainPassword  // NEVER!
```

### Cryptography Usage

```go
// Do: Use for signing/verification
signature := cryptographyHelper.Base64HMAC(data, secretKey)

// Do: Use different keys for different purposes
apiKey := config.APISecretKey
tokenKey := config.TokenSecretKey

// Don't: Use for password hashing
hash := cryptographyHelper.Base64HMAC(password, key)  // Wrong - use bcrypt
```

---

## Testing Utilities

### ULID Generation in Tests

```go
func TestUserCreation(t *testing.T) {
    id := utils.NewULID()
    
    assert.Len(t, id, 26)
    assert.Regexp(t, `^[0-9A-Z]+$`, id)
}
```

### Slug Testing

```go
func TestSlugify(t *testing.T) {
    tests := []struct {
        input    string
        expected string
    }{
        {"Hello World", "hello-world"},
        {"Test@123!", "test-123"},
        {"  Spaces  ", "spaces"},
    }
    
    for _, tt := range tests {
        assert.Equal(t, tt.expected, utils.Slugify(tt.input))
    }
}
```

### Password Testing

```go
func TestPasswordValidation(t *testing.T) {
    tests := []struct {
        password string
        wantErr  error
    }{
        {"Short1!", password.ErrPasswordTooShort},
        {"nouppercase1!", password.ErrPasswordNoUppercase},
        {"NOLOWERCASE1!", password.ErrPasswordNoLowercase},
        {"NoNumbers!", password.ErrPasswordNoNumber},
        {"ValidPass123", nil},
    }
    
    for _, tt := range tests {
        err := password.ValidatePasswordStrength(tt.password)
        assert.ErrorIs(t, err, tt.wantErr)
    }
}
```
