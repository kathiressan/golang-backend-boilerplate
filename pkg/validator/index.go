package validatorHelper

import (
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

// payloadValidator is a singleton instance of the validator
// This is used to validate struct fields using tags
var payloadValidator *validator.Validate

// InitValidator initializes the global validator instance
// This function:
// 1. Creates a new validator instance
// 2. Enables required struct validation
// 3. Configures common tag name mapping (e.g. uses json tags)
// 4. Should be called during application startup
func InitValidator() {
	payloadValidator = validator.New(validator.WithRequiredStructEnabled())

	// Register function to use JSON tag names instead of struct field names
	payloadValidator.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
}

// GetValidator returns the global validator instance
// This allows other packages to use the same validator instance
// for consistent validation across the application
func GetValidator() *validator.Validate {
	return payloadValidator
}
