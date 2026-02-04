// Package pathHelper provides utilities for parsing and handling API paths
// This package helps standardize path handling across the application
package pathHelper

import (
	"slices"
	"strings"
)

// Platforms represents valid platforms in the API
// These are the supported platforms that can make requests to the API
var Platforms = []string{"ovmsa", "web", "mobile"}

// ValidActions represents valid HTTP actions/methods
// These are the standard HTTP methods supported by the API
var ValidActions = []string{"GET", "POST", "PUT", "DELETE", "PATCH"}

// ActionMap maps action aliases to standard HTTP actions
// This allows for more intuitive action names in the API
var ActionMap = map[string]string{
	"list":   "GET",    // List resources
	"create": "POST",   // Create a new resource
	"update": "PUT",    // Update an existing resource
	"delete": "DELETE", // Delete a resource
}

// ParsedPath represents the parsed structure of an API path
// This struct contains all the components of a parsed API path
// Fields:
// - FullPath: The complete original path
// - APIVersion: The version of the API being used
// - Platform: The platform making the request (ovmsa, web, mobile)
// - MainRoute: The main route segment of the path
// - SubRoutes: Detailed information about sub-routes
// - FullSubRoute: The complete sub-route path
// - SubRouteWithoutAction: The sub-route without the action
// - SubRoute: The current sub-route being accessed
// - Action: The HTTP action being performed
// - Error: Any error that occurred during parsing
type ParsedPath struct {
	FullPath              string
	APIVersion            string
	Platform              string
	MainRoute             string
	SubRoutes             SubRoutes
	FullSubRoute          string
	SubRouteWithoutAction string
	SubRoute              string
	Action                string
	Error                 error
}

// SubRoutes represents the structure of sub-routes in a path
// This struct provides easy access to different parts of the sub-route
// Fields:
// - Count: Total number of sub-route segments
// - All: All sub-route segments as a slice
// - First: The first sub-route segment
// - Second: The second sub-route segment
// - Third: The third sub-route segment
// - Last: The last sub-route segment
type SubRoutes struct {
	Count  int
	All    []string
	First  string
	Second string
	Third  string
	Last   string
}

// ParseRequestPath parses a request path into its components
// This function:
// 1. Normalizes the path by removing trailing slashes
// 2. Identifies the platform (ovmsa, web, mobile)
// 3. Extracts the main route and sub-routes
// 4. Determines the API version and action
// 5. Returns a ParsedPath struct with all components
// Parameters:
// - path: The request path to parse
// Returns:
// - ParsedPath: A struct containing all parsed path components
func ParseRequestPath(path string) ParsedPath {
	// Remove trailing slash if present
	path = strings.TrimSuffix(path, "/")

	pathParts := strings.Split(path, "/")
	nParts := len(pathParts)

	// Find the starting position
	start := 2
	if pathParts[0] == "" || pathParts[0] == "/" {
		start = 3
	}

	// Initialize parsed path
	parsed := ParsedPath{
		FullPath: path,
	}

	// Get platform
	platform := "ovmsa"
	if start+1 < nParts {
		platform = pathParts[start+1]
		if !slices.Contains(Platforms, platform) {
			platform = "ovmsa"
		}
	}

	// Get main route
	mainRoute := ""
	if platform == "ovmsa" {
		if start+1 < nParts {
			mainRoute = pathParts[start+1]
		}
	} else {
		if start+2 < nParts {
			mainRoute = pathParts[start+2]
		}
		start++
	}

	// Find API position
	apiPosition := -1
	for i, part := range pathParts {
		if part == "api" {
			apiPosition = i
			break
		}
	}

	// Get sub-routes
	var subRoutes []string
	if apiPosition != -1 && apiPosition+2 < nParts {
		fullSubRoute := pathParts[apiPosition+2:]
		subRoute := ""
		if start+2 < nParts {
			subRoute = pathParts[start+2]
			subRoutes = append(subRoutes, subRoute)
		}

		for i := start + 3; i < nParts; i++ {
			subRoute += "/" + pathParts[i]
			subRoutes = append(subRoutes, pathParts[i])
		}

		// Get action
		action := ""
		if len(subRoutes) > 0 {
			action = subRoutes[len(subRoutes)-1]
			if mappedAction, exists := ActionMap[action]; exists {
				action = mappedAction
			}
			if !slices.Contains(ValidActions, action) {
				action = ""
			}
		}

		// Build sub-routes structure
		parsed.SubRoutes = SubRoutes{
			Count:  len(subRoutes),
			All:    subRoutes,
			First:  getOrDefault(subRoutes, 0, ""),
			Second: getOrDefault(subRoutes, 1, ""),
			Third:  getOrDefault(subRoutes, 2, ""),
			Last:   getOrDefault(subRoutes, len(subRoutes)-1, ""),
		}

		parsed.FullSubRoute = strings.Join(fullSubRoute, "/")
		if action != "" {
			parsed.SubRouteWithoutAction = strings.Join(subRoutes[:len(subRoutes)-1], "/")
		} else {
			parsed.SubRouteWithoutAction = strings.Join(subRoutes, "/")
		}
		parsed.SubRoute = strings.Join(subRoutes, "/")
		parsed.Action = strings.ToUpper(action)
	}

	parsed.APIVersion = getOrDefault(pathParts, start, "")
	parsed.Platform = platform
	parsed.MainRoute = mainRoute

	return parsed
}

// Helper function to safely get an element from a slice
func getOrDefault(slice []string, index int, defaultValue string) string {
	if index >= 0 && index < len(slice) {
		return slice[index]
	}
	return defaultValue
}
