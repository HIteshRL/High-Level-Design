// Rate limiter middleware using Redis sliding window.
// Maps to design.swift: Rate Limiter (Token Bucket / Sliding Window).
//
// Uses sorted sets in Redis for distributed rate limiting.
// Falls back to in-memory tracking when Redis is unavailable.
package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// RateLimiter implements sliding-window rate limiting.
type RateLimiter struct {
	rdb         *redis.Client
	maxRequests int
	window      time.Duration
	// In-memory fallback
	mu       sync.Mutex
	counters map[string][]time.Time
}

// NewRateLimiter creates a rate limiter. If rdb is nil, uses in-memory fallback.
func NewRateLimiter(rdb *redis.Client, maxRPM int) *RateLimiter {
	return &RateLimiter{
		rdb:         rdb,
		maxRequests: maxRPM,
		window:      time.Minute,
		counters:    make(map[string][]time.Time),
	}
}

// Middleware returns the rate-limiting HTTP middleware.
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip health checks
		if r.URL.Path == "/health" {
			next.ServeHTTP(w, r)
			return
		}

		key := rl.clientKey(r)
		allowed, remaining, retryAfter := rl.allow(r.Context(), key)

		w.Header().Set("X-RateLimit-Limit", strconv.Itoa(rl.maxRequests))
		w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))

		if !allowed {
			w.Header().Set("Retry-After", strconv.Itoa(int(retryAfter.Seconds())))
			http.Error(w, `{"error":"rate limit exceeded","code":429}`, http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (rl *RateLimiter) clientKey(r *http.Request) string {
	// H1 fix: Don't trust X-Forwarded-For — use RemoteAddr directly.
	// In production, strip/validate XFF at the reverse proxy (nginx, envoy) layer.
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return "rl:" + r.RemoteAddr
	}
	return "rl:" + host
}

func (rl *RateLimiter) allow(ctx context.Context, key string) (allowed bool, remaining int, retryAfter time.Duration) {
	if rl.rdb != nil {
		return rl.allowRedis(ctx, key)
	}
	return rl.allowMemory(key)
}

// allowRedis uses a Redis sorted set for the sliding window.
func (rl *RateLimiter) allowRedis(ctx context.Context, key string) (bool, int, time.Duration) {
	now := time.Now()
	windowStart := now.Add(-rl.window)

	pipe := rl.rdb.Pipeline()
	// Remove expired entries
	pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart.UnixMicro()))
	// Add current request
	pipe.ZAdd(ctx, key, redis.Z{Score: float64(now.UnixMicro()), Member: now.UnixMicro()})
	// Count entries in window
	countCmd := pipe.ZCard(ctx, key)
	// Set TTL on key
	pipe.Expire(ctx, key, rl.window+time.Second)

	if _, err := pipe.Exec(ctx); err != nil {
		slog.Warn("ratelimiter.redis_error", "error", err)
		// Fail open
		return true, rl.maxRequests, 0
	}

	count := int(countCmd.Val())
	remaining := rl.maxRequests - count
	if remaining < 0 {
		remaining = 0
	}

	if count > rl.maxRequests {
		return false, 0, rl.window
	}

	return true, remaining, 0
}

// allowMemory is the in-memory fallback (single-instance only).
func (rl *RateLimiter) allowMemory(key string) (bool, int, time.Duration) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-rl.window)

	// Prune old entries
	timestamps := rl.counters[key]
	valid := timestamps[:0]
	for _, ts := range timestamps {
		if ts.After(windowStart) {
			valid = append(valid, ts)
		}
	}

	if len(valid) >= rl.maxRequests {
		rl.counters[key] = valid
		return false, 0, rl.window
	}

	valid = append(valid, now)
	rl.counters[key] = valid
	remaining := rl.maxRequests - len(valid)
	if remaining < 0 {
		remaining = 0
	}
	return true, remaining, 0
}
