package server

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

type rateLimiter struct {
	limit  int
	window time.Duration
	mu     sync.Mutex
	store  map[string]clientWindow
}

type clientWindow struct {
	count int
	reset time.Time
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	return &rateLimiter{
		limit:  limit,
		window: window,
		store:  make(map[string]clientWindow),
	}
}

func (r *rateLimiter) allow(key string) (remaining int, retryAfter time.Duration, ok bool) {
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

func rateLimitMiddleware(l *rateLimiter, logger zerolog.Logger) gin.HandlerFunc {
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
