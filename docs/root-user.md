# Root User Credentials

## Default Root User

A root user is automatically created during database migration for testing and initial setup.

### Credentials

- **Email**: `root@system.local`
- **Password**: `RootPass123!`
- **Role**: System Root (global admin with bypass powers)

### Features

- ✅ Global system administrator
- ✅ Bypasses all RLS policies
- ✅ Can access all organizations
- ✅ Can perform administrative actions (e.g., logout-all)

### Security Notes

> [!CAUTION]
> **IMPORTANT**: This is a default account for development and testing purposes only.
> 
> **In production environments:**
> 1. Change the password immediately after first login
> 2. Consider disabling or removing this account
> 3. Create organization-specific admin accounts instead
> 4. Use strong, unique passwords for all accounts

### Testing Authentication APIs

Use these credentials to test the authentication endpoints:

#### 1. Login
```bash
curl -X POST http://localhost:8000/ovmsa/v1.0/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "root@system.local",
    "password": "RootPass123!"
  }'
```

**Response:**
```json
{
  "status": "success",
  "message": "Success",
  "data": {
    "access_token": "eyJhbGc...",
    "refresh_token": "64-char-hex-string",
    "expires_in": 3600,
    "token_type": "Bearer",
    "user": {
      "id": "user-id",
      "email": "root@system.local",
      "name": "System Root",
      "is_root": true,
      "org_id": "",
      "org_path": "/",
      "role": "root"
    }
  }
}
```

#### 2. Get Current User
```bash
curl -X GET http://localhost:8000/ovmsa/v1.0/auth/me \
  -H "Authorization: Bearer <access_token>"
```

#### 3. Logout All Sessions (Root Only)
```bash
curl -X POST http://localhost:8000/ovmsa/v1.0/auth/logout-all \
  -H "Authorization: Bearer <access_token>"
```

#### 4. Refresh Token
```bash
curl -X POST http://localhost:8000/ovmsa/v1.0/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "<refresh_token>"
  }'
```

#### 5. Logout
```bash
curl -X POST http://localhost:8000/ovmsa/v1.0/auth/logout \
  -H "Authorization: Bearer <access_token>"
```

### Migration Details

The root user is created in migration version `v1.0.1-seed-system-root-user`.

**Migration behavior:**
- Checks if root user already exists before creating
- Skips creation if `root@system.local` already exists in database
- Uses bcrypt hash for password storage (cost factor: 10)

### Changing the Root Password

To change the root password after first login, you'll need to implement a password change endpoint or update directly in the database:

```sql
-- Generate new hash using bcrypt (cost 10)
-- Then update the database:
UPDATE users 
SET password_hash = '<new_bcrypt_hash>' 
WHERE email = 'root@system.local';
```

Or use the password utility in Go:
```go
import "ovmsa-be/pkg/password"

newHash, err := password.HashPassword("YourNewSecurePassword123!")
// Update user record with newHash
```

### Creating Additional Admin Users

For production, create organization-specific admin users instead of relying on the root account:

1. Create a new user via registration endpoint
2. Assign admin role to specific organization
3. Use RBAC to grant appropriate permissions

This follows the principle of least privilege and improves security.
