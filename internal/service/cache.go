// Redis cache service for response caching.
// Maps to design.swift: Response Cache (Redis / In-Process)
//
// Uses a SHA-256 hash of prompt+model as the cache key.
// Nil-safe: if Redis is unavailable, all operations are no-ops.
package service

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

// Cache wraps Redis for response caching.
type Cache struct {
	rdb *redis.Client
	ttl time.Duration
}

// NewCache creates a cache service. rdb may be nil (no-op mode).
func NewCache(rdb *redis.Client, ttl time.Duration) *Cache {
	return &Cache{rdb: rdb, ttl: ttl}
}

// SemanticHash returns a deterministic cache key for an inference request.
// M4+M5 fix: Includes model, prompt, temperature, max_tokens, and userID
// to prevent cross-user and cross-parameter cache collisions.
func SemanticHash(prompt, model, userID string, temperature float64, maxTokens int) string {
	raw := fmt.Sprintf("%s:%s:%s:%.2f:%d", userID, model, prompt, temperature, maxTokens)
	h := sha256.Sum256([]byte(raw))
	return fmt.Sprintf("cache:inference:%x", h)
}

// Get retrieves a cached string value.
func (c *Cache) Get(ctx context.Context, key string) (string, error) {
	if c.rdb == nil {
		return "", nil
	}
	val, err := c.rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	if err != nil {
		slog.Warn("cache.get_error", "key", key, "error", err)
		return "", err
	}
	return val, nil
}

// Set stores a string value.
func (c *Cache) Set(ctx context.Context, key, value string) error {
	if c.rdb == nil {
		return nil
	}
	return c.rdb.Set(ctx, key, value, c.ttl).Err()
}

// GetJSON retrieves and unmarshals a cached JSON value.
func (c *Cache) GetJSON(ctx context.Context, key string, dest any) (bool, error) {
	raw, err := c.Get(ctx, key)
	if err != nil || raw == "" {
		return false, err
	}
	if err := json.Unmarshal([]byte(raw), dest); err != nil {
		slog.Warn("cache.unmarshal_error", "key", key, "error", err)
		return false, nil
	}
	return true, nil
}

// SetJSON marshals and stores a value as JSON.
func (c *Cache) SetJSON(ctx context.Context, key string, value any) error {
	if c.rdb == nil {
		return nil
	}
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("cache.marshal: %w", err)
	}
	return c.rdb.Set(ctx, key, data, c.ttl).Err()
}

// Invalidate removes a key from the cache.
func (c *Cache) Invalidate(ctx context.Context, key string) error {
	if c.rdb == nil {
		return nil
	}
	return c.rdb.Del(ctx, key).Err()
}

// Ping checks Redis connectivity.
func (c *Cache) Ping(ctx context.Context) error {
	if c.rdb == nil {
		return fmt.Errorf("redis not configured")
	}
	return c.rdb.Ping(ctx).Err()
}
