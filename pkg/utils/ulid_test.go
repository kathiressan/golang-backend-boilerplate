package utils

import (
	"strings"
	"testing"
)

func TestNewULID_should_return_non_empty_string(t *testing.T) {
	id := NewULID()
	if id == "" {
		t.Error("expected a non-empty ULID")
	}
}

func TestNewULID_should_return_26_character_string(t *testing.T) {
	id := NewULID()
	if len(id) != 26 {
		t.Errorf("expected 26 characters, got %d: %q", len(id), id)
	}
}

func TestNewULID_should_return_uppercase_crockford_base32(t *testing.T) {
	// ULID alphabet: 0123456789ABCDEFGHJKMNPQRSTVWXYZ
	const alphabet = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"
	id := NewULID()
	for _, c := range id {
		if !strings.ContainsRune(alphabet, c) {
			t.Errorf("character %q is not in the Crockford Base32 alphabet", c)
		}
	}
}

func TestNewULID_should_produce_unique_values_on_repeated_calls(t *testing.T) {
	seen := make(map[string]struct{}, 100)
	for i := 0; i < 100; i++ {
		id := NewULID()
		if _, exists := seen[id]; exists {
			t.Fatalf("duplicate ULID produced: %q", id)
		}
		seen[id] = struct{}{}
	}
}

func TestNewULID_should_produce_lexicographically_sortable_ids(t *testing.T) {
	// Because ULIDs embed a millisecond timestamp prefix, IDs generated serially
	// (within the same millisecond still share the same prefix, but across
	// milliseconds they must be monotonically sorted).  We can at least assert
	// the first ULID is <= the second when at least 1ms has elapsed.
	// For determinism we just check that IDs generated in sequence are never
	// strictly decreasing.
	a := NewULID()
	b := NewULID()
	if a > b {
		// Allow equal (same millisecond) but never a > b.
		t.Errorf("ULIDs are not sorted: %q > %q", a, b)
	}
}
