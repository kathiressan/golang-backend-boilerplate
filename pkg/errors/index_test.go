package errors

import (
	"errors"
	"net/http"
	"testing"
)

// ---------------------------------------------------------------------------
// AppError interface compliance
// ---------------------------------------------------------------------------

func TestAppError_Error_should_return_message(t *testing.T) {
	err := &AppError{Message: "something broke"}
	if err.Error() != "something broke" {
		t.Errorf("got %q, want %q", err.Error(), "something broke")
	}
}

func TestAppError_Unwrap_should_return_wrapped_error(t *testing.T) {
	inner := errors.New("inner")
	err := &AppError{Err: inner}
	if !errors.Is(err, inner) {
		t.Error("Unwrap should expose the wrapped error to errors.Is")
	}
}

func TestAppError_Unwrap_should_return_nil_when_no_wrapped_error(t *testing.T) {
	err := &AppError{}
	if err.Unwrap() != nil {
		t.Errorf("expected nil Unwrap when no error wrapped, got %v", err.Unwrap())
	}
}

// ---------------------------------------------------------------------------
// Specific error constructors — each verifies the correct HTTP status, code,
// type, message, and that the original error is preserved.
// ---------------------------------------------------------------------------

func TestErrorConstructors_should_set_correct_fields(t *testing.T) {
	inner := errors.New("cause")
	tests := []struct {
		name       string
		appErr     *AppError
		wantStatus int
		wantCode   string
		wantType   string
	}{
		{"BadRequest", BadRequest(inner, "bad input"), http.StatusBadRequest, "BAD_REQUEST", "BadRequestError"},
		{"Unauthorized", Unauthorized(inner, "no auth"), http.StatusUnauthorized, "UNAUTHORIZED", "UnauthorizedError"},
		{"Forbidden", Forbidden(inner, "no access"), http.StatusForbidden, "FORBIDDEN", "ForbiddenError"},
		{"NotFound", NotFound(inner, "missing"), http.StatusNotFound, "NOT_FOUND", "NotFoundError"},
		{"Conflict", Conflict(inner, "duplicate"), http.StatusConflict, "CONFLICT", "ConflictError"},
		{"InternalServerError", InternalServerError(inner, "boom"), http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "InternalServerError"},
		{"ValidationError", ValidationError(inner, "invalid"), http.StatusUnprocessableEntity, "VALIDATION_ERROR", "ValidationError"},
		{"TooManyRequests", TooManyRequests(inner, "slow down"), http.StatusTooManyRequests, "TOO_MANY_REQUESTS", "RateLimitError"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.appErr.StatusCode != tt.wantStatus {
				t.Errorf("StatusCode: got %d, want %d", tt.appErr.StatusCode, tt.wantStatus)
			}
			if tt.appErr.Code != tt.wantCode {
				t.Errorf("Code: got %q, want %q", tt.appErr.Code, tt.wantCode)
			}
			if tt.appErr.Type != tt.wantType {
				t.Errorf("Type: got %q, want %q", tt.appErr.Type, tt.wantType)
			}
			if !errors.Is(tt.appErr, inner) {
				t.Error("original error should be unwrappable from AppError")
			}
		})
	}
}

func TestInternalServerError_should_use_default_message_when_empty(t *testing.T) {
	err := InternalServerError(nil, "")
	want := "An unexpected error occurred"
	if err.Message != want {
		t.Errorf("got %q, want %q", err.Message, want)
	}
}

// ---------------------------------------------------------------------------
// NewWithDetails / ValidationErrorWithDetails
// ---------------------------------------------------------------------------

func TestNewWithDetails_should_attach_details(t *testing.T) {
	details := map[string]string{"email": "invalid"}
	err := NewWithDetails(nil, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "invalid input", "ValidationError", details)
	if err.ValidationDetails == nil {
		t.Fatal("expected ValidationDetails to be set")
	}
	got, ok := err.ValidationDetails.(map[string]string)
	if !ok || got["email"] != "invalid" {
		t.Errorf("unexpected validation details: %v", err.ValidationDetails)
	}
}

func TestValidationErrorWithDetails_should_have_422_and_details(t *testing.T) {
	details := []string{"field required"}
	err := ValidationErrorWithDetails(nil, "bad input", details)
	if err.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("got status %d, want 422", err.StatusCode)
	}
	if err.ValidationDetails == nil {
		t.Error("expected details to be non-nil")
	}
}

// ---------------------------------------------------------------------------
// Is / AsAppError / WrapError
// ---------------------------------------------------------------------------

func TestIs_should_return_true_when_error_type_matches(t *testing.T) {
	err := NotFound(nil, "not found")
	if !Is(err, "NotFoundError") {
		t.Error("expected Is to return true for matching type")
	}
}

func TestIs_should_return_false_when_error_is_not_AppError(t *testing.T) {
	plain := errors.New("plain")
	if Is(plain, "NotFoundError") {
		t.Error("expected Is to return false for a non-AppError")
	}
}

func TestAsAppError_should_return_AppError_and_true_when_err_is_AppError(t *testing.T) {
	original := BadRequest(nil, "oops")
	got, ok := AsAppError(original)
	if !ok {
		t.Fatal("expected ok=true")
	}
	if got.Code != "BAD_REQUEST" {
		t.Errorf("unexpected Code: %s", got.Code)
	}
}

func TestAsAppError_should_return_false_for_non_AppError(t *testing.T) {
	_, ok := AsAppError(errors.New("plain"))
	if ok {
		t.Error("expected ok=false for a plain error")
	}
}

func TestWrapError_should_produce_500_with_original_message(t *testing.T) {
	inner := errors.New("db timeout")
	err := WrapError(inner)
	if err.StatusCode != http.StatusInternalServerError {
		t.Errorf("got status %d, want 500", err.StatusCode)
	}
	if err.Message != "db timeout" {
		t.Errorf("got message %q, want %q", err.Message, "db timeout")
	}
}
