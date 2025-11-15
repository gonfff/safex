package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

// ZerologMiddleware returns gin.HandlerFunc for request logging via zerolog
func ZerologMiddleware(logger zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		bodySize := c.Writer.Size()

		if raw != "" {
			path = path + "?" + raw
		}

		logger.Info().
			Str("method", method).
			Str("path", path).
			Str("ip", clientIP).
			Int("status", statusCode).
			Int("size", bodySize).
			Dur("latency", latency).
			Str("user_agent", c.Request.UserAgent()).
			Msg("HTTP request")
	}
}

// RateLimiter request rate limiter
type RateLimiter struct {
	limit  int
	window time.Duration
	mu     sync.Mutex
	store  map[string]clientWindow
}

type clientWindow struct {
	count int
	reset time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		limit:  limit,
		window: window,
		store:  make(map[string]clientWindow),
	}
}

func (r *RateLimiter) allow(key string) (remaining int, retryAfter time.Duration, ok bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	entry, exists := r.store[key]
	if !exists || now.After(entry.reset) {
		entry = clientWindow{count: 0, reset: now.Add(r.window)}
	}

	if entry.count >= r.limit {
		return 0, entry.reset.Sub(now), false
	}

	entry.count++
	r.store[key] = entry
	return r.limit - entry.count, entry.reset.Sub(now), true
}

// RateLimitMiddleware returns middleware for request rate limiting
func RateLimitMiddleware(l *RateLimiter, logger zerolog.Logger) gin.HandlerFunc {
	if l == nil {
		return func(c *gin.Context) {
			c.Next()
		}
	}
	return func(c *gin.Context) {
		key := c.ClientIP()
		if key == "" {
			key = "anonymous"
		}
		remaining, retryAfter, allowed := l.allow(key)
		if !allowed {
			c.Header("Retry-After", strconv.Itoa(int(retryAfter.Seconds())+1))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			logger.Warn().Str("ip", key).Msg("rate limit exceeded")
			return
		}
		c.Header("X-RateLimit-Limit", strconv.Itoa(l.limit))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Next()
	}
}
