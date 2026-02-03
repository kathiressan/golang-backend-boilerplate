package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/microcosm-cc/bluemonday"
)

var strictPolicy = bluemonday.StrictPolicy()

func GlobalSanitizer() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 1. Content-Type Check (Simple)
        if (c.Request.Method == "POST" || c.Request.Method == "PUT") && 
            !strings.Contains(c.GetHeader("Content-Type"), "application/json") {
            c.AbortWithStatusJSON(415, gin.H{"error": "JSON required"})
            return
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