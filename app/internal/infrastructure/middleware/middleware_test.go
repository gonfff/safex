package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

func TestZerologMiddleware(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)

	var logBuffer bytes.Buffer
	logger := zerolog.New(&logBuffer)

	router := gin.New()
	router.Use(ZerologMiddleware(logger))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})

	// Test request
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	// Verify logging occurred (log buffer should contain something)
	if logBuffer.Len() == 0 {
		t.Error("expected log output but got none")
	}
}

func TestZerologMiddleware_WithQueryParams(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)

	var logBuffer bytes.Buffer
	logger := zerolog.New(&logBuffer)

	router := gin.New()
	router.Use(ZerologMiddleware(logger))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})

	// Test request with query params
	req := httptest.NewRequest("GET", "/test?param=value", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	// Verify logging occurred
	if logBuffer.Len() == 0 {
		t.Error("expected log output but got none")
	}
}

func TestNewRateLimiter(t *testing.T) {
	limiter := NewRateLimiter(10, time.Minute)

	if limiter == nil {
		t.Error("expected rate limiter, got nil")
		return
	}
	if limiter.limit != 10 {
		t.Errorf("expected limit 10, got %d", limiter.limit)
	}
}

func TestRateLimiter_Allow_FirstRequest(t *testing.T) {
	limiter := NewRateLimiter(5, time.Minute)

	_, _, allowed := limiter.allow("127.0.0.1")
	if !allowed {
		t.Error("expected first request to be allowed")
	}
}

func TestRateLimiter_Allow_UnderLimit(t *testing.T) {
	limiter := NewRateLimiter(5, time.Minute)
	client := "127.0.0.1"

	// Make requests under limit
	for i := 0; i < 5; i++ {
		_, _, allowed := limiter.allow(client)
		if !allowed {
			t.Errorf("expected request %d to be allowed", i+1)
		}
	}
}

func TestRateLimiter_Allow_OverLimit(t *testing.T) {
	limiter := NewRateLimiter(2, time.Minute)
	client := "127.0.0.1"

	// First two requests should be allowed
	_, _, ok1 := limiter.allow(client)
	if !ok1 {
		t.Error("expected first request to be allowed")
	}
	_, _, ok2 := limiter.allow(client)
	if !ok2 {
		t.Error("expected second request to be allowed")
	}

	// Third request should be blocked
	_, _, ok3 := limiter.allow(client)
	if ok3 {
		t.Error("expected third request to be blocked")
	}
}

func TestRateLimiter_Allow_DifferentClients(t *testing.T) {
	limiter := NewRateLimiter(1, time.Minute)

	// Different clients should have separate limits
	_, _, ok1 := limiter.allow("127.0.0.1")
	if !ok1 {
		t.Error("expected request from first client to be allowed")
	}
	_, _, ok2 := limiter.allow("127.0.0.2")
	if !ok2 {
		t.Error("expected request from second client to be allowed")
	}

	// But same client should be blocked
	_, _, ok3 := limiter.allow("127.0.0.1")
	if ok3 {
		t.Error("expected second request from first client to be blocked")
	}
}

func TestRateLimitMiddleware_Allowed(t *testing.T) {
	gin.SetMode(gin.TestMode)

	limiter := NewRateLimiter(10, 60)
	logger := zerolog.Nop()
	middleware := RateLimitMiddleware(limiter, logger)

	router := gin.New()
	router.Use(middleware)
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestRateLimitMiddleware_Blocked(t *testing.T) {
	gin.SetMode(gin.TestMode)

	limiter := NewRateLimiter(1, time.Minute)
	logger := zerolog.Nop()
	middleware := RateLimitMiddleware(limiter, logger)

	router := gin.New()
	router.Use(middleware)
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})

	// First request should pass
	req1 := httptest.NewRequest("GET", "/test", nil)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	if w1.Code != http.StatusOK {
		t.Errorf("expected first request status 200, got %d", w1.Code)
	}

	// Second request should be blocked
	req2 := httptest.NewRequest("GET", "/test", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	if w2.Code != http.StatusTooManyRequests {
		t.Errorf("expected second request status 429, got %d", w2.Code)
	}
}
