package redisrepo

import (
	"context"
	"fmt"
	"strconv"
	"time"
)

// RateLimitRepository provides rate limiting functionality
type RateLimitRepository interface {
	// Increment increments a counter and returns the current count
	// If the key doesn't exist, it's created with a value of 1
	Increment(ctx context.Context, key string, expiration time.Duration) (int, error)

	// IncrementPreserveTTL increments the counter and preserves the existing TTL if the key exists
	// If the key doesn't exist, it sets the provided expiration
	IncrementPreserveTTL(ctx context.Context, key string, defaultExpiration time.Duration) (int, error)

	// Check checks if a rate limit has been exceeded without incrementing
	Check(ctx context.Context, key string) (int, error)

	// Reset resets a rate limit
	Reset(ctx context.Context, key string) error

	// SetState sets the bucket state of a rate limit
	SetState(ctx context.Context, key string, state interface{}, expiration time.Duration) error

	// GetState gets the bucket state of a rate limit
	GetState(ctx context.Context, key string) (string, error)
}

// rateLimitRepository implements the RateLimitRepository interface
type rateLimitRepository struct {
	redis Repository
}

// NewRateLimitRepository creates a new rate limit repository
func NewRateLimitRepository(redis Repository) RateLimitRepository {
	return &rateLimitRepository{
		redis: redis,
	}
}

// Increment increments a counter and returns the current count
func (r *rateLimitRepository) Increment(ctx context.Context, key string, expiration time.Duration) (int, error) {
	// Check if the key exists
	exists, err := r.redis.Exists(ctx, key)
	if err != nil {
		return 0, fmt.Errorf("failed to check if key exists: %w", err)
	}

	var count int

	if !exists {
		// Key doesn't exist, create it with a value of 1
		if err := r.redis.Set(ctx, key, "1", expiration); err != nil {
			return 0, fmt.Errorf("failed to create rate limit counter: %w", err)
		}
		count = 1
	} else {
		// Key exists, increment it
		countStr, err := r.redis.Get(ctx, key)
		if err != nil {
			return 0, fmt.Errorf("failed to get rate limit counter: %w", err)
		}

		count, err = strconv.Atoi(countStr)
		if err != nil {
			return 0, fmt.Errorf("invalid rate limit counter value: %w", err)
		}

		count++

		if err := r.redis.Set(ctx, key, strconv.Itoa(count), expiration); err != nil {
			return 0, fmt.Errorf("failed to update rate limit counter: %w", err)
		}
	}

	return count, nil
}

// IncrementPreserveTTL increments a counter and preserves the existing TTL
func (r *rateLimitRepository) IncrementPreserveTTL(ctx context.Context, key string, defaultExpiration time.Duration) (int, error) {
	// Check if the key exists
	val, err := r.redis.Increment(ctx, key)
	if err != nil {
		return 0, fmt.Errorf("failed to increment counter: %w", err)
	}

	// If this is a new key (val == 1) or expiration is provided, set the expiration
	if val == 1 && defaultExpiration > 0 {
		// Only set expiration if it's greater than 0
		if defaultExpiration > 0 {
			if err := r.redis.Expire(ctx, key, defaultExpiration); err != nil {
				return 0, fmt.Errorf("failed to set expiration: %w", err)
			}
		}
	}

	return int(val), nil
}

// Check checks if a rate limit has been exceeded without incrementing
func (r *rateLimitRepository) Check(ctx context.Context, key string) (int, error) {
	// Check if the key exists
	exists, err := r.redis.Exists(ctx, key)
	if err != nil {
		return 0, fmt.Errorf("failed to check if key exists: %w", err)
	}

	if !exists {
		return 0, nil
	}

	// Key exists, get its value
	countStr, err := r.redis.Get(ctx, key)
	if err != nil {
		return 0, fmt.Errorf("failed to get rate limit counter: %w", err)
	}

	count, err := strconv.Atoi(countStr)
	if err != nil {
		return 0, fmt.Errorf("invalid rate limit counter value: %w", err)
	}

	return count, nil
}

// Reset resets a rate limit counter
func (r *rateLimitRepository) Reset(ctx context.Context, key string) error {
	return r.redis.Delete(ctx, key)
}

// SetState sets the bucket state of a rate limit
func (r *rateLimitRepository) SetState(ctx context.Context, key string, state interface{}, expiration time.Duration) error {
	err := r.redis.Set(ctx, key, state, expiration)
	if err != nil {
		return fmt.Errorf("failed to set state: %w", err)
	}
	return nil
}

// GetState gets the bucket state of a rate limit
func (r *rateLimitRepository) GetState(ctx context.Context, key string) (string, error) {
	val, err := r.redis.Get(ctx, key)
	if err != nil {
		return "", fmt.Errorf("failed to get state: %w", err)
	}
	return val, nil
}
