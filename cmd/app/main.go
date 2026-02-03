package main

import (
	"fmt"
	"net/http"
	config "ovmsa-be/configs"
	"ovmsa-be/pkg/logger"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.GetConfig()

	logger.Initialize(cfg.Environment)

	// Ensure logs are flushed on shutdown
	// defer statements run when the function exits
	defer logger.Sync()

	logger.Debug("Starting application", 
		"environment", cfg.Environment,
		"port", cfg.Port,
		"ginMode", cfg.GinMode,
	)

	// Set Gin mode based on configuration
	// Different modes have different behaviors:
	// - debug: verbose logging, error stack traces
	// - release: optimized for production, minimal logging
	// - test: for testing purpose
	gin.SetMode(cfg.GinMode)

	// Create a new router without default middleware
	// We'll add our own middleware stack for better control
	router := gin.New()

	router.GET("/ping", func(c *gin.Context) {
		// Add request ID to response for traceability
		// This helps correlate logs with specific requests
		requestID := c.GetString("X-Request-ID")

		c.JSON(http.StatusOK, gin.H{
			"message":   "pong",
			"requestId": requestID,
		})
	})

	// Create a server address string based on environment
	// In production, bind to all interfaces; in development, only localhost
	var addr string
	if cfg.Environment == "production" {
		addr = fmt.Sprintf(":%d", cfg.Port)
	} else {
		addr = fmt.Sprintf("localhost:%d", cfg.Port)
	}

	// Create a server with timeouts
	// Proper timeout settings are crucial for security and resource management:
	// - ReadTimeout: maximum time to read the entire request
	// - WriteTimeout: maximum time to write the response
	// - IdleTimeout: maximum time to wait for the next request when keep-alives are enabled
	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.ReadTimeoutSeconds) * time.Second,
		WriteTimeout: time.Duration(cfg.WriteTimeoutSeconds) * time.Second,
		IdleTimeout:  time.Duration(cfg.IdleTimeoutSeconds) * time.Second,
	}

	// Start server in a goroutine so it doesn't block
	// A goroutine is a lightweight thread managed by the Go runtime
	go func() {
		logger.Infof("Server starting on port %d", 8000)

		// ListenAndServe starts the HTTP server
		// It will block until the server is shut down
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", "error", err)
		}
	}()

	// Wait indefinitely
	// This keeps the main goroutine alive
	// Graceful shutdown is handled by signal handlers in middleware
	select {}
}