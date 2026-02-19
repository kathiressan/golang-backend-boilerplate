package middleware

import (
	"sync"
	"time"

	config "ovmsa-be/configs"
	"ovmsa-be/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/hashicorp/golang-lru/v2/expirable"
)

// rateLimitEntry tracks request count and window start time for an IP
type rateLimitEntry struct {
	mu          sync.Mutex
	count       int
	windowStart time.Time
}

// rateLimitCache stores per-IP rate limit entries with a 2-minute TTL so that
// entries for IPs that have gone quiet are automatically evicted, preventing
// unbounded memory growth. Capacity is set to 10 000 unique IPs; the LRU
// eviction policy drops the least-recently-seen IP when the cap is reached.
var rateLimitCache = expirable.NewLRU[string, *rateLimitEntry](10_000, nil, 2*time.Minute)

// rateLimitCacheMu guards cache-level get-or-create operations so that two
// goroutines from the same IP cannot both observe a cache miss and insert
// duplicate entries simultaneously.
var rateLimitCacheMu sync.Mutex

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

		// Get or create the entry for this IP under the cache-level lock to
		// eliminate the TOCTOU race between Get and Add.
		rateLimitCacheMu.Lock()
		entry, ok := rateLimitCache.Get(clientIP)
		if !ok {
			entry = &rateLimitEntry{count: 0, windowStart: now}
			rateLimitCache.Add(clientIP, entry)
		}
		rateLimitCacheMu.Unlock()

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
