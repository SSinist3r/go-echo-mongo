package strategy

import (
	"context"
	"fmt"
	"time"

	"go-echo-mongo/pkg/ratelimit"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// FixedWindowStore implements a simple fixed-window rate limiter
type FixedWindowStore struct {
	repo       ratelimit.RateLimitRepo
	limit      int           // Maximum requests per window
	windowSize time.Duration // Time window size
	keyPrefix  string        // Key prefix for rate limit
}

// NewFixedWindowStore creates a new fixed window rate limiter
func NewFixedWindowStore(repo ratelimit.RateLimitRepo, limit int, windowSize time.Duration) *FixedWindowStore {
	return &FixedWindowStore{
		repo:       repo,
		limit:      limit,
		windowSize: windowSize,
		keyPrefix:  "rate_limit_fixed_window",
	}
}

// Allow implements the RateLimiterStore interface
func (s *FixedWindowStore) Allow(identifier string) (bool, error) {
	ctx := context.Background()

	// Create a key that includes the current time window
	windowNum := time.Now().Unix() / int64(s.windowSize.Seconds())
	key := fmt.Sprintf("%s:%s:%d", s.keyPrefix, identifier, windowNum)

	// Increment the counter for this window
	count, err := s.repo.IncrementPreserveTTL(ctx, key, s.windowSize)
	if err != nil {
		return false, fmt.Errorf("failed to increment rate limit counter: %w", err)
	}

	return count <= s.limit, nil
}

// GetRateLimitInfo returns information about the current rate limit state
func (s *FixedWindowStore) GetRateLimitInfo(identifier string) (*ratelimit.RateLimitResponse, error) {
	ctx := context.Background()
	now := time.Now()
	windowNum := now.Unix() / int64(s.windowSize.Seconds())
	key := fmt.Sprintf("%s:%s:%d", s.keyPrefix, identifier, windowNum)

	count, err := s.repo.Check(ctx, key)
	if err != nil {
		return nil, err
	}

	nextWindow := (windowNum + 1) * int64(s.windowSize.Seconds())
	remaining := s.limit - count
	if remaining < 0 {
		remaining = 0
	}

	return &ratelimit.RateLimitResponse{
		Limit:     s.limit,
		Remaining: remaining,
		Reset:     nextWindow,
	}, nil
}

// SetRateLimitHeaders sets the rate limit headers in the response
func (s *FixedWindowStore) SetRateLimitHeaders(c echo.Context, info *ratelimit.RateLimitResponse) {
	c.Response().Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", info.Limit))
	c.Response().Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", info.Remaining))
	c.Response().Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", info.Reset))
}

// ErrorHandler handles internal errors
func (s *FixedWindowStore) ErrorHandler(c echo.Context, err error) error {
	return c.JSON(500, map[string]string{
		"error": "Internal rate limit error",
	})
}

// DenyHandler handles rate limit exceeded errors
func (s *FixedWindowStore) DenyHandler(c echo.Context, identifier string, err error) error {
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

// NewFixedWindowMiddleware creates a new fixed window rate limiting middleware
func NewFixedWindowMiddleware(limit int, windowSize time.Duration) echo.MiddlewareFunc {
	store := NewFixedWindowStore(ratelimit.GetRateLimitRepo(), limit, windowSize)

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

// NewFixedWindowMiddlewarePerPath creates a new fixed window rate limiting middleware that's path-specific
func NewFixedWindowMiddlewarePerPath(limit int, windowSize time.Duration) echo.MiddlewareFunc {
	store := NewFixedWindowStore(ratelimit.GetRateLimitRepo(), limit, windowSize)

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
