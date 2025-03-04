package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	lru "github.com/hashicorp/golang-lru/v2"
	"golang.org/x/time/rate"
)

type limiterItem struct {
	limiter  *rate.Limiter // Rate limiter for the IP address
	lastSeen time.Time     // Last time the IP address was seen
}

type IPRateLimiter struct {
	cache *lru.Cache[string, *limiterItem] // Stores IP addresses and their limiters
	rate  rate.Limit                       // Rate limit per second (1 token per second)
	burst int                              // Burst size (maximum bucket size, e.g 5 tokens)
}

// NewIPRateLimiter creates a new rate limiter with LRU cache for IP-based rate limiting
// Parameters:
//   - r: The rate limit per second (e.g., 1 means 1 request/second, 0.5 means 1 request/2 seconds)
//   - b: The burst size - maximum number of requests allowed to be made at once before being rate limited
//   - size: The maximum number of IP addresses to track in the LRU cache. Once this limit is reached,
//     the least recently used IP addresses will be evicted from memory
//
// Returns:
//   - An IPRateLimiter instance and nil error on success
//   - nil and error if the LRU cache creation fails
func NewIPRateLimiter(r rate.Limit, b int, size int) (*IPRateLimiter, error) {
	cache, err := lru.New[string, *limiterItem](size)
	if err != nil {
		return nil, err
	}

	return &IPRateLimiter{
		cache: cache,
		rate:  r,
		burst: b,
	}, nil
}

// GetLimiter returns the rate limiter for the provided IP.
// It provides if IP address a bucket of tokens that can be used to make requests.
func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	now := time.Now()

	// Check if the IP address is already in the cache
	if item, exists := i.cache.Get(ip); exists {
		item.lastSeen = now
		return item.limiter
	}

	// If not, create a new rate limiter and add it to the cache
	limiter := rate.NewLimiter(i.rate, i.burst)
	i.cache.Add(ip, &limiterItem{
		limiter:  limiter,
		lastSeen: now,
	})

	return limiter
}

// RateLimit middleware for Gin
func RateLimit(limiter *IPRateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()                  // Get the IP address of the client
		ipLimiter := limiter.GetLimiter(ip) // Get the rate limiter for the IP address (counter of available tokens)

		if !ipLimiter.Allow() { // Check if the IP address has available tokens
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "rate limit exceeded",
				"retry_after": time.Duration(1000/limiter.rate) * time.Millisecond,
			})
			c.Abort()
			return
		}

		// Add remaining tokens to response header
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%.2f", ipLimiter.Tokens()))
		c.Next()
	}
}
