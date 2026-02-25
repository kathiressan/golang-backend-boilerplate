package middleware

import (
	"strings"

	appErrors "ovmsa-be/pkg/errors"
	"ovmsa-be/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/microcosm-cc/bluemonday"
)

var strictPolicy = bluemonday.StrictPolicy()

	// XSSSanitizer is a middleware that sanitizes path and query parameters to prevent XSS attacks.
// It uses the bluemonday library to strip potentially dangerous HTML/Script tags.
func XSSSanitizer() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Content-Type Check: POST, PUT, and PATCH requests block unsupported types
		method := c.Request.Method
		if method == "POST" || method == "PUT" || method == "PATCH" {
			contentType := c.GetHeader("Content-Type")
			// If content type is specified, we must verify it is one of the supported types
			if contentType != "" &&
				!strings.Contains(contentType, "application/json") &&
				!strings.Contains(contentType, "multipart/form-data") &&
				!strings.Contains(contentType, "application/x-www-form-urlencoded") {
				
				appErr := appErrors.New(nil, 415, "UNSUPPORTED_MEDIA_TYPE", "Unsupported media type", "UnsupportedMediaType")
				response.Error(c, appErr)
				c.Abort()
				return
			}
		}

		// 2. Sanitize Path Params (removes <script>, etc.)
		for i := range c.Params {
			c.Params[i].Value = strictPolicy.Sanitize(c.Params[i].Value)
		}

		// 3. Sanitize Query Params
		queries := c.Request.URL.Query()
		for key, values := range queries {
			for i := range values {
				queries[key][i] = strictPolicy.Sanitize(values[i])
			}
		}
		c.Request.URL.RawQuery = queries.Encode()

		c.Next()
	}
}
