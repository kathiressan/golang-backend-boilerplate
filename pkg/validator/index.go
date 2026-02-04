package validatorHelper

import "github.com/go-playground/validator/v10"

// payloadValidator is a singleton instance of the validator
// This is used to validate struct fields using tags
var payloadValidator *validator.Validate

// InitValidator initializes the global validator instance
// This function:
// 1. Creates a new validator instance
// 2. Enables required struct validation
// 3. Should be called during application startup
func InitValidator() {
	payloadValidator = validator.New(validator.WithRequiredStructEnabled())
}

// GetValidator returns the global validator instance
// This allows other packages to use the same validator instance
// for consistent validation across the application
func GetValidator() *validator.Validate {
	return payloadValidator
}
