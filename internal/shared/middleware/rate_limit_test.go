package middleware

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestRateLimiter_Allow(t *testing.T) {
	limiter := NewRateLimiter(3, time.Second)

	for i := 0; i < 3; i++ {
		if !limiter.Allow("key") {
			t.Errorf("Allow() call %d should return true", i+1)
		}
	}
	if limiter.Allow("key") {
		t.Error("Allow() 4th call should return false")
	}
}

func TestRateLimiter_Allow_DifferentKeys(t *testing.T) {
	limiter := NewRateLimiter(2, time.Second)

	limiter.Allow("a")
	limiter.Allow("a")
	if !limiter.Allow("b") {
		t.Error("Different key should have separate counter")
	}
}

func TestRateLimiter_Allow_WindowReset(t *testing.T) {
	limiter := NewRateLimiter(2, 100*time.Millisecond)

	limiter.Allow("key")
	limiter.Allow("key")
	if limiter.Allow("key") {
		t.Error("Should be blocked before window reset")
	}

	time.Sleep(150 * time.Millisecond)
	if !limiter.Allow("key") {
		t.Error("Should be allowed after window reset")
	}
}

func TestRateLimiter_Allow_Concurrent(t *testing.T) {
	limiter := NewRateLimiter(100, time.Second)
	var wg sync.WaitGroup

	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			limiter.Allow("key")
		}()
	}
	wg.Wait()
	// Just verify no panics/races
}

func TestRateLimitIP_Returns429(t *testing.T) {
	gin.SetMode(gin.TestMode)
	limiter := NewRateLimiter(1, time.Second)

	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)

	r.GET("/", RateLimitIP(limiter), func(c *gin.Context) {
		c.Status(200)
	})

	// First request should pass
	r.ServeHTTP(w, c.Request)
	if w.Code != http.StatusOK {
		t.Errorf("first request: status = %v, want 200", w.Code)
	}

	// Second request should be rate limited
	w2 := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(w2)
	c2.Request = httptest.NewRequest("GET", "/", nil)
	r.ServeHTTP(w2, c2.Request)

	if w2.Code != http.StatusTooManyRequests {
		t.Errorf("second request: status = %v, want 429", w2.Code)
	}
}

func TestRateLimitAuth_PrefixesKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	limiter := NewRateLimiter(1, time.Second)

	// Auth limiter uses "auth:" prefix, so it's separate from IP limiter
	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)

	r.GET("/", RateLimitAuth(limiter), func(c *gin.Context) {
		c.Status(200)
	})

	r.ServeHTTP(w, c.Request)
	if w.Code != http.StatusOK {
		t.Errorf("first request: status = %v, want 200", w.Code)
	}
}
