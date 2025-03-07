package ratelimit

import (
	"context"
	"time"
)

// RateLimitRepo defines repository operations for rate limiting
type RateLimitRepo interface {
	// IncrementPreserveTTL increments the counter and preserves the existing TTL if the key exists
	// If the key doesn't exist, it sets the provided expiration
	IncrementPreserveTTL(ctx context.Context, key string, defaultExpiration time.Duration) (int, error)

	// Check returns the current count for the given key without incrementing
	Check(ctx context.Context, key string) (int, error)

	// Reset resets the counter for the given key
	Reset(ctx context.Context, key string) error

	// SetState sets the bucket state of a rate limit
	SetState(ctx context.Context, key string, state interface{}, expiration time.Duration) error

	// GetState gets the bucket state of a rate limit
	GetState(ctx context.Context, key string) (string, error)
}

// RateLimitResponse represents the rate limit information returned in headers
type RateLimitResponse struct {
	Limit     int   `json:"limit"`
	Remaining int   `json:"remaining"`
	Reset     int64 `json:"reset"`
}

var repo RateLimitRepo

// SetRateLimitRepo sets the rate limit repository implementation
func SetRateLimitRepo(r RateLimitRepo) {
	repo = r
}

// GetRateLimitRepo returns the current rate limit repository implementation
func GetRateLimitRepo() RateLimitRepo {
	return repo
}
