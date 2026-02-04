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
	return func(c *gin.Context) {
		cfg := config.GetConfig()

		// Get allowed origins from configuration
		// These are the domains that are allowed to make requests to the API
		allowedOrigins := strings.Join(cfg.AllowedOrigins, ",")

		// Log CORS configuration at startup (only on the root path)
		// This helps with debugging and monitoring
		if c.Request.URL.Path == "/" && c.Request.Method == "GET" {
			logger.Info("CORS configuration",
				"environment", cfg.Environment,
				"allowedOrigins", allowedOrigins,
			)
		}

		// Get the origin of the request
		// This is the domain making the request
		origin := c.Request.Header.Get("Origin")

		// Check if the request origin is allowed
		if origin != "" {
			allowed := false
			for _, allowedOrigin := range cfg.AllowedOrigins {
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
			}
		}

		// Set CORS headers
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")                                                                                                                    // Allow credentials (cookies, etc.)
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, Accept, Origin, Cache-Control, X-Requested-With") // Allowed request headers
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")                                                                                      // Allowed HTTP methods
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")                                                                                                                             // Cache preflight results for 24 hours

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
