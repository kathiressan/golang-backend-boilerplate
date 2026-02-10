package utils

import (
	"regexp"
	"strings"
)

var (
	regexpNonAlphaNumeric = regexp.MustCompile(`[^a-z0-9]+`)
	regexpDashes          = regexp.MustCompile(`^-+|-+$`)
)

// Slugify converts a string into a URL-friendly slug.
// It lowercase the string, replaces non-alphanumeric characters with dashes,
// and removes leading/trailing dashes.
func Slugify(s string) string {
	s = strings.ToLower(s)
	s = regexpNonAlphaNumeric.ReplaceAllString(s, "-")
	s = regexpDashes.ReplaceAllString(s, "")
	return s
}
