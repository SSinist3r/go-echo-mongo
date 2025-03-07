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

// TokenBucketStore implements a true token bucket algorithm
type TokenBucketStore struct {
	repo      ratelimit.RateLimitRepo
	rate      float64 // tokens per second
	burst     int     // maximum bucket size
	expiresIn time.Duration
	keyPrefix string // Key prefix for rate limit
}

type tokenBucketState struct {
	Tokens     float64   `json:"tokens"`
	LastRefill time.Time `json:"last_refill"`
}

// NewTokenBucketStore creates a new token bucket rate limiter
func NewTokenBucketStore(repo ratelimit.RateLimitRepo, rate float64, burst int, expiresIn time.Duration) *TokenBucketStore {
	return &TokenBucketStore{
		repo:      repo,
		rate:      rate,
		burst:     burst,
		expiresIn: expiresIn,
		keyPrefix: "rate_limit_token_bucket",
	}
}

// Allow implements the RateLimiterStore interface
func (s *TokenBucketStore) Allow(identifier string) (bool, error) {
	ctx := context.Background()
	key := fmt.Sprintf("%s:%s", s.keyPrefix, identifier)

	// Get or initialize bucket state
	state, err := s.getBucketState(ctx, key)
	if err != nil {
		return false, fmt.Errorf("failed to get bucket state: %w", err)
	}

	// Calculate token refill
	now := time.Now()
	elapsed := now.Sub(state.LastRefill).Seconds()
	newTokens := state.Tokens + (elapsed * s.rate)

	// Cap tokens at burst size
	if newTokens > float64(s.burst) {
		newTokens = float64(s.burst)
	}

	// Check if we have enough tokens
	if newTokens < 1 {
		return false, nil
	}

	// Update bucket state
	state.Tokens = newTokens - 1
	state.LastRefill = now

	// Save updated state
	if err := s.saveBucketState(ctx, key, state); err != nil {
		return false, fmt.Errorf("failed to save bucket state: %w", err)
	}

	return true, nil
}

func (s *TokenBucketStore) getBucketState(ctx context.Context, key string) (*tokenBucketState, error) {
	stateJSON, err := s.repo.GetState(ctx, key)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return &tokenBucketState{
				Tokens:     float64(s.burst),
				LastRefill: time.Now(),
			}, nil
		}
		return nil, err
	}

	var state tokenBucketState
	if err := json.Unmarshal([]byte(stateJSON), &state); err != nil {
		return nil, fmt.Errorf("failed to unmarshal token bucket state: %w", err)
	}
	return &state, nil
}

func (s *TokenBucketStore) saveBucketState(ctx context.Context, key string, state *tokenBucketState) error {
	stateJSON, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("failed to marshal token bucket state: %w", err)
	}
	return s.repo.SetState(ctx, key, string(stateJSON), s.expiresIn)
}

// GetRateLimitInfo returns information about the current rate limit state
func (s *TokenBucketStore) GetRateLimitInfo(identifier string) (*ratelimit.RateLimitResponse, error) {
	ctx := context.Background()
	key := fmt.Sprintf("%s:%s", s.keyPrefix, identifier)

	state, err := s.getBucketState(ctx, key)
	if err != nil {
		return nil, err
	}

	// Calculate current tokens
	now := time.Now()
	elapsed := now.Sub(state.LastRefill).Seconds()
	currentTokens := state.Tokens + (elapsed * s.rate)
	if currentTokens > float64(s.burst) {
		currentTokens = float64(s.burst)
	}

	// Calculate when tokens will be fully replenished
	tokensNeeded := float64(s.burst) - currentTokens
	timeToFull := time.Duration(tokensNeeded/s.rate) * time.Second

	return &ratelimit.RateLimitResponse{
		Limit:     s.burst,
		Remaining: int(currentTokens),
		Reset:     now.Add(timeToFull).Unix(),
	}, nil
}

// SetRateLimitHeaders sets the rate limit headers in the response
func (s *TokenBucketStore) SetRateLimitHeaders(c echo.Context, info *ratelimit.RateLimitResponse) {
	c.Response().Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", info.Limit))
	c.Response().Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", info.Remaining))
	c.Response().Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", info.Reset))
}

// ErrorHandler handles internal errors
func (s *TokenBucketStore) ErrorHandler(c echo.Context, err error) error {
	return c.JSON(500, map[string]string{
		"error": "Internal rate limit error",
	})
}

// DenyHandler handles rate limit exceeded errors
func (s *TokenBucketStore) DenyHandler(c echo.Context, identifier string, err error) error {
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

// NewTokenBucketMiddleware creates a new token bucket rate limiting middleware
func NewTokenBucketMiddleware(rate float64, burst int, expiresIn time.Duration) echo.MiddlewareFunc {
	store := NewTokenBucketStore(ratelimit.GetRateLimitRepo(), rate, burst, expiresIn)

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

// NewTokenBucketMiddlewarePerPath creates a new token bucket rate limiting middleware that's path-specific
func NewTokenBucketMiddlewarePerPath(rate float64, burst int, expiresIn time.Duration) echo.MiddlewareFunc {
	store := NewTokenBucketStore(ratelimit.GetRateLimitRepo(), rate, burst, expiresIn)

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
