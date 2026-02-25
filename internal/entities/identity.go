package entities

import "slices"

// ContextKey is a custom type for context keys to prevent collisions
type ContextKey string

const IdentityCtxKey ContextKey = "identity"

// Identity represents the authenticated requester's context.
// This is the "Identity Card" that travels with every request.
type Identity struct {
	UserID       string         `json:"user_id"`
	SessionID    string         `json:"session_id"`    // Session identifier for revocation checks
	OrgID        string         `json:"org_id"`
	OrgPath      string         `json:"org_path"` // Materialized Path: /corp/dept/team
	Role         string         `json:"role"`
	Attributes   map[string]any `json:"attributes"`     // Dynamic attributes for ABAC
	IsRoot       bool           `json:"is_root"`        // System-wide super admin bypass
	IsOrgAdmin   bool           `json:"is_org_admin"`   // Admin within the current Org
	LinkedOrgIDs []string       `json:"linked_org_ids"` // Cross-org access grants
}

// IsOwnerOf checks if the identity has ownership/admin rights over a specific org.
func (i *Identity) IsOwnerOf(targetOrgID string) bool {
	if i.IsRoot {
		return true
	}
	if i.OrgID == targetOrgID && i.IsOrgAdmin {
		return true
	}
	return slices.Contains(i.LinkedOrgIDs, targetOrgID)
}
