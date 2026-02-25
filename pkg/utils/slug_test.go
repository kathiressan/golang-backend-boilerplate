package utils

import (
	"testing"
)

func TestSlugify(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"should_convert_normal_string_to_slug", "Hello World", "hello-world"},
		{"should_remove_non_alphanumeric_chars", "Foo@Bar!Test", "foo-bar-test"},
		{"should_handle_leading_dashes", "!@#Hello", "hello"},
		{"should_handle_trailing_dashes", "Hello!@#", "hello"},
		{"should_collapse_multiple_dashes", "A---B___C", "a-b-c"},
		{"should_return_empty_for_empty_string", "", ""},
		{"should_return_empty_for_special_chars_only", "@#$%", ""},
		{"should_handle_numbers", "Test123", "test123"},
		{"should_preserve_numbers_in_middle", "abc123def", "abc123def"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Slugify(tt.input)
			if result != tt.expected {
				t.Errorf("Slugify(%q) = %q; want %q", tt.input, result, tt.expected)
			}
		})
	}
}
