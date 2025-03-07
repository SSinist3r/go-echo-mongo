package strategy

import (
	"context"
	"fmt"
	"time"

	"go-echo-mongo/pkg/ratelimit"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// SlidingWindowStore implements a sliding window rate limiter
type SlidingWindowStore struct {
	repo       ratelimit.RateLimitRepo
	limit      int           // Maximum requests per window
	windowSize time.Duration // Time window size
	keyPrefix  string        // Key prefix for rate limit
}

// NewSlidingWindowStore creates a new sliding window rate limiter
func NewSlidingWindowStore(repo ratelimit.RateLimitRepo, limit int, windowSize time.Duration) *SlidingWindowStore {
	return &SlidingWindowStore{
		repo:       repo,
		limit:      limit,
		windowSize: windowSize,
		keyPrefix:  "rate_limit_sliding_window",
	}
}

// Allow implements the RateLimiterStore interface
func (s *SlidingWindowStore) Allow(identifier string) (bool, error) {
	ctx := context.Background()
	now := time.Now()

	// Get the current and previous window numbers
	currentWindow := now.Unix() / int64(s.windowSize.Seconds())
	previousWindow := currentWindow - 1

	// Create keys for current and previous windows
	currentKey := fmt.Sprintf("%s:%s:%d", s.keyPrefix, identifier, currentWindow)
	previousKey := fmt.Sprintf("%s:%s:%d", s.keyPrefix, identifier, previousWindow)

	// Get counts for both windows
	currentCount, err := s.repo.Check(ctx, currentKey)
	if err != nil {
		return false, err
	}

	previousCount, err := s.repo.Check(ctx, previousKey)
	if err != nil {
		return false, err
	}

	// Calculate the weight of the previous window
	// This represents how much of the previous window should be counted
	offset := float64(now.Unix()%int64(s.windowSize.Seconds())) / float64(s.windowSize.Seconds())
	previousWeight := 1 - offset

	// Calculate the weighted sum of requests
	weightedCount := int(float64(previousCount)*previousWeight) + currentCount

	// Check if adding this request would exceed the limit
	if weightedCount >= s.limit {
		return false, nil
	}

	// If we're still under the limit, increment the current window
	_, err = s.repo.IncrementPreserveTTL(ctx, currentKey, s.windowSize*2)
	if err != nil {
		return false, err
	}

	return true, nil
}

// GetRateLimitInfo returns information about the current rate limit state
func (s *SlidingWindowStore) GetRateLimitInfo(identifier string) (*ratelimit.RateLimitResponse, error) {
	ctx := context.Background()
	now := time.Now()

	currentWindow := now.Unix() / int64(s.windowSize.Seconds())
	previousWindow := currentWindow - 1

	currentKey := fmt.Sprintf("%s:%s:%d", s.keyPrefix, identifier, currentWindow)
	previousKey := fmt.Sprintf("%s:%s:%d", s.keyPrefix, identifier, previousWindow)

	currentCount, err := s.repo.Check(ctx, currentKey)
	if err != nil {
		return nil, err
	}

	previousCount, err := s.repo.Check(ctx, previousKey)
	if err != nil {
		return nil, err
	}

	offset := float64(now.Unix()%int64(s.windowSize.Seconds())) / float64(s.windowSize.Seconds())
	previousWeight := 1 - offset

	weightedCount := int(float64(previousCount)*previousWeight) + currentCount
	remaining := s.limit - weightedCount
	if remaining < 0 {
		remaining = 0
	}

	// Calculate when the current window ends
	nextReset := (currentWindow + 1) * int64(s.windowSize.Seconds())

	return &ratelimit.RateLimitResponse{
		Limit:     s.limit,
		Remaining: remaining,
		Reset:     nextReset,
	}, nil
}

// SetRateLimitHeaders sets the rate limit headers in the response
func (s *SlidingWindowStore) SetRateLimitHeaders(c echo.Context, info *ratelimit.RateLimitResponse) {
	c.Response().Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", info.Limit))
	c.Response().Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", info.Remaining))
	c.Response().Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", info.Reset))
}

// ErrorHandler handles internal errors
func (s *SlidingWindowStore) ErrorHandler(c echo.Context, err error) error {
	return c.JSON(500, map[string]string{
		"error": "Internal rate limit error",
	})
}

// DenyHandler handles rate limit exceeded errors
func (s *SlidingWindowStore) DenyHandler(c echo.Context, identifier string, err error) error {
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

// NewSlidingWindowMiddleware creates a new sliding window rate limiting middleware
func NewSlidingWindowMiddleware(limit int, windowSize time.Duration) echo.MiddlewareFunc {
	store := NewSlidingWindowStore(ratelimit.GetRateLimitRepo(), limit, windowSize)

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

// NewSlidingWindowMiddlewarePerPath creates a new sliding window rate limiting middleware that's path-specific
func NewSlidingWindowMiddlewarePerPath(limit int, windowSize time.Duration) echo.MiddlewareFunc {
	store := NewSlidingWindowStore(ratelimit.GetRateLimitRepo(), limit, windowSize)

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
