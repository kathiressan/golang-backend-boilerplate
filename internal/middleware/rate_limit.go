package middleware

import (
	config "ovmsa-be/configs"
	"ovmsa-be/pkg/response"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hashicorp/golang-lru/v2/expirable"
)

// rateLimitEntry tracks request count and window start time for an IP
type rateLimitEntry struct {
	count      int
	windowStart time.Time
}

var (
	// rateLimitCache tracks requests per IP address
	// Key: IP address, Value: rateLimitEntry
	rateLimitCache = expirable.NewLRU[string, *rateLimitEntry](10000, nil, 1*time.Minute)
)

// RateLimitMiddleware implements a simple rate limiter based on IP address
// It limits the number of requests per minute per IP
func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		cfg := config.GetConfig()

		// Skip if rate limiting is disabled
		if !cfg.RateLimitEnabled {
			c.Next()
			return
		}

		clientIP := c.ClientIP()
		now := time.Now()

		// Get or create entry for this IP
		entry, exists := rateLimitCache.Get(clientIP)
		if !exists {
			// First request from this IP in the current window
			entry = &rateLimitEntry{
				count:      1,
				windowStart: now,
			}
			rateLimitCache.Add(clientIP, entry)
			c.Next()
			return
		}

		// Check if we're still in the same 1-minute window
		if now.Sub(entry.windowStart) > 1*time.Minute {
			// New window, reset counter
			entry.count = 1
			entry.windowStart = now
			rateLimitCache.Add(clientIP, entry)
			c.Next()
			return
		}

		// Increment counter for current window
		entry.count++

		// Check if limit exceeded
		if entry.count > cfg.RateLimitRequestsPerMinute {
			response.TooManyRequestsResponse(c, nil, "Rate limit exceeded. Please try again later.")
			c.Abort()
			return
		}

		// Update cache with new count
		rateLimitCache.Add(clientIP, entry)
		c.Next()
	}
}
