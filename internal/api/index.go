package api

import (
	"net/http"
	"ovmsa-be/pkg/response"

	"github.com/gin-gonic/gin"
)

// ApiHandler is the main function that sets up all API routes
func ApiHandler(router *gin.Engine) {
	// Register base routes
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, response.Success("success", "server up and running"))
	})

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, response.Success("success", "server healthy"))
	})

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, response.Success("success", "pong"))
	})

	// Silent handler for browser favicon requests to keep logs clean
	router.GET("/favicon.ico", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})
}
