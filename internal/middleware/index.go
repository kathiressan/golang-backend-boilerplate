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
}