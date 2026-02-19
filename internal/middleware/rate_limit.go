package middleware

import (
	"sync"
	"time"

	config "ovmsa-be/configs"
	"ovmsa-be/pkg/response"

	"github.com/gin-gonic/gin"
)

// rateLimitEntry tracks request count and window start time for an IP
type rateLimitEntry struct {
	mu          sync.Mutex
	count       int
	windowStart time.Time
}

// rateLimitStore provides atomic get-or-create semantics for rate limit entries.
// sync.Map is used over a plain LRU here because LoadOrStore is atomic, which
// eliminates the check-then-act race where two goroutines from the same IP both
// see a cache miss and clobber each other's entries.
var rateLimitStore sync.Map

// RateLimitMiddleware implements a simple rate limiter based on IP address.
// It limits the number of requests per minute per IP.
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

		// LoadOrStore atomically gets the existing entry or stores a new one.
		// This eliminates the TOCTOU race that existed with the LRU check-then-add pattern.
		newEntry := &rateLimitEntry{count: 0, windowStart: now}
		actual, _ := rateLimitStore.LoadOrStore(clientIP, newEntry)
		entry := actual.(*rateLimitEntry)

		entry.mu.Lock()
		defer entry.mu.Unlock()

		// Check if we're still in the same 1-minute window
		if now.Sub(entry.windowStart) > 1*time.Minute {
			// New window: reset counter and slide the window
			entry.count = 1
			entry.windowStart = now
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

		c.Next()
	}
}
