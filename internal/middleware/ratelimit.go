package middleware

import (
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/nexbic/platform/pkg/response"
)

type RateLimiter struct {
	mu       sync.Mutex
	requests map[string]*rateEntry
	limit    int
	window   time.Duration
}

type rateEntry struct {
	count       int
	windowStart time.Time
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[string]*rateEntry),
		limit:    limit,
		window:   window,
	}
	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(rl.window)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, entry := range rl.requests {
			if now.Sub(entry.windowStart) > rl.window {
				delete(rl.requests, key)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) Check(key string) (allowed bool, remaining int) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	entry, exists := rl.requests[key]

	if !exists || now.Sub(entry.windowStart) > rl.window {
		rl.requests[key] = &rateEntry{
			count:       1,
			windowStart: now,
		}
		return true, rl.limit - 1
	}

	if entry.count >= rl.limit {
		return false, 0
	}

	entry.count++
	return true, rl.limit - entry.count
}

func RateLimit(limit int, window time.Duration) fiber.Handler {
	limiter := NewRateLimiter(limit, window)

	return func(c *fiber.Ctx) error {
		key := c.IP()

		allowed, remaining := limiter.Check(key)

		c.Set("X-RateLimit-Limit", itoa(limit))
		c.Set("X-RateLimit-Remaining", itoa(remaining))

		if !allowed {
			return response.TooManyRequests(c, "rate limit exceeded")
		}

		return c.Next()
	}
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	s := ""
	neg := false
	if i < 0 {
		neg = true
		i = -i
	}
	for i > 0 {
		s = string(rune('0'+i%10)) + s
		i /= 10
	}
	if neg {
		s = "-" + s
	}
	return s
}
