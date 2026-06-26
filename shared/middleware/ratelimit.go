package middleware

import (
	"context"
	"log/slog"
	"strconv"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/nexbic/platform/shared/utils"
	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	mu       sync.Mutex
	requests map[string]*rateEntry
	limit    int
	window   time.Duration
	rdb      *redis.Client
	useRedis bool
}

type rateEntry struct {
	count       int
	windowStart time.Time
}

func NewRateLimiter(limit int, window time.Duration, rdb *redis.Client) *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[string]*rateEntry),
		limit:    limit,
		window:   window,
		rdb:      rdb,
		useRedis: rdb != nil,
	}

	if !rl.useRedis {
		go rl.cleanup()
	}

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
	if rl.useRedis {
		return rl.checkRedis(key)
	}
	return rl.checkMemory(key)
}

func (rl *RateLimiter) checkRedis(key string) (bool, int) {
	window := int64(rl.window.Seconds())
	now := time.Now().Unix()
	windowKey := "ratelimit:" + key

	pipe := rl.rdb.Pipeline()
	pipe.ZRemRangeByScore(context.Background(), windowKey, "0", strconv.FormatInt(now-window, 10))
	countCmd := pipe.ZCard(context.Background(), windowKey)
	pipe.ZAdd(context.Background(), windowKey, redis.Z{
		Score:  float64(now),
		Member: float64(now) + float64(now%1000)/1000,
	})
	pipe.Expire(context.Background(), windowKey, time.Duration(window)*time.Second)
	_, err := pipe.Exec(context.Background())
	if err != nil {
		slog.Warn("redis rate limit error, falling back to in-memory", "error", err)
		return rl.checkMemory(key)
	}

	count := int(countCmd.Val())
	remaining := rl.limit - count
	if remaining < 0 {
		remaining = 0
	}
	return count <= rl.limit, remaining
}

func (rl *RateLimiter) checkMemory(key string) (bool, int) {
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
		RecordRateLimitExceeded()
		return false, 0
	}

	return true, rl.limit - entry.count
}

func RateLimit(limit int, window time.Duration, rdb *redis.Client) fiber.Handler {
	rl := NewRateLimiter(limit, window, rdb)

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
