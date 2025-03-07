package mwutil

import (
	"log"
	"time"

	"go-echo-mongo/pkg/ratelimit"
	"go-echo-mongo/pkg/ratelimit/strategy"

	"github.com/labstack/echo/v4"
)

// RateLimitStrategy represents the type of rate limiting algorithm to use
type RateLimitStrategy string

const (
	// FixedWindow represents a fixed window rate limiting strategy
	FixedWindow RateLimitStrategy = "fixed_window"
	// SlidingWindow represents a sliding window rate limiting strategy
	SlidingWindow RateLimitStrategy = "sliding_window"
	// TokenBucket represents a token bucket rate limiting strategy
	TokenBucket RateLimitStrategy = "token_bucket"
	// LeakyBucket represents a leaky bucket rate limiting strategy
	LeakyBucket RateLimitStrategy = "leaky_bucket"
)

// RateLimitConfig holds the configuration for rate limiting
type RateLimitConfig struct {
	// Strategy is the rate limiting strategy to use
	Strategy RateLimitStrategy
	// Limit is the maximum number of requests allowed per window
	Limit int
	// Window is the time window for rate limiting
	Window time.Duration
	// Burst is the maximum burst size (only used for token bucket)
	Burst int
	// Rate is the rate at which tokens are added (token bucket) or water leaks (leaky bucket)
	Rate float64
}

// NewRateLimiter creates a new rate limiting middleware based on the provided strategy
func NewRateLimiter(config RateLimitConfig) echo.MiddlewareFunc {
	if ratelimit.GetRateLimitRepo() == nil {
		log.Fatal("echo: rate limit repository is not set")
	}
	switch config.Strategy {
	case FixedWindow:
		return strategy.NewFixedWindowMiddleware(config.Limit, config.Window)
	case SlidingWindow:
		return strategy.NewSlidingWindowMiddleware(config.Limit, config.Window)
	case TokenBucket:
		return strategy.NewTokenBucketMiddleware(config.Rate, config.Burst, config.Window)
	case LeakyBucket:
		return strategy.NewLeakyBucketMiddleware(config.Burst, config.Rate, config.Window)
	default:
		// Default to fixed window if strategy is not recognized
		return strategy.NewFixedWindowMiddleware(config.Limit, config.Window)
	}
}

// NewFixedRateLimiter creates a new fixed window rate limiter
// limit: maximum number of requests per window
// window: time window for rate limiting
func NewFixedRateLimiter(limit int, window time.Duration) echo.MiddlewareFunc {
	if ratelimit.GetRateLimitRepo() == nil {
		log.Fatal("echo: rate limit repository is not set")
	}
	return strategy.NewFixedWindowMiddleware(limit, window)
}

// NewSlidingRateLimiter creates a new sliding window rate limiter
// limit: maximum number of requests per window
// window: time window for rate limiting
func NewSlidingRateLimiter(limit int, window time.Duration) echo.MiddlewareFunc {
	if ratelimit.GetRateLimitRepo() == nil {
		log.Fatal("echo: rate limit repository is not set")
	}
	return strategy.NewSlidingWindowMiddleware(limit, window)
}

// NewTokenBucketLimiter creates a new token bucket rate limiter
// rate: tokens per second
// burst: maximum bucket size
// window: expiration time for bucket state
func NewTokenBucketLimiter(rate float64, burst int, window time.Duration) echo.MiddlewareFunc {
	if ratelimit.GetRateLimitRepo() == nil {
		log.Fatal("echo: rate limit repository is not set")
	}
	return strategy.NewTokenBucketMiddleware(rate, burst, window)
}

// NewLeakyBucketLimiter creates a new leaky bucket rate limiter
// capacity: maximum bucket capacity
// leakRate: requests per second that leak out
// window: expiration time for bucket state
func NewLeakyBucketLimiter(capacity int, leakRate float64, window time.Duration) echo.MiddlewareFunc {
	if ratelimit.GetRateLimitRepo() == nil {
		log.Fatal("echo: rate limit repository is not set")
	}
	return strategy.NewLeakyBucketMiddleware(capacity, leakRate, window)
}

// NewFixedRateLimiterPerPath creates a fixed window rate limiter that's path-specific
// limit: maximum number of requests per window
// window: time window for rate limiting
func NewFixedRateLimiterPerPath(limit int, window time.Duration) echo.MiddlewareFunc {
	if ratelimit.GetRateLimitRepo() == nil {
		log.Fatal("echo: rate limit repository is not set")
	}
	return strategy.NewFixedWindowMiddlewarePerPath(limit, window)
}

// NewSlidingRateLimiterPerPath creates a sliding window rate limiter that's path-specific
// limit: maximum number of requests per window
// window: time window for rate limiting
func NewSlidingRateLimiterPerPath(limit int, window time.Duration) echo.MiddlewareFunc {
	if ratelimit.GetRateLimitRepo() == nil {
		log.Fatal("echo: rate limit repository is not set")
	}
	return strategy.NewSlidingWindowMiddlewarePerPath(limit, window)
}

// NewTokenBucketLimiterPerPath creates a token bucket rate limiter that's path-specific
// rate: tokens per second
// burst: maximum bucket size
// window: expiration time for bucket state
func NewTokenBucketLimiterPerPath(rate float64, burst int, window time.Duration) echo.MiddlewareFunc {
	if ratelimit.GetRateLimitRepo() == nil {
		log.Fatal("echo: rate limit repository is not set")
	}
	return strategy.NewTokenBucketMiddlewarePerPath(rate, burst, window)
}

// NewLeakyBucketLimiterPerPath creates a leaky bucket rate limiter that's path-specific
// capacity: maximum bucket capacity
// leakRate: requests per second that leak out
// window: expiration time for bucket state
func NewLeakyBucketLimiterPerPath(capacity int, leakRate float64, window time.Duration) echo.MiddlewareFunc {
	if ratelimit.GetRateLimitRepo() == nil {
		log.Fatal("echo: rate limit repository is not set")
	}
	return strategy.NewLeakyBucketMiddlewarePerPath(capacity, leakRate, window)
}
