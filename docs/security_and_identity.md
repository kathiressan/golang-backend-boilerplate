# Security and Identity Deep Dive

This document explains the internal mechanics of the authentication system, identity extraction, and session management.

## 🛡 Authentication Flows

The system supports two primary authentication mechanisms handled by the `AuthMiddleware` in `internal/middleware/identity.go`.

### 1. User Authentication (JWT)
Standard users authenticate via an OIDC/JWT flow. The backend validates the access token using a rotation-aware key lookup.
- **Key Discovery**: Public keys are retrieved from the database based on the `kid` (Version) in the JWT header.
- **Caching**: Successfully retrieved keys are cached in memory (`signingKeyCache`, TTL 5m) to minimize database hits.

### 2. External Service Tokens
Internal or trusted external services use a simplified HMAC-based token format.
- **Format**: `requesterID:timestamp:nonce:signature`
- **Verification**: The signature is generated as `Base64HMAC(timestamp + ":" + nonce, SecretKey)`.
- **Roles**: Requesters are pre-configured in `configs/requesters.go` with specific roles (e.g., `root`, `service-admin`).

---

## 🔒 Identity Consistency and Staleness

To prevent the "Stale Identity" problem (where a user's permissions change but their JWT is still valid), the middleware performs real-time checks.

### Identity Check Cache
The `identityCheckCache` (TTL 1m) stores the results of user/membership validation.
- **Global Check**: Verifies the user's `IsRoot` status matches the JWT.
- **Membership Check**: Verifies the user's role within the specific `OrgID` in the JWT still exists and matches.

### Immediate Revocation
Even before a JWT expires, a session can be invalidated.
- **Session Check**: The middleware extracts the `SessionID` from the JWT and verifies its existence in the `sessions` table.
- **Caching**: This check is cached in `sessionCache` (TTL 30s) to balance security and performance.

---

## 🛡 Replay Protection

For external service tokens, the system implements a strict replay protection mechanism.
- **Replay Cache**: The `replayCache` (TTL 5m) stores the base64-encoded signature of every incoming service token.
- **Time Window**: Tokens are rejected if their timestamp is more than 5 minutes old (or in the future).
- **Abuse Prevention**: If a signature is seen twice within the 5-minute window, the request is immediately aborted with a "Replay detected" error.

---

## 🧩 Contextual Identity (The Identity Card)

Once validated, the identity is stored in the Gin context and the Go `context.Context` as an `*entities.Identity` object.

```go
type Identity struct {
    UserID    string
    SessionID string
    OrgID     string
    OrgPath   string
    Role      string
    IsRoot    bool
}
```

This "Identity Card" is the source of truth for all downstream layers (Services, Repositories, RLS Plugin).
