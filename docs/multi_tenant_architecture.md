# Enterprise Multi-Tenant Architecture

This document explains the architecture implemented in this repository for handling multi-tenancy, hierarchical data access, and cross-organization administration using **Go** and **PostgreSQL Row-Level Security (RLS)**.

---

## 1. Multi-Tenancy Strategy: "Shared Schema, Isolated Rows"

The system uses a single PostgreSQL database with a shared schema. Data isolation is enforced at the database level using Row-Level Security (RLS), meaning the isolation is "leaky-proof" even if the application layer has bugs.

### Key Tenancy Models:
1.  **Flat Tenancy**: Standard isolation where Org A cannot see Org B.
2.  **Hierarchical Tenancy**: Supports parent-child organizations (Corp -> Division -> Team). Parents can see children's data, but children cannot see siblings' or parents' data.
3.  **Cross-Org Administration**: Allows specific users (e.g., Support, Auditors) to access data in another organization without becoming a permanent member.

---

## 2. Core Entities & Metadata

### `BaseEntity` (`internal/entities/base.go`)
Every tenanted model must embed `BaseEntity`. This ensures the table contains the metadata required for RLS:
-   **`OrgID`**: The unique identifier for the tenant.
-   **`OrgPath`**: A materialized path (e.g., `/root/parent/child`) used for hierarchical queries.

### `Identity` (`internal/entities/identity.go`)
The "Identity Card" of the requester. This object is populated during authentication and travels through the request context.
-   Contains `UserID`, `OrgID`, `OrgPath`, and `IsRoot` (Super-admin bypass).

---

## 3. Database Enforcement (RLS)

### RLS Policies (`pkg/database/migrations.go`)
During migration, every tenanted table is configured with a `tenant_isolation_policy`. This policy is evaluated by PostgreSQL for every row during `SELECT`, `UPDATE`, or `DELETE`.

**The Isolation Logic:**
A user can access a row if:
1.  `app.is_root` is 'true' (System Admin bypass).
2.  `org_id` matches the user's current session `org_id`.
3.  `org_path` starts with the user's current `org_path` (Hierarchical access).

### GORM RLS Plugin (`pkg/database/rls_plugin.go`)
To connect the Go application to the Postgres RLS, we use a custom GORM plugin. 
Before any query runs, the plugin executes `SET LOCAL` commands to pass the `Identity` data into the current Postgres session transaction:
-   `SET LOCAL app.current_org_id = '...'`
-   `SET LOCAL app.current_org_path = '...'`

For a deeper dive into how repositories handle this data and ensure security, see [Repository Layer Architecture](repository_layer.md).

---

## 4. How to Use in Development

### Defining Tenanted Models
Always embed `entities.BaseEntity` for any model that should be private to an organization.

```go
type Document struct {
    entities.BaseEntity
    Title string `json:"title"`
}
```

### Querying Data
Use `database.ScopedDB` to inject the user's identity into the database handler.

```go
func (s *Service) GetMyDocuments(ctx context.Context) ([]entities.Document, error) {
    id := entities.GetIdentity(ctx)
    var docs []entities.Document
    
    // RLS automatically filters results to the user's OrgID/Path
    err := database.ScopedDB(database.DB, id).Find(&docs).Error
    return docs, err
}
```

### Super-Admin Bypass
Setting `IsRoot: true` in the identity will allow access to all data across all tenants. Use this only for system background jobs or global analytics.

---

## 6. Concrete Examples

### Organization Hierarchy & Pathing

| Organization Name | Level | ID | **OrgPath** |
| :--- | :--- | :--- | :--- |
| **Global Corp** | Root | `org_001` | `/org_001` |
| ∟ **North America** | Division | `org_002` | `/org_001/org_002` |
| &nbsp;&nbsp;&nbsp;∟ **Engineering** | Dept | `org_003` | `/org_001/org_002/org_003` |
| &nbsp;&nbsp;&nbsp;∟ **Sales** | Dept | `org_004` | `/org_001/org_002/org_004` |
| ∟ **Europe** | Division | `org_005` | `/org_001/org_005` |

### Visibility Rules (RLS in Action)

The RLS policy logic (`org_path LIKE app.current_org_path || '%'`) determines what data a user can see:

