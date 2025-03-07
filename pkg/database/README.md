# Database Utilities

This package provides database connection utilities for MongoDB and Redis.

## Features

### MongoDB
- Configurable MongoDB connection settings
- Connection pooling
- Automatic retry logic
- Connection health checks
- Graceful disconnection

### Redis
- Configurable Redis connection settings
- Connection pooling
- Automatic retry logic
- Connection health checks
- Graceful disconnection
- Health statistics monitoring

## Usage

### MongoDB

#### Basic Connection

```go
import (
    "context"
    "log"
    "time"
    "github.com/yourusername/go-echo-mongo/pkg/database"
)

func main() {
    // Use default configuration
    config := database.DefaultConfig()
    config.URI = "mongodb://localhost:27017"
    config.Database = "myapp"

    // Connect to MongoDB
    db, err := database.Connect(config)
    if err != nil {
        log.Fatalf("Failed to connect to MongoDB: %v", err)
    }

    // Ensure disconnection when the application exits
    defer func() {
        if err := database.Disconnect(context.Background()); err != nil {
            log.Printf("Error disconnecting from MongoDB: %v", err)
        }
    }()

    // Use the database
    collection := db.Collection("users")
    // ... perform database operations
}
```

#### Custom Configuration

```go
config := database.Config{
    URI:             "mongodb://localhost:27017",
    Database:        "myapp",
    ConnectTimeout:  15 * time.Second,
    MaxConnIdleTime: 45 * time.Second,
    MaxPoolSize:     200,
    MinPoolSize:     20,
    RetryWrites:     true,
    RetryReads:      true,
    MaxRetries:      5,
}

db, err := database.Connect(config)
if err != nil {
    log.Fatalf("Failed to connect to MongoDB: %v", err)
}
```

#### Health Checks

```go
// Check if the database connection is healthy
ctx := context.Background()
if database.IsConnected(ctx) {
    log.Println("Database connection is healthy")
} else {
    log.Println("Database connection is not healthy")
}
```

### Redis

#### Basic Connection

```go
import (
    "context"
    "log"
    "log/slog"
    "time"
    "github.com/yourusername/go-echo-mongo/pkg/database"
)

func main() {
    // Use default configuration
    config := database.DefaultRedisConfig()
    config.Addr = "localhost:6379"
    config.Password = ""
    config.DB = 0

    // Create Redis service
    redisService, err := database.NewRedisService(config)
    if err != nil {
        log.Fatalf("Failed to connect to Redis: %v", err)
    }

    // Get Redis client
    redisClient := redisService.GetClient()

    // Ensure disconnection when the application exits
    defer func() {
        ctx := context.Background()
        if err := redisService.Disconnect(ctx); err != nil {
            slog.Error("Error disconnecting from Redis", "error", err)
        }
    }()

    // Use the Redis client
    // ... perform Redis operations
}
```

#### Custom Configuration

```go
config := database.RedisConfig{
    Addr:            "localhost:6379",
    Password:        "mypassword",
    DB:              1,
    ConnectTimeout:  15 * time.Second,
    ReadTimeout:     5 * time.Second,
    WriteTimeout:    5 * time.Second,
    PoolSize:        50,
    MinIdleConns:    10,
    MaxRetries:      3,
    MaxRetryBackoff: 2 * time.Second,
}

redisService, err := database.NewRedisService(config)
if err != nil {
    log.Fatalf("Failed to connect to Redis: %v", err)
}
```

#### Health Checks

```go
// Check if the Redis connection is healthy
ctx := context.Background()
if redisService.IsConnected(ctx) {
    slog.Info("Redis connection is healthy")
} else {
    slog.Error("Redis connection is not healthy")
}

// Get detailed health statistics
healthStats := redisService.Health()
for key, value := range healthStats {
    slog.Info("Redis health", key, value)
}
```

## Configuration Options

### MongoDB
- `URI`: MongoDB connection string
- `Database`: Name of the database to connect to
- `ConnectTimeout`: Maximum time to wait for connection
- `MaxConnIdleTime`: Maximum time a connection can remain idle
- `MaxPoolSize`: Maximum number of connections in the pool
- `MinPoolSize`: Minimum number of connections to maintain
- `RetryWrites`: Enable automatic retry of write operations
- `RetryReads`: Enable automatic retry of read operations
- `MaxRetries`: Maximum number of retry attempts for operations

### Redis
- `Addr`: Redis server address (host:port)
- `Password`: Redis server password
- `DB`: Redis database number
- `ConnectTimeout`: Maximum time to wait for connection
- `ReadTimeout`: Timeout for read operations
- `WriteTimeout`: Timeout for write operations
- `PoolSize`: Maximum number of connections in the pool
- `MinIdleConns`: Minimum number of idle connections to maintain
- `MaxRetries`: Maximum number of retry attempts for operations
- `MaxRetryBackoff`: Maximum backoff time between retries
- `DialTimeout`: Timeout for establishing new connections
- `PoolTimeout`: Timeout for getting a connection from the pool
- `ConnMaxIdleTime`: Maximum time a connection can remain idle
- `ConnMaxLifetime`: Maximum lifetime of a connection

## Best Practices

1. Always use the configuration struct to set connection parameters
2. Implement proper error handling for connection failures
3. Use connection pooling appropriately for your application's needs
4. Implement health checks in your application
5. Ensure proper disconnection when shutting down
6. Monitor connection pool metrics in production
7. Use the appropriate Redis repository for your use case (cache, session, rate limit)
8. Set appropriate expiration times for cached data
9. Use tags for efficient cache invalidation
10. Implement rate limiting for public APIs

## Error Handling

The package provides detailed error messages and implements retry logic for:
- Initial connection attempts
- Database ping/health checks
- Read and write operations (when enabled)

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request. 