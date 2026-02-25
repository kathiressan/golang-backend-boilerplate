package validatorHelper

import (
	"reflect"
	"strings"
	"testing"

	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
)

// ---------------------------------------------------------------------------
// FormatterRegistry
// ---------------------------------------------------------------------------

func TestFormatterRegistry_Format_should_use_registered_formatter(t *testing.T) {
	reg := NewFormatterRegistry()
	fe := newFakeFieldError("email", "email", "", reflect.String)
	got := reg.Format(fe)
	if got != "Must be a valid email address" {
		t.Errorf("unexpected message: %q", got)
	}
}

func TestFormatterRegistry_Format_should_fall_back_for_unknown_tag(t *testing.T) {
	reg := NewFormatterRegistry()
	fe := newFakeFieldError("field", "unknowntag", "", reflect.String)
	got := reg.Format(fe)
	if !strings.Contains(got, "unknowntag") {
		t.Errorf("expected fallback message to contain tag name, got: %q", got)
	}
}

func TestFormatterRegistry_Register_should_override_existing_formatter(t *testing.T) {
	reg := NewFormatterRegistry()
	reg.Register("email", &staticFormatter{msg: "custom email msg"})
	fe := newFakeFieldError("email", "email", "", reflect.String)
	if got := reg.Format(fe); got != "custom email msg" {
		t.Errorf("got %q, want %q", got, "custom email msg")
	}
}

func TestFormatterRegistry_FormatAll_should_handle_non_ValidationErrors(t *testing.T) {
	reg := NewFormatterRegistry()
	plain := plainErr("something went wrong")
	result := reg.FormatAll(plain)
	if result["general"] != "something went wrong" {
		t.Errorf("unexpected result: %v", result)
	}
}

// ---------------------------------------------------------------------------
// Concrete formatters (each covers the distinct branch in its Format method)
// ---------------------------------------------------------------------------

func TestMinFormatter_should_use_characters_wording_for_strings(t *testing.T) {
	f := &MinFormatter{}
	fe := newFakeFieldError("name", "min", "3", reflect.String)
	got := f.Format(fe)
	if !strings.Contains(got, "characters") || !strings.Contains(got, "3") {
		t.Errorf("unexpected string-min message: %q", got)
	}
}

func TestMinFormatter_should_use_value_wording_for_numbers(t *testing.T) {
	f := &MinFormatter{}
	fe := newFakeFieldError("age", "min", "18", reflect.Int)
	got := f.Format(fe)
	if !strings.Contains(got, "Value") || !strings.Contains(got, "18") {
		t.Errorf("unexpected numeric-min message: %q", got)
	}
}

func TestMaxFormatter_should_use_characters_wording_for_strings(t *testing.T) {
	f := &MaxFormatter{}
	fe := newFakeFieldError("bio", "max", "500", reflect.String)
	got := f.Format(fe)
	if !strings.Contains(got, "characters") || !strings.Contains(got, "500") {
		t.Errorf("unexpected string-max message: %q", got)
	}
}

func TestMaxFormatter_should_use_value_wording_for_numbers(t *testing.T) {
	f := &MaxFormatter{}
	fe := newFakeFieldError("score", "max", "100", reflect.Int)
	got := f.Format(fe)
	if !strings.Contains(got, "Value") || !strings.Contains(got, "100") {
		t.Errorf("unexpected numeric-max message: %q", got)
	}
}

func TestOneOfFormatter_should_include_param_in_message(t *testing.T) {
	f := &OneOfFormatter{}
	fe := newFakeFieldError("status", "oneof", "active inactive", reflect.String)
	got := f.Format(fe)
	if !strings.Contains(got, "active inactive") {
		t.Errorf("expected param in message, got: %q", got)
	}
}

func TestLenFormatter_should_include_param_in_message(t *testing.T) {
	f := &LenFormatter{}
	fe := newFakeFieldError("code", "len", "6", reflect.String)
	got := f.Format(fe)
	if !strings.Contains(got, "6") {
		t.Errorf("expected length param in message, got: %q", got)
	}
}

// ---------------------------------------------------------------------------
// GetFormatterRegistry global singleton
// ---------------------------------------------------------------------------

func TestGetFormatterRegistry_should_return_same_instance_on_repeated_calls(t *testing.T) {
	a := GetFormatterRegistry()
	b := GetFormatterRegistry()
	if a != b {
		t.Error("expected the same singleton instance on repeated calls")
	}
}

// ---------------------------------------------------------------------------
// Validator (index.go)
// ---------------------------------------------------------------------------

func TestInitValidator_should_make_GetValidator_return_non_nil(t *testing.T) {
	// Reset global state so this test is hermetic.
	payloadValidator = nil
	InitValidator()
	if GetValidator() == nil {
		t.Error("expected a non-nil validator after InitValidator")
	}
}

func TestInitValidator_should_use_json_tag_names_in_errors(t *testing.T) {
	payloadValidator = nil
	InitValidator()

	type Req struct {
		UserEmail string `json:"user_email" validate:"required"`
	}

	err := GetValidator().Struct(Req{})
	if err == nil {
		t.Fatal("expected validation error for missing required field")
	}

	var ve validator.ValidationErrors
	if ok := isValidationErrors(err, &ve); !ok {
		t.Fatalf("unexpected error type: %T", err)
	}

	if ve[0].Field() != "user_email" {
		t.Errorf("expected field name %q via json tag, got %q", "user_email", ve[0].Field())
	}
}

// ---------------------------------------------------------------------------
// Test helpers / doubles
// ---------------------------------------------------------------------------

type fakeFieldError struct {
	field string
	tag   string
	param string
	kind  reflect.Kind
}

func newFakeFieldError(field, tag, param string, kind reflect.Kind) validator.FieldError {
	return &fakeFieldError{field: field, tag: tag, param: param, kind: kind}
}

func (f *fakeFieldError) Tag() string            { return f.tag }
func (f *fakeFieldError) ActualTag() string       { return f.tag }
func (f *fakeFieldError) Namespace() string       { return "" }
func (f *fakeFieldError) StructNamespace() string { return "" }
func (f *fakeFieldError) Field() string           { return f.field }
func (f *fakeFieldError) StructField() string     { return f.field }
func (f *fakeFieldError) Value() interface{}      { return nil }
func (f *fakeFieldError) Param() string           { return f.param }
func (f *fakeFieldError) Kind() reflect.Kind      { return f.kind }
func (f *fakeFieldError) Type() reflect.Type      { return nil }
func (f *fakeFieldError) Translate(_ ut.Translator) string { return "" }
func (f *fakeFieldError) Error() string { return f.tag }

type staticFormatter struct{ msg string }

func (s *staticFormatter) Format(_ validator.FieldError) string { return s.msg }

type plainErr string

func (e plainErr) Error() string { return string(e) }

func isValidationErrors(err error, target *validator.ValidationErrors) bool {
	ve, ok := err.(validator.ValidationErrors)
	if ok {
		*target = ve
	}
	return ok
}
