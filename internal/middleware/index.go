package middleware

import (
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
)

// SetupMiddleware configures all middleware for the application
func SetupMiddleware(router *gin.Engine) {
	// Generate unique request IDs for each request
	// This helps track requests through the system
	router.Use(requestid.New())

	// Add our custom structured logger
	// This logs request details and response information
	router.Use(LoggerMiddleware())

	// Add recovery middleware to handle panics
	// This prevents the server from crashing on unhandled errors
	router.Use(gin.Recovery())

	// Add CORS headers to allow cross-origin requests
	// This is necessary for web applications making API calls
	router.Use(CORSMiddleware())

	// Add input sanitizer middleware to prevent SQL injection, XSS, and MongoDB injection
	// This ensures that all incoming requests are sanitized
	router.Use(GlobalSanitizer())
}