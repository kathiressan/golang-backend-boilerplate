package middleware

import (
	config "ovmsa-be/configs"
	"ovmsa-be/pkg/logger"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORSMiddleware handles CORS headers with proper configuration
// This middleware:
// 1. Configures Cross-Origin Resource Sharing (CORS) headers
// 2. Validates allowed origins from configuration
// 3. Sets security headers for enhanced protection
// 4. Handles preflight requests
// 5. Logs CORS configuration at startup
func CORSMiddleware() gin.HandlerFunc {
	cfg := config.GetConfig()
	allowedOrigins := cfg.AllowedOrigins
	allowedOriginsStr := strings.Join(allowedOrigins, ",")

	logger.Info("CORS middleware initialized",
		"environment", cfg.Environment,
		"allowedOrigins", allowedOriginsStr,
	)

	return func(c *gin.Context) {

		// Get the origin of the request
		// This is the domain making the request
		origin := c.Request.Header.Get("Origin")

		// Check if the request origin is allowed
		if origin != "" {
			allowed := false
			for _, allowedOrigin := range allowedOrigins {
				// Allow all origins if '*' is specified
				// Otherwise, check for exact match
				if allowedOrigin == "*" || allowedOrigin == origin {
					allowed = true
					break
				}
			}

			// Set the allowed origin header if the origin is permitted
			if allowed {
				c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
				// Allow-Credentials is only meaningful (and safe to send) when a
				// specific, permitted origin is reflected — not for unmatched origins.
				c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
			}
		}

		// Set remaining CORS headers (non-origin-specific)
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, Accept, Origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")

		// Add security headers to protect against common web vulnerabilities
		c.Writer.Header().Set("X-Content-Type-Options", "nosniff")                                // Prevent MIME type sniffing
		c.Writer.Header().Set("X-XSS-Protection", "1; mode=block")                                // Enable XSS protection
		c.Writer.Header().Set("X-Frame-Options", "DENY")                                          // Prevent clickjacking
		c.Writer.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains") // Enforce HTTPS

		// Handle preflight requests (OPTIONS)
		// These are sent by browsers before certain types of requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204) // No content response for preflight
			return
		}

		c.Next()
	}
}
