package redisrepo

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// CacheRepository provides caching functionality using Redis
type CacheRepository interface {
	// Cache an item with automatic serialization/deserialization
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string, dest interface{}) error
	Invalidate(ctx context.Context, keys ...string) error

	// Cache with tags for group invalidation
	SetWithTags(ctx context.Context, key string, value interface{}, expiration time.Duration, tags ...string) error
	InvalidateByTag(ctx context.Context, tag string) error
}

// cacheRepository implements the CacheRepository interface
type cacheRepository struct {
	redis Repository
}

// NewCacheRepository creates a new cache repository
func NewCacheRepository(redis Repository) CacheRepository {
	return &cacheRepository{
		redis: redis,
	}
}

// Set stores a serialized value in the cache
func (c *cacheRepository) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	// Serialize the value to JSON
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to serialize value: %w", err)
	}

	// Store in Redis
	return c.redis.Set(ctx, key, data, expiration)
}

// Get retrieves and deserializes a value from the cache
func (c *cacheRepository) Get(ctx context.Context, key string, dest interface{}) error {
	// Get from Redis
	data, err := c.redis.Get(ctx, key)
	if err != nil {
		return err
	}

	// Deserialize the value from JSON
	return json.Unmarshal([]byte(data), dest)
}

// Invalidate removes keys from the cache
func (c *cacheRepository) Invalidate(ctx context.Context, keys ...string) error {
	return c.redis.Delete(ctx, keys...)
}

// SetWithTags stores a value in the cache and associates it with tags for later invalidation
func (c *cacheRepository) SetWithTags(ctx context.Context, key string, value interface{}, expiration time.Duration, tags ...string) error {
	// First store the value
	if err := c.Set(ctx, key, value, expiration); err != nil {
		return err
	}

	// Then associate the key with each tag
	for _, tag := range tags {
		tagKey := fmt.Sprintf("tag:%s", tag)
		if err := c.redis.SAdd(ctx, tagKey, key); err != nil {
			return fmt.Errorf("failed to associate key with tag %s: %w", tag, err)
		}

		// Make sure the tag set doesn't expire
		if err := c.redis.Expire(ctx, tagKey, 0); err != nil {
			return fmt.Errorf("failed to set expiration for tag %s: %w", tag, err)
		}
	}

	return nil
}

// InvalidateByTag invalidates all cache entries associated with the given tag
func (c *cacheRepository) InvalidateByTag(ctx context.Context, tag string) error {
	tagKey := fmt.Sprintf("tag:%s", tag)

	// Get all keys associated with the tag
	keys, err := c.redis.SMembers(ctx, tagKey)
	if err != nil {
		return fmt.Errorf("failed to get keys for tag %s: %w", tag, err)
	}

	if len(keys) == 0 {
		return nil
	}

	// Delete all the keys
	if err := c.redis.Delete(ctx, keys...); err != nil {
		return fmt.Errorf("failed to delete keys for tag %s: %w", tag, err)
	}

	// Clear the tag set itself
	return c.redis.Delete(ctx, tagKey)
}
