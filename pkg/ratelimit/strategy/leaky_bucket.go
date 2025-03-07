package strategy

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"go-echo-mongo/pkg/ratelimit"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// LeakyBucketStore implements the leaky bucket algorithm
type LeakyBucketStore struct {
	repo      ratelimit.RateLimitRepo
	capacity  int     // Maximum bucket capacity
	leakRate  float64 // Requests per second that leak out
	expiresIn time.Duration
	keyPrefix string // Key prefix for rate limit
}

type leakyBucketState struct {
	Water    int       `json:"water"`     // Current water level
	LastLeak time.Time `json:"last_leak"` // Last time we leaked water
}

// NewLeakyBucketStore creates a new leaky bucket rate limiter
func NewLeakyBucketStore(repo ratelimit.RateLimitRepo, capacity int, leakRate float64, expiresIn time.Duration) *LeakyBucketStore {
	return &LeakyBucketStore{
		repo:      repo,
		capacity:  capacity,
		leakRate:  leakRate,
		expiresIn: expiresIn,
		keyPrefix: "rate_limit_leaky_bucket",
	}
}

// Allow implements the RateLimiterStore interface
func (s *LeakyBucketStore) Allow(identifier string) (bool, error) {
	ctx := context.Background()
	key := fmt.Sprintf("%s:%s", s.keyPrefix, identifier)

	// Get or initialize bucket state
	state, err := s.getBucketState(ctx, key)
	if err != nil {
		return false, err
	}

	// Calculate leakage
	now := time.Now()
	elapsed := now.Sub(state.LastLeak).Seconds()
	leaked := int(elapsed * s.leakRate)

	// Update water level
	state.Water = max(0, state.Water-leaked)

	// Check if we can add more water
	if state.Water >= s.capacity {
		// Save state even if we're denying the request
		if err := s.saveBucketState(ctx, key, state); err != nil {
			return false, err
		}
		return false, nil
	}

	// Add one unit of water and update last leak time
	state.Water++
	state.LastLeak = now

	// Save updated state
	if err := s.saveBucketState(ctx, key, state); err != nil {
		return false, err
	}

	return true, nil
}

func (s *LeakyBucketStore) getBucketState(ctx context.Context, key string) (*leakyBucketState, error) {
	stateJSON, err := s.repo.GetState(ctx, key)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return &leakyBucketState{
				Water:    0,
				LastLeak: time.Now(),
			}, nil
		}
		return nil, err
	}

	var state leakyBucketState
	if err := json.Unmarshal([]byte(stateJSON), &state); err != nil {
		return nil, fmt.Errorf("failed to unmarshal leaky bucket state: %w", err)
	}
	return &state, nil
}

func (s *LeakyBucketStore) saveBucketState(ctx context.Context, key string, state *leakyBucketState) error {
	stateJSON, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("failed to marshal leaky bucket state: %w", err)
	}
	return s.repo.SetState(ctx, key, string(stateJSON), s.expiresIn)
}

// GetRateLimitInfo returns information about the current rate limit state
func (s *LeakyBucketStore) GetRateLimitInfo(identifier string) (*ratelimit.RateLimitResponse, error) {
	ctx := context.Background()
	key := fmt.Sprintf("%s:%s", s.keyPrefix, identifier)

	state, err := s.getBucketState(ctx, key)
	if err != nil {
		return nil, err
	}

	// Calculate current capacity
	now := time.Now()
	elapsed := now.Sub(state.LastLeak).Seconds()
	leaked := int(elapsed * s.leakRate)
	currentWater := max(0, state.Water-leaked)

	remaining := s.capacity - currentWater
	if remaining < 0 {
		remaining = 0
	}

	// Calculate when the bucket will have space again
	var reset int64
	if currentWater > 0 {
		timeToEmpty := float64(currentWater) / s.leakRate
		reset = now.Unix() + int64(timeToEmpty)
	} else {
		reset = now.Unix()
	}

	return &ratelimit.RateLimitResponse{
		Limit:     s.capacity,
		Remaining: remaining,
		Reset:     reset,
	}, nil
}

// SetRateLimitHeaders sets the rate limit headers in the response
func (s *LeakyBucketStore) SetRateLimitHeaders(c echo.Context, info *ratelimit.RateLimitResponse) {
	c.Response().Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", info.Limit))
	c.Response().Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", info.Remaining))
	c.Response().Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", info.Reset))
}

// ErrorHandler handles internal errors
func (s *LeakyBucketStore) ErrorHandler(c echo.Context, err error) error {
	return c.JSON(500, map[string]string{
		"error": "Internal rate limit error",
	})
}

// DenyHandler handles rate limit exceeded errors
func (s *LeakyBucketStore) DenyHandler(c echo.Context, identifier string, err error) error {
	if err != nil {
		return c.JSON(500, map[string]string{
			"error": "Internal rate limit error",
		})
	}

	info, err := s.GetRateLimitInfo(identifier)
	if err != nil {
		return c.JSON(500, map[string]string{
			"error": "Failed to get rate limit info",
		})
	}

	s.SetRateLimitHeaders(c, info)
	return c.JSON(429, map[string]string{
		"error": "Rate limit exceeded",
	})
}

// NewLeakyBucketMiddleware creates a new leaky bucket rate limiting middleware
func NewLeakyBucketMiddleware(capacity int, leakRate float64, expiresIn time.Duration) echo.MiddlewareFunc {
	store := NewLeakyBucketStore(ratelimit.GetRateLimitRepo(), capacity, leakRate, expiresIn)

	config := middleware.RateLimiterConfig{
		Store: store,
		IdentifierExtractor: func(c echo.Context) (string, error) {
			// Try to get API key first
			apiKey := c.Request().Header.Get("X-API-Key")
			if apiKey != "" {
				return fmt.Sprintf("api:%s", apiKey), nil
			}
			// Fall back to IP address
			return fmt.Sprintf("ip:%s", c.RealIP()), nil
		},
		ErrorHandler: store.ErrorHandler,
		DenyHandler:  store.DenyHandler,
	}

	return middleware.RateLimiterWithConfig(config)
}

// NewLeakyBucketMiddlewarePerPath creates a new leaky bucket rate limiting middleware that's path-specific
func NewLeakyBucketMiddlewarePerPath(capacity int, leakRate float64, expiresIn time.Duration) echo.MiddlewareFunc {
	store := NewLeakyBucketStore(ratelimit.GetRateLimitRepo(), capacity, leakRate, expiresIn)

	config := middleware.RateLimiterConfig{
		Store: store,
		IdentifierExtractor: func(c echo.Context) (string, error) {
			path := c.Request().URL.Path
			// Try to get API key first
			apiKey := c.Request().Header.Get("X-API-Key")
			if apiKey != "" {
				return fmt.Sprintf("api:%s:%s", apiKey, path), nil
			}
			// Fall back to IP address
			return fmt.Sprintf("ip:%s:%s", c.RealIP(), path), nil
		},
		ErrorHandler: store.ErrorHandler,
		DenyHandler:  store.DenyHandler,
	}

	return middleware.RateLimiterWithConfig(config)
}
