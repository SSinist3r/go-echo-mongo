# Rate Limiting Utilities

This package provides rate limiting utilities for controlling the rate of requests to your API.

## Features

- Repository interface for rate limiting operations
- Multiple rate limiting strategies:
  - Fixed Window
  - Sliding Window
  - Token Bucket
  - Leaky Bucket
- Configurable rate limits and time windows
- Support for different storage backends

## Packages

- **strategy**: Implementation of various rate limiting algorithms

## Usage

### Setting Up a Rate Limiter Repository

```go
import (
    "context"
    "time"
    "github.com/yourusername/go-echo-mongo/pkg/ratelimit"
)

// Set up a rate limit repository implementation
repo := NewRedisRateLimitRepo(redisClient)
ratelimit.SetRateLimitRepo(repo)

// Get the repository for use in rate limiters
repo := ratelimit.GetRateLimitRepo()
```

### Using Rate Limiting Strategies

```go
import (
    "github.com/yourusername/go-echo-mongo/pkg/ratelimit/strategy"
    "github.com/labstack/echo/v4"
    "github.com/labstack/echo/v4/middleware"
)

// Create a new Echo instance
e := echo.New()

// Set up rate limiter middleware with token bucket strategy
repo := ratelimit.GetRateLimitRepo()
tokenBucket := strategy.NewTokenBucketStore(repo, 10, 30, 1*time.Hour)

// Apply rate limiting middleware to all routes
e.Use(middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{
    Store: tokenBucket,
    IdentifierExtractor: func(c echo.Context) (string, error) {
        return c.RealIP(), nil
    },
    ErrorHandler: func(c echo.Context, err error) error {
        return c.JSON(429, map[string]string{
            "error": "Too many requests",
        })
    },
}))
```

## Rate Limiting Strategies

### Fixed Window

The fixed window algorithm divides time into fixed windows and limits the number of requests in each window.

### Sliding Window

The sliding window algorithm uses a rolling time window to limit requests, providing smoother rate limiting.

### Token Bucket

The token bucket algorithm uses a bucket that fills with tokens at a constant rate. Each request consumes a token.

### Leaky Bucket

The leaky bucket algorithm processes requests at a constant rate, with excess requests either queued or discarded.

## Best Practices

1. Choose the appropriate rate limiting strategy for your use case
2. Set reasonable rate limits based on your API's capacity
3. Use different rate limits for different endpoints or user tiers
4. Provide clear rate limit information in response headers
5. Handle rate limit errors gracefully on the client side 