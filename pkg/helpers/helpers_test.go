package helpers

import (
	"net/http"
	"testing"

	"ovmsa-be/internal/entities"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// ---------------------------------------------------------------------------
// RouteBuilder
// ---------------------------------------------------------------------------

func TestNewRoute_should_default_to_unprotected(t *testing.T) {
	r := NewRoute("/foo", "GET").Build()
	if r.ProtectedBy != entities.UNPROTECTED {
		t.Errorf("got %q, want UNPROTECTED", r.ProtectedBy)
	}
}

func TestRouteBuilder_should_set_path_and_method(t *testing.T) {
	r := NewRoute("/users", "POST").Build()
	if r.Path != "/users" {
		t.Errorf("Path: got %q, want %q", r.Path, "/users")
	}
	if r.Method != "POST" {
		t.Errorf("Method: got %q, want %q", r.Method, "POST")
	}
}

func TestRouteBuilder_ProtectedByJWT_should_set_jwt_strategy(t *testing.T) {
	r := NewRoute("/secure", "GET").ProtectedByJWT().Build()
	if r.ProtectedBy != entities.JWT {
		t.Errorf("got %q, want JWT", r.ProtectedBy)
	}
}

func TestRouteBuilder_ProtectedByRBAC_should_set_roles(t *testing.T) {
	r := NewRoute("/admin", "GET").ProtectedByRBAC("admin", "superuser").Build()
	if r.ProtectedBy != entities.RBAC_AUTH {
		t.Errorf("ProtectedBy: got %q, want RBAC_AUTH", r.ProtectedBy)
	}
	if r.Permissions == nil {
		t.Fatal("expected Permissions to be set")
	}
	if len(r.Permissions.AllowedRoles) != 2 || r.Permissions.AllowedRoles[0] != "admin" {
		t.Errorf("unexpected AllowedRoles: %v", r.Permissions.AllowedRoles)
	}
}

func TestRouteBuilder_ProtectedByABAC_should_set_attributes(t *testing.T) {
	attrs := map[string]any{"department": "engineering"}
	r := NewRoute("/resource", "GET").ProtectedByABAC(attrs).Build()
	if r.ProtectedBy != entities.ABAC_AUTH {
		t.Errorf("ProtectedBy: got %q, want ABAC_AUTH", r.ProtectedBy)
	}
	if r.Attributes == nil {
		t.Fatal("expected Attributes to be set")
	}
	if r.Attributes.RequiredAttributes["department"] != "engineering" {
		t.Errorf("unexpected attributes: %v", r.Attributes.RequiredAttributes)
	}
}

func TestRouteBuilder_ProtectedByCombined_should_set_both_rbac_and_abac(t *testing.T) {
	attrs := map[string]any{"region": "us"}
	r := NewRoute("/combined", "GET").ProtectedByCombined([]string{"manager"}, attrs).Build()
	if r.ProtectedBy != entities.COMBINED_AUTH {
		t.Errorf("ProtectedBy: got %q, want COMBINED_AUTH", r.ProtectedBy)
	}
	if r.Permissions == nil || r.Attributes == nil {
		t.Error("expected both Permissions and Attributes to be set")
	}
}

func TestRouteBuilder_WithSuccessCode_should_override_default(t *testing.T) {
	r := POST("/items").WithSuccessCode(http.StatusCreated).Build()
	if r.SuccessCode != http.StatusCreated {
		t.Errorf("SuccessCode: got %d, want 201", r.SuccessCode)
	}
}

func TestRouteBuilder_WithSchema_should_store_schema(t *testing.T) {
	type Req struct{ Name string }
	r := POST("/items").WithSchema(Req{}).Build()
	if r.Schema == nil {
		t.Error("expected Schema to be non-nil")
	}
}

func TestRouteBuilder_Unprotected_should_reset_to_unprotected_after_jwt(t *testing.T) {
	r := NewRoute("/open", "GET").ProtectedByJWT().Unprotected().Build()
	if r.ProtectedBy != entities.UNPROTECTED {
		t.Errorf("got %q, want UNPROTECTED after chaining Unprotected()", r.ProtectedBy)
	}
}

// ---------------------------------------------------------------------------
// HTTP method convenience constructors
// ---------------------------------------------------------------------------

func TestHTTPConvenienceBuilders_should_set_correct_methods(t *testing.T) {
	tests := []struct {
		name    string
		builder *RouteBuilder
		want    string
	}{
		{"GET", GET("/a"), "GET"},
		{"POST", POST("/a"), "POST"},
		{"PUT", PUT("/a"), "PUT"},
		{"DELETE", DELETE("/a"), "DELETE"},
		{"PATCH", PATCH("/a"), "PATCH"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := tt.builder.Build()
			if r.Method != tt.want {
				t.Errorf("Method: got %q, want %q", r.Method, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// ErrorHandlerChain (helpers/errors.go)
// ---------------------------------------------------------------------------

func TestErrorHandlerChain_Handle_should_return_false_when_chain_is_empty(t *testing.T) {
	chain := NewErrorHandlerChain()
	ctx, _ := gin.CreateTestContext(nil)
	if chain.Handle(ctx, ErrIdentityNotFound) {
		t.Error("expected false for empty chain")
	}
}

func TestErrorHandlerChain_Add_should_wire_handlers_in_order(t *testing.T) {
	// Two handlers: first matches ErrIdentityNotFound, second matches ErrInvalidIdentityType.
	// Passing ErrInvalidIdentityType should reach the second handler.
	var reached string

	h1 := NewSpecificErrorHandler(ErrIdentityNotFound, 401, "not found")
	h2 := &captureHandler{onHandle: func() { reached = "h2" }}

	chain := NewErrorHandlerChain().Add(h1).Add(h2)
	ctx, w := gin.CreateTestContext(nil)
	_ = w
	chain.Handle(ctx, ErrInvalidIdentityType)

	if reached != "h2" {
		t.Errorf("expected h2 to handle the error, but reached=%q", reached)
	}
}

func TestHandleServiceError_should_return_false_nil_when_err_is_nil(t *testing.T) {
	ctx, _ := gin.CreateTestContext(nil)
	handled, err := HandleServiceError(ctx, nil, NewErrorHandlerChain())
	if handled || err != nil {
		t.Errorf("expected (false, nil), got (%v, %v)", handled, err)
	}
}

func TestHandleServiceError_should_return_unhandled_error_when_chain_does_not_match(t *testing.T) {
	chain := NewErrorHandlerChain() // empty — nothing matches
	ctx, _ := gin.CreateTestContext(nil)
	handled, returnedErr := HandleServiceError(ctx, ErrIdentityNotFound, chain)
	if handled {
		t.Error("expected handled=false for empty chain")
	}
	if returnedErr != ErrIdentityNotFound {
		t.Errorf("expected original error back, got %v", returnedErr)
	}
}

// captureHandler is a test double that records whether it was called.
type captureHandler struct {
	BaseErrorHandler
	onHandle func()
}

func (h *captureHandler) Handle(_ *gin.Context, _ error) bool {
	if h.onHandle != nil {
		h.onHandle()
	}
	return true
}
