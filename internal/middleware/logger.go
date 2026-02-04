package middleware

import (
	"ovmsa-be/pkg/logger"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// LoggerMiddleware logs the request and response using structured logging
// This middleware:
// 1. Generates or retrieves a unique request ID for tracking
// 2. Records request start time for latency calculation
// 3. Captures request details (method, path, client IP, etc.)
// 4. Logs request completion with structured data
// 5. Separately logs any errors that occurred during request processing
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Record the start time of the request for latency calculation
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		// Get or generate a unique request ID
		// This ID is used to correlate logs for the same request
		requestID := c.GetString("X-Request-ID")
		if requestID == "" {
			// Generate a new UUID if no request ID exists
			requestID = uuid.New().String()
			c.Set("X-Request-ID", requestID)
		}

		// Process the request and let other middleware/handlers run
		c.Next()

		// Calculate the total time taken to process the request
		latency := time.Since(start)

		// Collect additional request information
		clientIP := c.ClientIP()                                       // Client's IP address
		statusCode := c.Writer.Status()                                // HTTP status code
		userAgent := c.Request.UserAgent()                             // Client's browser/application
		errorMessage := c.Errors.ByType(gin.ErrorTypePrivate).String() // Any errors that occurred

		// Log the completed request with structured data
		// This makes it easier to search and analyze logs
		logger.Info("Request completed",
			"requestID", requestID, // Unique identifier for this request
			"clientIP", clientIP, // Client's IP address
			"method", method, // HTTP method (GET, POST, etc.)
			"path", path, // Requested URL path
			"status", statusCode, // HTTP status code
			"latency", latency, // Time taken to process the request
			"userAgent", userAgent, // Client's browser/application
			"errorCount", len(c.Errors), // Number of errors that occurred
		)

		// If there were any errors, log them separately
		// This makes it easier to find and debug issues
		if errorMessage != "" {
			logger.Error("Request errors",
				"requestID", requestID, // Link errors to the specific request
				"errors", errorMessage, // Detailed error information
			)
		}
	}
}
