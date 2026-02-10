package validatorHelper

import (
	"fmt"
	"reflect"

	"github.com/go-playground/validator/v10"
)

// ValidationFormatter defines the interface for formatting validation errors
type ValidationFormatter interface {
	Format(err validator.FieldError) string
}

// FormatterRegistry manages validation error formatters
type FormatterRegistry struct {
	formatters map[string]ValidationFormatter
}

// NewFormatterRegistry creates a new formatter registry with default formatters
func NewFormatterRegistry() *FormatterRegistry {
	registry := &FormatterRegistry{
		formatters: make(map[string]ValidationFormatter),
	}

	// Register default formatters
	registry.Register("required", &RequiredFormatter{})
	registry.Register("email", &EmailFormatter{})
	registry.Register("min", &MinFormatter{})
	registry.Register("max", &MaxFormatter{})
	registry.Register("url", &URLFormatter{})
	registry.Register("oneof", &OneOfFormatter{})
	registry.Register("len", &LenFormatter{})
	registry.Register("alphanum", &AlphanumFormatter{})
	registry.Register("numeric", &NumericFormatter{})

	return registry
}

// Register adds a formatter for a specific validation tag
func (r *FormatterRegistry) Register(tag string, formatter ValidationFormatter) {
	r.formatters[tag] = formatter
}

// Format formats a validation error using the appropriate formatter
func (r *FormatterRegistry) Format(err validator.FieldError) string {
	if formatter, exists := r.formatters[err.Tag()]; exists {
		return formatter.Format(err)
	}
	// Default formatter for unknown tags
	return fmt.Sprintf("Invalid value (failed %s)", err.Tag())
}

// FormatAll formats all validation errors into a map
func (r *FormatterRegistry) FormatAll(err error) map[string]string {
	validationErrors := make(map[string]string)

	if validationErrs, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrs {
			fieldName := e.Field()
			validationErrors[fieldName] = r.Format(e)
		}
	} else {
		validationErrors["general"] = err.Error()
	}

	return validationErrors
}

// Concrete formatters

// RequiredFormatter formats required field errors
type RequiredFormatter struct{}

func (f *RequiredFormatter) Format(err validator.FieldError) string {
	return "This field is required"
}

// EmailFormatter formats email validation errors
type EmailFormatter struct{}

func (f *EmailFormatter) Format(err validator.FieldError) string {
	return "Must be a valid email address"
}

// MinFormatter formats minimum value/length errors
type MinFormatter struct{}

func (f *MinFormatter) Format(err validator.FieldError) string {
	if err.Kind() == reflect.String {
		return fmt.Sprintf("Must be at least %s characters", err.Param())
	}
	return fmt.Sprintf("Value must be at least %s", err.Param())
}

// MaxFormatter formats maximum value/length errors
type MaxFormatter struct{}

func (f *MaxFormatter) Format(err validator.FieldError) string {
	if err.Kind() == reflect.String {
		return fmt.Sprintf("Must be no more than %s characters", err.Param())
	}
	return fmt.Sprintf("Value must be no more than %s", err.Param())
}

// URLFormatter formats URL validation errors
type URLFormatter struct{}

func (f *URLFormatter) Format(err validator.FieldError) string {
	return "Must be a valid URL"
}

// OneOfFormatter formats oneof validation errors
type OneOfFormatter struct{}

func (f *OneOfFormatter) Format(err validator.FieldError) string {
	return fmt.Sprintf("Must be one of [%s]", err.Param())
}

// LenFormatter formats length validation errors
type LenFormatter struct{}

func (f *LenFormatter) Format(err validator.FieldError) string {
	return fmt.Sprintf("Length must be exactly %s", err.Param())
}

// AlphanumFormatter formats alphanumeric validation errors
type AlphanumFormatter struct{}

func (f *AlphanumFormatter) Format(err validator.FieldError) string {
	return "Must contain only letters and numbers"
}

// NumericFormatter formats numeric validation errors
type NumericFormatter struct{}

func (f *NumericFormatter) Format(err validator.FieldError) string {
	return "Must be a valid number"
}

// Global formatter registry instance
var globalRegistry *FormatterRegistry

// GetFormatterRegistry returns the global formatter registry
func GetFormatterRegistry() *FormatterRegistry {
	if globalRegistry == nil {
		globalRegistry = NewFormatterRegistry()
	}
	return globalRegistry
}
