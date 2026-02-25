// Package password provides secure password hashing and validation utilities
// This package uses bcrypt for password hashing with best practices
package password

import (
	"errors"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

const (
	// MinPasswordLength is the minimum required password length
	MinPasswordLength = 8
	// BcryptCost is the cost factor for bcrypt hashing (12 is recommended for production)
	BcryptCost = 12
)

var (
	// ErrPasswordTooShort is returned when password is shorter than minimum length
	ErrPasswordTooShort = errors.New("password must be at least 8 characters long")
	// ErrPasswordNoUppercase is returned when password has no uppercase letter
	ErrPasswordNoUppercase = errors.New("password must contain at least one uppercase letter")
	// ErrPasswordNoLowercase is returned when password has no lowercase letter
	ErrPasswordNoLowercase = errors.New("password must contain at least one lowercase letter")
	// ErrPasswordNoNumber is returned when password has no number
	ErrPasswordNoNumber = errors.New("password must contain at least one number")
	// ErrInvalidPassword is returned when password verification fails
	ErrInvalidPassword = errors.New("invalid password")
)

// HashPassword generates a bcrypt hash from a plain text password.
// The bcrypt algorithm automatically includes a salt in the hash.
// Returns the hashed password or an error if hashing fails.
func HashPassword(password string) (string, error) {
	// Validate password strength before hashing
	if err := ValidatePasswordStrength(password); err != nil {
		return "", err
	}

	// Generate bcrypt hash
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), BcryptCost)
	if err != nil {
		return "", err
	}

	return string(hashedBytes), nil
}

// VerifyPassword compares a plain text password with a bcrypt hash.
// Returns nil if the password matches, or an error if it doesn't match or verification fails.
func VerifyPassword(password, hash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return ErrInvalidPassword
		}
		return err
	}
	return nil
}

// ValidatePasswordStrength checks if a password meets security requirements:
// - Minimum 8 characters
// - At least one uppercase letter
// - At least one lowercase letter
// - At least one number
// Returns nil if password is strong enough, or a descriptive error otherwise.
func ValidatePasswordStrength(password string) error {
	if len(password) < MinPasswordLength {
		return ErrPasswordTooShort
	}

	var (
		hasUpper  bool
		hasLower  bool
		hasNumber bool
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		}
	}

	if !hasUpper {
		return ErrPasswordNoUppercase
	}
	if !hasLower {
		return ErrPasswordNoLowercase
	}
	if !hasNumber {
		return ErrPasswordNoNumber
	}

	return nil
}
