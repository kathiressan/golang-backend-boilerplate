# Troubleshooting Guide

This document covers common issues, pitfalls, and debugging strategies for developers and AI agents working on this repository.

## 🏢 Multi-Tenancy & RLS Issues

### Issue: "Record not found" or empty results for existing data
If a query successfully executes but returns no data (or `401 Unauthorized`), it is often a Row-Level Security (RLS) mismatch.
- **Cause**: The user's `OrgID` or `OrgPath` in their token does not match the record's metadata.
- **Debugging**:
    1. Check the `current_org_id` and `current_org_path` in the Postgres session.
    2. Verify the record actually has the correct `org_id` and `org_path`.
    3. Ensure the `org_path` has a trailing slash (e.g., `/corp/dept/` not `/corp/dept`).
    4. Try running the query as a `root` user (`IsRoot: true`) to bypass RLS and confirm the data exists.

### Issue: "Duplicate key" on custom OrgPath
- **Cause**: The `OrganizationService` calculates the `org_path` based on a slugified name. If two siblings have names that slugify to the same value, a collision occurs.
- **Fix**: Check `OrganizationService.CheckOrgNameAvailability` before creation.

---

## 🔐 Authentication & Permissions

### Issue: Service tokens failing with "Invalid signature"
- **Cause**: The `SecretKey` must match exactly between the sender and the backend configuration in `configs/requesters.go`.
- **Debugging**: Ensure the timestamp is in Unix seconds and the signature is generated correctly using HMAC-SHA256.

### Issue: "Replay detected" error
- **Cause**: You are reusing a service token signature within a 5-minute window.
- **Fix**: Generate a new `nonce` and `timestamp` for every request.

---

## 🚀 Database & Migrations

### Issue: Deadlock during migrations
- **Cause**: If multiple instances start simultaneously, they might fight for the advisory lock.
- **Fix**: The system uses a 5-second timeout for the lock. Ensure no other manual migrations are running.

### Issue: Column "xyz" does not exist after migration
- **Cause**: You added a field to an entity but forgot to call `AutoMigrate` in your migration action.
- **Fix**: Add `tx.AutoMigrate(&entities.YourModel{})` to the top of the relevant `MigrationStep`.

---

## 🛠 General Debugging Tips

1.  **Check the Logs**: The system uses Zap for structured logging. Check for "error" level logs in the console.
2.  **Database Trace**: Set `ENVIRONMENT=development` to see all SQL queries executed by GORM, including the `SET LOCAL` commands for RLS.
3.  **Purge Caches**: If you've changed permissions in the database but they aren't reflecting, wait 1 minute for the `identityCheckCache` to expire, or restart the server.
