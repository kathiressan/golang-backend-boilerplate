package middleware

import (
	"os"
	"os/signal"
	"ovmsa-be/pkg/logger"
	"syscall"
	"time"

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

	// Add error handler middleware for 404 and 500 errors
	// This provides consistent error responses
	router.Use(Error404n500Handler())

	// Add token handler middleware for external API authentication
	// This validates tokens from external services
	router.Use(ExtRequesterHandler())

	// Add path handler middleware for route parsing
	// This normalizes and parses request paths
	router.Use(PathHandler())

	// Set up graceful shutdown handling
	// This ensures clean shutdown when the server is stopped
	setupGracefulShutdown()
}

// setupGracefulShutdown configures the application to handle shutdown signals gracefully
// This function:
// 1. Listens for shutdown signals (SIGINT, SIGTERM, SIGQUIT)
// 2. Waits for ongoing requests to complete
// 3. Flushes any buffered logs
// 4. Exits cleanly
func setupGracefulShutdown() {
	// Create a channel to receive OS signals
	c := make(chan os.Signal, 1)

	// Register for interrupt and termination signals
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

	// Start a goroutine to handle shutdown
	go func() {
		// Wait for a shutdown signal
		<-c
		logger.Info("Shutdown signal received, beginning graceful shutdown")

		// Allow ongoing requests to finish
		// This gives time for existing requests to complete
		time.Sleep(5 * time.Second)

		// Flush any buffered logs to ensure all logs are written
		logger.Sync()

		// Log completion and exit
		logger.Info("Graceful shutdown completed")
		os.Exit(0)
	}()
}
