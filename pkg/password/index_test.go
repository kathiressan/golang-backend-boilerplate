package password

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name        string
		password    string
		shouldError bool
	}{
		{
			name:        "valid strong password",
			password:    "SecurePass123",
			shouldError: false,
		},
		{
			name:        "password too short",
			password:    "Short1",
			shouldError: true,
		},
		{
			name:        "no uppercase",
			password:    "securepass123",
			shouldError: true,
		},
		{
			name:        "no lowercase",
			password:    "SECUREPASS123",
			shouldError: true,
		},
		{
			name:        "no number",
			password:    "SecurePassword",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := HashPassword(tt.password)
			if tt.shouldError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if hash == "" {
					t.Errorf("expected hash but got empty string")
				}
				// Verify the hash is a valid bcrypt hash
				if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(tt.password)); err != nil {
					t.Errorf("generated hash is not valid: %v", err)
				}
			}
		})
	}
}

func TestHashPassword_DifferentHashes(t *testing.T) {
	password := "SecurePass123"
	hash1, err1 := HashPassword(password)
	hash2, err2 := HashPassword(password)

	if err1 != nil || err2 != nil {
		t.Fatalf("unexpected errors: %v, %v", err1, err2)
	}

	// Bcrypt should produce different hashes for the same password (due to random salt)
	if hash1 == hash2 {
		t.Errorf("expected different hashes for same password, got identical hashes")
	}
}

func TestVerifyPassword(t *testing.T) {
	password := "SecurePass123"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	tests := []struct {
		name        string
		password    string
		hash        string
		shouldError bool
	}{
		{
			name:        "correct password",
			password:    password,
			hash:        hash,
			shouldError: false,
		},
		{
			name:        "incorrect password",
			password:    "WrongPassword123",
			hash:        hash,
			shouldError: true,
		},
		{
			name:        "empty password",
			password:    "",
			hash:        hash,
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := VerifyPassword(tt.password, tt.hash)
			if tt.shouldError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestValidatePasswordStrength(t *testing.T) {
	tests := []struct {
		name        string
		password    string
		expectedErr error
	}{
		{
			name:        "valid password",
			password:    "SecurePass123",
			expectedErr: nil,
		},
		{
			name:        "too short",
			password:    "Short1",
			expectedErr: ErrPasswordTooShort,
		},
		{
			name:        "no uppercase",
			password:    "securepass123",
			expectedErr: ErrPasswordNoUppercase,
		},
		{
			name:        "no lowercase",
			password:    "SECUREPASS123",
			expectedErr: ErrPasswordNoLowercase,
		},
		{
			name:        "no number",
			password:    "SecurePassword",
			expectedErr: ErrPasswordNoNumber,
		},
		{
			name:        "exactly 8 characters",
			password:    "Secure12",
			expectedErr: nil,
		},
		{
			name:        "with special characters",
			password:    "Secure@Pass123!",
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePasswordStrength(tt.password)
			if err != tt.expectedErr {
				t.Errorf("expected error %v, got %v", tt.expectedErr, err)
			}
		})
	}
}
