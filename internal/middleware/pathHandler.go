package middleware

import (
	pathHelper "ovmsa-be/pkg/path"

	"github.com/gin-gonic/gin"
)

// PathHandler is a middleware that parses the request path and adds it to the context
// This middleware:
// 1. Takes the raw URL path from the request
// 2. Uses the path helper to parse and normalize the path
// 3. Stores the parsed path in the request context for use by subsequent handlers
// The parsed path can be used for:
// - Route matching
// - Access control
// - Logging
// - Metrics collection
func PathHandler(parser pathHelper.PathParser) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse the request path using the path helper
		// This normalizes the path and handles any special cases
		parsedPath := parser.ParseRequestPath(c.Request.URL.Path)

		// Store the parsed path in the request context
		// This makes it available to all subsequent handlers in the chain
		c.Set("parsedPath", parsedPath)

		// Continue to the next handler in the chain
		c.Next()
	}
}
