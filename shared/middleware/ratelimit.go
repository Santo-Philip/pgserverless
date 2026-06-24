package middleware

import (
	"strconv"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/nexbic/platform/shared/utils"
)

type RateLimiter struct {
	mu       sync.Mutex
	requests map[string]*rateEntry
	limit    int
	window   time.Duration
}

type rateEntry struct {
	count    int
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
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, entry := range rl.requests {
			if now.Sub(entry.windowStart) > rl.window*2 {
				delete(rl.requests, key)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) Check(key string) (bool, int) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	entry, exists := rl.requests[key]

	if !exists || now.Sub(entry.windowStart) >= rl.window {
		rl.requests[key] = &rateEntry{
			count:       1,
			windowStart: now,
		}
		return true, rl.limit - 1
	}

	entry.count++
	if entry.count > rl.limit {
		return false, 0
	}

	return true, rl.limit - entry.count
}

func RateLimit(limit int, window time.Duration) fiber.Handler {
	rl := NewRateLimiter(limit, window)

	return func(c *fiber.Ctx) error {
		key := c.IP()
		allowed, remaining := rl.Check(key)

		c.Set("X-RateLimit-Limit", strconv.Itoa(limit))
		c.Set("X-RateLimit-Remaining", strconv.Itoa(remaining))

		if !allowed {
			return utils.Error(c, fiber.StatusTooManyRequests, "rate_limit_exceeded", "too many requests")
		}

		return c.Next()
	}
}
