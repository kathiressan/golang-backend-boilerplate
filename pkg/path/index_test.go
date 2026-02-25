package pathHelper

import (
	"testing"
)

// parser is the shared instance used by all tests.
var parser = NewParser("ovmsa")

// Trace of ParseRequestPath("/api/v1.0/ovmsa/users") with platformName="ovmsa":
//   pathParts = ["", "api", "v1.0", "ovmsa", "users"], nParts=5, start=3
//   platform  = pathParts[start+1=4] = "users" → not in platforms → fallback "ovmsa"
//   mainRoute = pathParts[start+1=4] = "users"  (because platform == platformName)
//   APIVersion= pathParts[start=3]   = "ovmsa"  (the platform segment, not "v1.0")
//   apiPosition=1 → FullSubRoute = pathParts[3:] = "ovmsa/users"

func TestParseRequestPath_should_set_platform_and_main_route(t *testing.T) {
	p := parser.ParseRequestPath("/api/v1.0/ovmsa/users")

	if p.Platform != "ovmsa" {
		t.Errorf("Platform: got %q, want %q", p.Platform, "ovmsa")
	}
	if p.MainRoute != "users" {
		t.Errorf("MainRoute: got %q, want %q", p.MainRoute, "users")
	}
}

func TestParseRequestPath_should_strip_trailing_slash(t *testing.T) {
	withSlash := parser.ParseRequestPath("/api/v1.0/ovmsa/users/")
	without := parser.ParseRequestPath("/api/v1.0/ovmsa/users")

	if withSlash.FullPath != "/api/v1.0/ovmsa/users" {
		t.Errorf("trailing slash not stripped: FullPath=%q", withSlash.FullPath)
	}
	if withSlash.MainRoute != without.MainRoute {
		t.Errorf("MainRoute differs with/without trailing slash: %q vs %q",
			withSlash.MainRoute, without.MainRoute)
	}
}

func TestParseRequestPath_should_map_action_aliases_to_HTTP_methods(t *testing.T) {
	tests := []struct {
		alias      string
		wantAction string
	}{
		{"list", "GET"},
		{"create", "POST"},
		{"update", "PUT"},
		{"delete", "DELETE"},
	}

	for _, tt := range tests {
		t.Run(tt.alias, func(t *testing.T) {
			p := parser.ParseRequestPath("/api/v1.0/ovmsa/users/" + tt.alias)
			if p.Action != tt.wantAction {
				t.Errorf("Action: got %q, want %q", p.Action, tt.wantAction)
			}
		})
	}
}

func TestParseRequestPath_should_set_empty_action_when_last_segment_is_not_valid(t *testing.T) {
	p := parser.ParseRequestPath("/api/v1.0/ovmsa/users/profile")
	if p.Action != "" {
		t.Errorf("expected empty Action for non-action segment, got %q", p.Action)
	}
}

func TestParseRequestPath_should_use_valid_HTTP_verb_as_action_directly(t *testing.T) {
	p := parser.ParseRequestPath("/api/v1.0/ovmsa/users/GET")
	if p.Action != "GET" {
		t.Errorf("expected Action=GET, got %q", p.Action)
	}
}

func TestParseRequestPath_should_fall_back_to_platform_name_for_unknown_platform(t *testing.T) {
	p := parser.ParseRequestPath("/api/v1.0/unknown-platform/users")
	if p.Platform != "ovmsa" {
		t.Errorf("Platform: got %q, want fallback %q", p.Platform, "ovmsa")
	}
}

func TestParseRequestPath_should_build_sub_route_fields_correctly(t *testing.T) {
	// "/api/v1.0/ovmsa/users/settings/GET"
	// apiPosition=1, start=3
	// subRoutes at start+2=5 onwards: ["settings", "GET"]
	p := parser.ParseRequestPath("/api/v1.0/ovmsa/users/settings/GET")

	if p.SubRoutes.First != "settings" {
		t.Errorf("SubRoutes.First: got %q, want %q", p.SubRoutes.First, "settings")
	}
	if p.Action != "GET" {
		t.Errorf("Action: got %q, want %q", p.Action, "GET")
	}
	if p.SubRouteWithoutAction != "settings" {
		t.Errorf("SubRouteWithoutAction: got %q, want %q", p.SubRouteWithoutAction, "settings")
	}
}

func TestParseRequestPath_should_preserve_full_path(t *testing.T) {
	path := "/api/v1.0/ovmsa/users"
	p := parser.ParseRequestPath(path)
	if p.FullPath != path {
		t.Errorf("FullPath: got %q, want %q", p.FullPath, path)
	}
}

// ---------------------------------------------------------------------------
// getOrDefault (internal helper)
// ---------------------------------------------------------------------------

func TestGetOrDefault_should_return_value_at_valid_index(t *testing.T) {
	s := []string{"a", "b", "c"}
	if got := getOrDefault(s, 1, "x"); got != "b" {
		t.Errorf("got %q, want %q", got, "b")
	}
}

func TestGetOrDefault_should_return_default_for_out_of_bounds_index(t *testing.T) {
	s := []string{"a"}
	if got := getOrDefault(s, 5, "default"); got != "default" {
		t.Errorf("got %q, want %q", got, "default")
	}
}

func TestGetOrDefault_should_return_default_for_negative_index(t *testing.T) {
	s := []string{"a", "b"}
	if got := getOrDefault(s, -1, "fallback"); got != "fallback" {
		t.Errorf("got %q, want %q", got, "fallback")
	}
}