*   **The "Engineering" Manager** (`/org_001/org_002/org_003`):
    *   **Can See**: Data tagged exactly with their specific path.
    *   **Isolated From**: "Sales" (`/org_004`) and "Europe" (`/org_005`).
*   **The "North America" Executive** (`/org_001/org_002`):
    *   **Can See**: North America, Engineering, and Sales data.
    *   **Why?**: Both child paths start with the North America prefix.
*   **The "Europe" Admin** (`/org_001/org_005`):
    *   **Isolated From**: Everything in North America (`/org_001/org_002/...`).

### Access Control Truth Table

This table summarizes exactly how the RLS engine evaluates access based on the Hierarchy.

| User's Session Path | Target Record's Path | Access | Why? |
| :--- | :--- | :---: | :--- |
| `/org1` | `/org1` | ✅ | **Exact Match**: You always see your own data. |
| `/org1` | `/org1/deptA` | ✅ | **Direct Child**: Parents see their sub-divisions. |
| `/org1` | `/org1/deptA/team1` | ✅ | **Descendant**: Visibility cascades down the whole tree. |
| `/org1/deptA` | `/org1` | ❌ | **Upward Block**: Children cannot see parent (Corporate) data. |
| `/org1/deptA` | `/org1/deptB` | ❌ | **Side Block**: Sibling departments are strictly isolated. |
| `/org1/deptA` | `/org1/deptA/team1` | ✅ | **Child Access**: Dept-level users see their teams. |
| `/org1/deptA/team1` | `/org1/deptA` | ❌ | **Upward Block**: Teams cannot see Dept data. |
| `/` (Special/Root) | `/any/path` | ✅ | **Root Access**: System Admins see the entire tree. |

---

## 7. Security: Delimiter Enforcement

To prevent **Prefix Collisions** (e.g., a user at `/org1` accidentally accessing data at `/org10`), the system strictly enforces trailing slashes on all materialized paths.

### Enforcement Mechanisms:
1.  **GORM Hooks (`BaseEntity`)**: Automatically appends a `/` to `OrgPath` before saving it to the database.
2.  **RLS Plugin (`RLSPlugin`)**: Ensures that the `app.current_org_path` session variable always ends with a `/` before queries are executed.
3.  **GORM Scopes (`ScopeByPath`)**: Appends a `/` to any path provided to the manual scope for consistency.

### Resulting Query Logic:
A request for `/org1` will translate to the following SQL:
```sql
WHERE org_path LIKE '/org1/%%'
```
This correctly matches `/org1/deptA/` but **strictly ignores** `/org10/`.

---

## 8. Data Ownership & Visibility Patterns

### Sibling Isolation Scenario
**Scenario**: You want a "Infrastructure Config" to be visible to the **VP of Engineering** (Parent) and the **DevOps Team** (Child A), but hidden from the **Sales Team** (Sibling B).

**Solution**: Tag the record with the most granular path (**Child A**).

| Actor | Session Path | Record's `OrgPath` | Access | Why? |
| :--- | :--- | :--- | :---: | :--- |
| **Parent** | `/org1/` | `/org1/deptA/` | ✅ | `/org1/deptA/` starts with `/org1/` |
| **Child A** | `/org1/deptA/` | `/org1/deptA/` | ✅ | `/org1/deptA/` starts with `/org1/deptA/` |
| **Sibling B** | `/org1/deptB/` | `/org1/deptA/` | ❌ | No prefix match. |

### Downward Visibility vs Upward Isolation
1.  **Direct Visibility**: A user always sees records where `org_path` matches their session path exactly.
2.  **Downward Visibility**: A user at a higher node (`/org1/`) automatically sees all data in descendant nodes (`/org1/deptA/`, `/org1/deptA/team1/`).
3.  **Upward Isolation**: A user at a lower node (`/org1/deptA/`) **cannot** see data tagged at the parent level (`/org1/`). This is a security feature: a team member should not see corporate-wide secrets unless they are explicitly shared with that team's ID.

### Best Practice: Lease-Privilege Tagging
Always tag records at the **lowest possible level** in the tree where the data is relevant. This ensures the data is isolated from all siblings while remaining visible to the necessary management chain (parents).
