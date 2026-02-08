# Database Migrations

This document explains the versioned migration system used in this repository to manage schema changes, data seeding, and Row-Level Security (RLS) policies.

---

## 1. Core Philosophy

We use a **Versioned + Explicit** migration strategy.

-   **Versioned**: Every migration is assigned a unique version string (e.g., `v1.0.1`). Once applied, it is recorded in the `migration_records` table and never run again.
-   **Explicit**: Large-scale schema changes are defined in these versioned files rather than relying on GORM's ambient `AutoMigrate` throughout the app. This ensures different environments (Dev, Staging, Prod) stay in perfect sync.

---

## 2. The Migration Registry

All migrations are defined in `pkg/database/migrations.go` within the `Migrations` slice.

```go
var Migrations = []MigrationStep{
    {
        Version: "v1.0.0-initial-schema",
        Action:  func(tx *gorm.DB) error { ... },
    },
}
```

### Components of a Migration:
1.  **Version**: A unique string. We recommend using `vX.Y.Z-description`.
2.  **Action**: A function that receives a `*gorm.DB` transaction. If this function returns an error, the entire migration (and its record entry) is rolled back.

---

## 3. Reliability Mechanisms

### Distributed Locking
To prevent multiple instances of the application (e.g., in a Kubernetes cluster) from running migrations simultaneously, the system uses a **PostgreSQL Advisory Lock**:
-   **Lock ID**: `123456789`
-   The application acquires this lock at the start of `Migrate()` and releases it at the end.

### Transactional Integrity
Each migration step runs inside a single database transaction. This ensures that if a step fails halfway through (e.g., a constraint violation), the database is not left in a "partially migrated" state.

---

## 4. How to Add a New Migration

1.  Open `pkg/database/migrations.go`.
2.  Append a new `MigrationStep` to the `Migrations` slice.
3.  Implement your logic.

### Rule of Thumb: `AutoMigrate`
-   **If you are adding a field to a Go struct**: You **must** call `tx.AutoMigrate(&entities.YourModel{})` at the start of the migration action.
-   **If you are only adding data (seeding)**: You can skip `AutoMigrate` if the schema is already up to date.

---

## 5. Integrating Row-Level Security (RLS)

A unique feature of this boilerplate is the RLS integration. To enable RLS on a new table:
1.  Add the table name to the `tenantedTables` list in the initial migration, OR
2.  Call `enableRLS(tx, "table_name")` in your new migration.

```go
func enableRLS(tx *gorm.DB, tableName string) error {
    // Enables RLS and creates the 'tenant_isolation_policy'
}
```

---

## 6. Common Patterns & Troubleshooting

### Error: `column "xyz" does not exist`
This usually happens when you try to insert data into a field you just added to a Go struct.
-   **Fix**: Ensure your migration calls `tx.AutoMigrate(&entities.Model{})` **before** the `tx.Create()` call.

### Manual SQL
For complex changes (indexing, triggers, ltree setup), use raw SQL:
```go
tx.Exec("CREATE INDEX CONCURRENTLY idx_custom ON table (col)")
```

### Migration Order
Migrations are executed strictly in the order they appear in the `Migrations` slice. Never insert a new migration in the middle of the list; always append to the end.
