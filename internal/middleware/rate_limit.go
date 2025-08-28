package middleware

import (
	"net/http"
	"sync"
	"time"
	"user-management-api/internal/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// RateLimiter holds the rate limiters for different clients
type RateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	rate     rate.Limit
	burst    int
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(rps rate.Limit, burst int) *RateLimiter {
	return &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rate:     rps,
		burst:    burst,
	}
}

// GetLimiter returns the rate limiter for the given key (usually IP address)
func (rl *RateLimiter) GetLimiter(key string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.limiters[key]
	if !exists {
		limiter = rate.NewLimiter(rl.rate, rl.burst)
		rl.limiters[key] = limiter
	}

	return limiter
}

// CleanupLimiters removes old limiters to prevent memory leaks
func (rl *RateLimiter) CleanupLimiters() {
	ticker := time.NewTicker(time.Minute * 5)
	go func() {
		for range ticker.C {
			rl.mu.Lock()
			for key, limiter := range rl.limiters {
				// Remove limiters that haven't been used recently
				if limiter.TokensAt(time.Now()) == float64(rl.burst) {
					delete(rl.limiters, key)
				}
			}
			rl.mu.Unlock()
		}
	}()
}

// RateLimitMiddleware creates a rate limiting middleware
// rps: requests per second, burst: maximum burst size
func RateLimitMiddleware(rps rate.Limit, burst int) gin.HandlerFunc {
	limiter := NewRateLimiter(rps, burst)
	limiter.CleanupLimiters()

	return func(c *gin.Context) {
		// Use IP address as the key for rate limiting
		key := c.ClientIP()
		
		// Get the limiter for this client
		clientLimiter := limiter.GetLimiter(key)
		
		// Check if request is allowed
		if !clientLimiter.Allow() {
			c.JSON(http.StatusTooManyRequests, models.APIResponse{
				Success: false,
				Message: "Rate limit exceeded. Please try again later.",
				Error:   "RATE_LIMIT_EXCEEDED",
			})
			c.Abort()
			return
		}
		
		c.Next()
	}
}

// StrictRateLimit - Very restrictive (1 req/sec, burst 2)
func StrictRateLimit() gin.HandlerFunc {
	return RateLimitMiddleware(1, 2)
}

// ModerateRateLimit - Moderate (10 req/sec, burst 20)
func ModerateRateLimit() gin.HandlerFunc {
	return RateLimitMiddleware(10, 20)
}

// LenientRateLimit - Lenient (100 req/sec, burst 200)
func LenientRateLimit() gin.HandlerFunc {
	return RateLimitMiddleware(100, 200)
}