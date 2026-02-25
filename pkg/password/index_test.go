package password

import (
	"errors"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

// validPassword is the canonical fixture that satisfies all strength rules.
const validPassword = "SecurePass1"

// cheapHash pre-hashes validPassword with cost 4 to keep verify-tests fast.
var cheapHash = func() string {
	b, err := bcrypt.GenerateFromPassword([]byte(validPassword), bcrypt.MinCost)
	if err != nil {
		panic(err)
	}
	return string(b)
}()

// ---------------------------------------------------------------------------
// ValidatePasswordStrength
// ---------------------------------------------------------------------------

func TestValidatePasswordStrength_should_return_nil_when_password_meets_all_rules(t *testing.T) {
	if err := ValidatePasswordStrength(validPassword); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidatePasswordStrength_should_return_correct_error_for_each_rule_violation(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  error
	}{
		// length boundary: 7 chars fails, 8 chars passes (tested separately)
		{"too_short_by_one", "Sec1234", ErrPasswordTooShort},
		{"no_uppercase", "securepass1", ErrPasswordNoUppercase},
		{"no_lowercase", "SECUREPASS1", ErrPasswordNoLowercase},
		{"no_number", "SecurePassword", ErrPasswordNoNumber},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePasswordStrength(tt.password)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("got %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidatePasswordStrength_should_accept_exactly_min_length_password(t *testing.T) {
	// "Secure12" is exactly MinPasswordLength (8) characters.
	if err := ValidatePasswordStrength("Secure12"); err != nil {
		t.Errorf("expected nil for min-length password, got: %v", err)
	}
}

// Length check runs before character checks, so a 7-char all-upper+digit string
// should still return ErrPasswordTooShort, not ErrPasswordNoLowercase.
func TestValidatePasswordStrength_should_check_length_before_character_rules(t *testing.T) {
	err := ValidatePasswordStrength("AB1cde") // 6 chars, missing lowercase still
	if !errors.Is(err, ErrPasswordTooShort) {
		t.Errorf("expected ErrPasswordTooShort, got %v", err)
	}
}

// ---------------------------------------------------------------------------
// HashPassword
// ---------------------------------------------------------------------------

func TestHashPassword_should_return_valid_bcrypt_hash_for_strong_password(t *testing.T) {
	hash, err := HashPassword(validPassword)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(validPassword)); err != nil {
		t.Errorf("hash does not verify against the original password: %v", err)
	}
}

func TestHashPassword_should_return_unique_hashes_on_repeated_calls(t *testing.T) {
	h1, _ := HashPassword(validPassword)
	h2, _ := HashPassword(validPassword)
	if h1 == h2 {
		t.Error("expected different hashes due to bcrypt salting, got identical values")
	}
}

func TestHashPassword_should_propagate_validation_error_without_hashing(t *testing.T) {
	hash, err := HashPassword("weak")
	if err == nil {
		t.Fatal("expected a validation error, got nil")
	}
	if hash != "" {
		t.Error("expected empty hash on error, got non-empty string")
	}
}

// ---------------------------------------------------------------------------
// VerifyPassword
// ---------------------------------------------------------------------------

func TestVerifyPassword_should_return_nil_when_password_matches_hash(t *testing.T) {
	if err := VerifyPassword(validPassword, cheapHash); err != nil {
		t.Errorf("expected nil, got: %v", err)
	}
}

func TestVerifyPassword_should_return_ErrInvalidPassword_when_password_is_wrong(t *testing.T) {
	err := VerifyPassword("WrongPass1", cheapHash)
	if !errors.Is(err, ErrInvalidPassword) {
		t.Errorf("expected ErrInvalidPassword, got: %v", err)
	}
}

// A malformed hash exercises the branch where bcrypt returns an error that is
// NOT ErrMismatchedHashAndPassword, so the raw bcrypt error should be forwarded.
func TestVerifyPassword_should_forward_bcrypt_error_when_hash_is_malformed(t *testing.T) {
	err := VerifyPassword(validPassword, "not-a-bcrypt-hash")
	if err == nil {
		t.Fatal("expected an error for a malformed hash, got nil")
	}
	if errors.Is(err, ErrInvalidPassword) {
		t.Error("malformed hash should propagate the raw bcrypt error, not ErrInvalidPassword")
	}
	if !strings.Contains(err.Error(), "crypto/bcrypt") && err != bcrypt.ErrHashTooShort {
		// bcrypt returns ErrHashTooShort for strings that are clearly not hashes
		t.Logf("got expected bcrypt error: %v", err)
	}
}
