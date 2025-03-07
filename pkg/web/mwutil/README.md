# Middleware Utilities (mwutil)

This package provides custom middleware utilities for Echo applications.

## Features

- Logger middleware for HTTP request logging
- CORS middleware for Cross-Origin Resource Sharing
- JWT middleware for authentication
- API Key middleware for authentication
- Recovery middleware for panic recovery
- Rate Limiting middleware for request rate limiting

## Usage

### Logger Middleware

```go
import (
    "github.com/yourusername/go-echo-mongo/pkg/web/mwutil"
    "github.com/labstack/echo/v4"
)

func main() {
    e := echo.New()
    
    // Use default logger
    e.Use(mwutil.Logger())
    
    // Or with custom config
    config := mwutil.LoggerConfig{
        Format: "time=${time_rfc3339}, method=${method}, uri=${uri}, status=${status}\n",
    }
    e.Use(mwutil.LoggerWithConfig(config))
}
```

### CORS Middleware

```go
import (
    "github.com/yourusername/go-echo-mongo/pkg/web/mwutil"
    "github.com/labstack/echo/v4"
)

func main() {
    e := echo.New()
    
    // Use default CORS settings
    e.Use(mwutil.CORS())
    
    // Or with custom config
    config := mwutil.CORSConfig{
        AllowOrigins: []string{"https://example.com", "https://api.example.com"},
        AllowMethods: []string{echo.GET, echo.PUT, echo.POST, echo.DELETE},
    }
    e.Use(mwutil.CORSWithConfig(config))
}
```

### JWT Middleware

```go
import (
    "github.com/yourusername/go-echo-mongo/pkg/web/mwutil"
    "github.com/labstack/echo/v4"
)

func main() {
    e := echo.New()
    
    // Use JWT middleware with a secret key
    e.Use(mwutil.JWT("your-secret-key"))
    
    // Or with custom config
    config := mwutil.JWTConfig{
        Secret:      "your-secret-key",
        TokenLookup: "cookie:token",
    }
    e.Use(mwutil.JWTWithConfig(config))
    
    // Access the user in your handlers
    e.GET("/protected", func(c echo.Context) error {
        user := c.Get("user").(map[string]interface{})
        return c.JSON(200, user)
    })
}
```

### API Key Middleware

```go
// Implement the APIKeyValidator interface in your service
type UserService struct {
    // ... your service fields
}

func (s *UserService) GetByApiKey(ctx context.Context, apiKey string) (*model.User, error) {
    // Validate API key and return user
    return s.repo.FindByApiKey(ctx, apiKey)
}

// Use the middleware
func main() {
    e := echo.New()
    userService := NewUserService()
    
    // Set the global validator
    mwutil.SetAPIKeyValidator(userService)

    // Use default configuration with global validator
    e.Use(mwutil.NewAPIKeyAuth())

    // Or with custom configuration
    config := mwutil.APIKeyAuthConfig{
        KeyLookup:  "header:X-API-Key",  // Can also use "query:api_key"
        ContextKey: "user",
        Validator:  userService,  // Override global validator if needed
        ErrorHandler: func(c echo.Context, err error) error {
            return c.JSON(http.StatusUnauthorized, map[string]string{
                "error": "Invalid API key",
            })
        },
    }
    e.Use(mwutil.NewAPIKeyAuthWithConfig(config))
}

// Access user in handlers
func handler(c echo.Context) error {
    user := c.Get("user").(*model.User)
    // ... use user object
}
```

The API Key middleware:
- Requires a global validator to be set using `SetAPIKeyValidator`
- Extracts API key from request header or query parameter
- Validates the API key against a database using the validator
- Stores the user object in the context if validation succeeds
- Returns 401 Unauthorized if validation fails

### Rate Limiting Middleware

```go
import (
    "time"
    "github.com/yourusername/go-echo-mongo/pkg/web/mwutil"
    "github.com/labstack/echo/v4"
)

func main() {
    e := echo.New()
    
    // Set the global rate limit repository
    repo := NewRedisRateLimitRepo()  // Your implementation
    ratelimit.SetRateLimitRepo(repo)

    // Use different rate limiting strategies:

    // 1. Fixed Window - 100 requests per minute
    e.Use(mwutil.NewFixedRateLimiter(100, time.Minute))
    
    // 2. Sliding Window - 100 requests per minute
    e.Use(mwutil.NewSlidingRateLimiter(100, time.Minute))
    
    // 3. Token Bucket - 10 tokens per second, burst of 100
    e.Use(mwutil.NewTokenBucketLimiter(10, 100, time.Hour))
    
    // 4. Leaky Bucket - capacity 100, leak rate 10 per second
    e.Use(mwutil.NewLeakyBucketLimiter(100, 10, time.Hour))

    // Path-specific rate limiting:
    
    // 1. Fixed Window per path
    e.Use(mwutil.NewFixedRateLimiterPerPath(100, time.Minute))
    
    // 2. Sliding Window per path
    e.Use(mwutil.NewSlidingRateLimiterPerPath(100, time.Minute))
    
    // 3. Token Bucket per path
    e.Use(mwutil.NewTokenBucketLimiterPerPath(10, 100, time.Hour))
    
    // 4. Leaky Bucket per path
    e.Use(mwutil.NewLeakyBucketLimiterPerPath(100, 10, time.Hour))

    // Or use configuration-based approach
    config := mwutil.RateLimitConfig{
        Strategy: mwutil.FixedWindow,
        Limit:    100,
        Window:   time.Minute,
        // For Token/Leaky Bucket:
        Burst:    100,  // Used by Token Bucket
        Rate:     10,   // Tokens per second or leak rate
    }
    e.Use(mwutil.NewRateLimiter(config))
}
```

The Rate Limiting middleware:
- Requires a global repository to be set using `SetRateLimitRepo`
- Supports four rate limiting strategies: Fixed Window, Sliding Window, Token Bucket, and Leaky Bucket
- Provides both global and path-specific rate limiting
- Uses API key for identification if present, falls back to IP address
- Sets rate limit headers (X-RateLimit-Limit, X-RateLimit-Remaining, X-RateLimit-Reset)
- Returns 429 Too Many Requests when limit is exceeded

### Recovery Middleware

```go
import (
    "github.com/yourusername/go-echo-mongo/pkg/web/mwutil"
    "github.com/labstack/echo/v4"
)

func main() {
    e := echo.New()
    
    // Use default recovery middleware
    e.Use(mwutil.Recovery())
    
    // Or with custom config
    config := mwutil.RecoveryConfig{
        LogErrorFunc: func(c echo.Context, err error, stack []byte) {
            // Custom logging logic
            fmt.Printf("Error: %v\n", err)
            fmt.Printf("Stack: %s\n", stack)
        },
    }
    e.Use(mwutil.RecoveryWithConfig(config))
    
    // Or use the custom implementation
    e.Use(mwutil.CustomRecovery())
} 