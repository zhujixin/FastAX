package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/fastax/fastax-server/internal/shared/response"
	"github.com/gin-gonic/gin"
)

type rateEntry struct {
	count    int
	resetAt  time.Time
}

type RateLimiter struct {
	mu       sync.Mutex
	entries  map[string]*rateEntry
	limit    int
	window   time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		entries: make(map[string]*rateEntry),
		limit:   limit,
		window:  window,
	}
}

func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	entry, exists := rl.entries[key]

	if !exists || now.After(entry.resetAt) {
		rl.entries[key] = &rateEntry{
			count:   1,
			resetAt: now.Add(rl.window),
		}
		return true
	}

	if entry.count >= rl.limit {
		return false
	}

	entry.count++
	return true
}

// RateLimitIP limits by client IP (60 req/min default)
func RateLimitIP(limiter *RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !limiter.Allow(ip) {
			response.Error(c, http.StatusTooManyRequests, response.CodeRateLimited)
			return
		}
		c.Next()
	}
}

// RateLimitAuth limits auth endpoints (5 req/min)
func RateLimitAuth(limiter *RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := "auth:" + c.ClientIP()
		if !limiter.Allow(key) {
			response.Error(c, http.StatusTooManyRequests, response.CodeRateLimited)
			return
		}
		c.Next()
	}
}
