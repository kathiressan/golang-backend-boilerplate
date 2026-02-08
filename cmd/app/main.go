package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	config "ovmsa-be/configs"
	"ovmsa-be/internal/api"
	"ovmsa-be/internal/middleware"
	"ovmsa-be/pkg/database"
	"ovmsa-be/pkg/logger"
	validatorHelper "ovmsa-be/pkg/validator"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.GetConfig()

	logger.Initialize(cfg.Environment)

	// Ensure logs are flushed on shutdown
	defer logger.Sync()

	// Initialize Postgres DB
	db, err := database.Initialize(database.DBConfig{
		Host:        cfg.DBHost,
		Port:        cfg.DBPort,
		User:        cfg.DBUser,
		Password:    cfg.DBPassword,
		DBName:      cfg.DBName,
		SSLMode:     cfg.DBSSLMode,
		DatabaseURL: cfg.DatabaseURL,
	})
	if err != nil {
		logger.Fatal("Failed to initialize database", "error", err)
	}
	_ = db // Used via database.DB package variable
	logger.Info("Database initialize successfully")

	logger.Debug("Starting application",
		"environment", cfg.Environment,
		"port", cfg.Port,
		"ginMode", cfg.GinMode,
	)

	// Set Gin mode based on configuration
	gin.SetMode(cfg.GinMode)

	// Create a new router without default middleware
	router := gin.New()

	// Initialize the validator
	validatorHelper.InitValidator()

	// Setup application middleware
	middleware.SetupMiddleware(router)

	// Register all API routes
	api.ApiHandler(router)

	// Create a server address string based on environment
	var addr string
	if cfg.Environment == "production" {
		addr = fmt.Sprintf(":%d", cfg.Port)
	} else {
		addr = fmt.Sprintf("localhost:%d", cfg.Port)
	}

	// Create a server with timeouts
	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.ReadTimeoutSeconds) * time.Second,
		WriteTimeout: time.Duration(cfg.WriteTimeoutSeconds) * time.Second,
		IdleTimeout:  time.Duration(cfg.IdleTimeoutSeconds) * time.Second,
	}

	// Start server in a goroutine so it doesn't block
	go func() {
		logger.Infof("Server starting on %s", addr)

		// ListenAndServe starts the HTTP server
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", "error", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	// kill (no parameter) sends syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be caught, so no need to add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", "error", err)
	}

	logger.Info("Server exiting")
}
