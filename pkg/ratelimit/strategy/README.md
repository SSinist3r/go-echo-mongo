# Rate Limiting Strategies

This package provides implementations of various rate limiting algorithms for use with Echo's rate limiter middleware.

## Strategies

### Fixed Window (`fixed_window.go`)

The fixed window algorithm divides time into fixed windows (e.g., 1 minute) and limits the number of requests in each window.

```go
// Create a fixed window rate limiter with 100 requests per minute
store := strategy.NewFixedWindowStore(repo, 100, 1*time.Minute)
```

### Sliding Window (`sliding_window.go`)

The sliding window algorithm uses a rolling time window to limit requests, providing smoother rate limiting than fixed windows.

```go
// Create a sliding window rate limiter with 100 requests per minute
store := strategy.NewSlidingWindowStore(repo, 100, 1*time.Minute)
```

### Token Bucket (`token_bucket.go`)

The token bucket algorithm uses a bucket that fills with tokens at a constant rate. Each request consumes a token.

```go
// Create a token bucket rate limiter with 10 tokens per second and a burst of 30
store := strategy.NewTokenBucketStore(repo, 10, 30, 1*time.Hour)
```

### Leaky Bucket (`leaky_bucket.go`)

The leaky bucket algorithm processes requests at a constant rate, with excess requests either queued or discarded.

```go
// Create a leaky bucket rate limiter with 10 requests per second
store := strategy.NewLeakyBucketStore(repo, 10, 1*time.Hour)
```

## Usage with Echo

All strategies implement Echo's `middleware.RateLimiterStore` interface and can be used with the rate limiter middleware:

```go
import (
    "github.com/yourusername/go-echo-mongo/pkg/ratelimit"
    "github.com/yourusername/go-echo-mongo/pkg/ratelimit/strategy"
    "github.com/labstack/echo/v4"
    "github.com/labstack/echo/v4/middleware"
)

// Create a new Echo instance
e := echo.New()

// Get the rate limit repository
repo := ratelimit.GetRateLimitRepo()

// Create a rate limiter store
store := strategy.NewTokenBucketStore(repo, 10, 30, 1*time.Hour)

// Apply rate limiting middleware
e.Use(middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{
    Store: store,
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

## Choosing a Strategy

- **Fixed Window**: Simple to understand and implement, but can lead to request spikes at window boundaries.
- **Sliding Window**: More even distribution of requests, but slightly more complex and resource-intensive.
- **Token Bucket**: Good for APIs with burst traffic patterns, allowing temporary spikes while maintaining a long-term rate.
- **Leaky Bucket**: Good for APIs that need a constant processing rate, smoothing out traffic spikes.

## Implementation Details

All strategies use a repository interface for storage, allowing different backends (Redis, in-memory, etc.) to be used. 