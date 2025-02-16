package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/time/rate"
)

func TestBasicRateLimiting(t *testing.T) {
	// Create a new rate limiter: 2 requests per second with burst of 2
	limiter, err := NewIPRateLimiter(rate.Limit(2), 2, 1000)
	require.NoError(t, err)

	// Create a test router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RateLimit(limiter))
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// First two requests should succeed (burst)
	assert.Equal(t, http.StatusOK, makeRequestFromIP(router, "192.168.1.1"))
	assert.Equal(t, http.StatusOK, makeRequestFromIP(router, "192.168.1.1"))

	// Third request should be rate limited
	assert.Equal(t, http.StatusTooManyRequests, makeRequestFromIP(router, "192.168.1.1"))

	// Wait for rate limit to reset
	time.Sleep(1 * time.Second)

	// Should succeed again
	assert.Equal(t, http.StatusOK, makeRequestFromIP(router, "192.168.1.1"))
}

func TestMultipleIPRateLimiting(t *testing.T) {
	limiter, err := NewIPRateLimiter(rate.Limit(2), 2, 1000)
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RateLimit(limiter))
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// Test two different IPs
	// IP1 should not affect IP2's rate limit
	assert.Equal(t, http.StatusOK, makeRequestFromIP(router, "192.168.1.1"))
	assert.Equal(t, http.StatusOK, makeRequestFromIP(router, "192.168.1.1"))
	assert.Equal(t, http.StatusTooManyRequests, makeRequestFromIP(router, "192.168.1.1"))

	assert.Equal(t, http.StatusOK, makeRequestFromIP(router, "192.168.1.2"))
	assert.Equal(t, http.StatusOK, makeRequestFromIP(router, "192.168.1.2"))
	assert.Equal(t, http.StatusTooManyRequests, makeRequestFromIP(router, "192.168.1.2"))
}

func TestCacheEviction(t *testing.T) {
	// Small cache size to test eviction
	limiter, err := NewIPRateLimiter(rate.Limit(1), 1, 2)
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RateLimit(limiter))
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// Fill cache
	assert.Equal(t, http.StatusOK, makeRequestFromIP(router, "192.168.1.1"))
	assert.Equal(t, http.StatusOK, makeRequestFromIP(router, "192.168.1.2"))

	// Add third IP, should evict first IP
	assert.Equal(t, http.StatusOK, makeRequestFromIP(router, "192.168.1.3"))

	// First IP should work again as it was evicted
	assert.Equal(t, http.StatusOK, makeRequestFromIP(router, "192.168.1.1"))
}

// helper function to make a request from an IP address
func makeRequestFromIP(router *gin.Engine, ip string) int {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Real-IP", ip)
	router.ServeHTTP(w, req)
	return w.Code
}
