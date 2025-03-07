package database

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisConfig holds Redis connection configuration
type RedisConfig struct {
	Addr            string
	Password        string
	DB              int
	ConnectTimeout  time.Duration
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	PoolSize        int
	MinIdleConns    int
	MaxRetries      int
	MaxRetryBackoff time.Duration
	DialTimeout     time.Duration
	PoolTimeout     time.Duration
	ConnMaxIdleTime time.Duration
	ConnMaxLifetime time.Duration
}

// DefaultRedisConfig returns a default Redis configuration
func DefaultRedisConfig() RedisConfig {
	return RedisConfig{
		Addr:            "localhost:6379",
		Password:        "",
		DB:              0,
		ConnectTimeout:  10 * time.Second,
		ReadTimeout:     3 * time.Second,
		WriteTimeout:    3 * time.Second,
		PoolSize:        10,
		MinIdleConns:    5,
		MaxRetries:      3,
		MaxRetryBackoff: time.Second,
		DialTimeout:     5 * time.Second,
		PoolTimeout:     4 * time.Second,
		ConnMaxIdleTime: 5 * time.Minute,
		ConnMaxLifetime: 30 * time.Minute,
	}
}

// RedisService defines the interface for Redis operations
type RedisService interface {
	GetClient() *redis.Client
	Health() map[string]string
	IsConnected(ctx context.Context) bool
	Disconnect(ctx context.Context) error
}

// redisService implements the RedisService interface
type redisService struct {
	client *redis.Client
}

// NewRedisService creates a new Redis service instance
func NewRedisService(config RedisConfig) (RedisService, error) {
	client, err := connectRedis(config)
	if err != nil {
		return nil, err
	}

	return &redisService{
		client: client,
	}, nil
}

// GetClient returns the Redis client
func (r *redisService) GetClient() *redis.Client {
	return r.client
}

// IsConnected checks if the Redis client is connected and healthy
func (r *redisService) IsConnected(ctx context.Context) bool {
	if r.client == nil {
		return false
	}
	_, err := r.client.Ping(ctx).Result()
	return err == nil
}

// Disconnect closes the Redis connection
func (r *redisService) Disconnect(ctx context.Context) error {
	if r.client == nil {
		return nil
	}

	if err := r.client.Close(); err != nil {
		return fmt.Errorf("failed to disconnect from Redis: %w", err)
	}

	slog.Info("Successfully disconnected from Redis")
	return nil
}

// connectRedis establishes connection to Redis and returns the client instance
func connectRedis(config RedisConfig) (*redis.Client, error) {
	slog.Info("Attempting to connect to Redis")

	// Set up context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), config.ConnectTimeout)
	defer cancel()

	// Create Redis client options
	options := &redis.Options{
		Addr:            config.Addr,
		Password:        config.Password,
		DB:              config.DB,
		DialTimeout:     config.DialTimeout,
		ReadTimeout:     config.ReadTimeout,
		WriteTimeout:    config.WriteTimeout,
		PoolSize:        config.PoolSize,
		MinIdleConns:    config.MinIdleConns,
		MaxRetries:      config.MaxRetries,
		MaxRetryBackoff: config.MaxRetryBackoff,
		PoolTimeout:     config.PoolTimeout,
		ConnMaxIdleTime: config.ConnMaxIdleTime,
		ConnMaxLifetime: config.ConnMaxLifetime,
	}

	// Connect to Redis
	client := redis.NewClient(options)

	// Ping Redis with retry logic
	var err error
	for i := 0; i <= config.MaxRetries; i++ {
		_, err = client.Ping(ctx).Result()
		if err == nil {
			break
		}
		if i < config.MaxRetries {
			slog.Warn("Failed to ping Redis", "attempt", i+1, "maxRetries", config.MaxRetries, "error", err)
			time.Sleep(time.Duration(i+1) * time.Second) // Exponential backoff
		}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to ping Redis after %d attempts: %w", config.MaxRetries, err)
	}

	slog.Info("Successfully connected to Redis")
	return client, nil
}

// Health returns the health status and statistics of the Redis server.
func (s *redisService) Health() map[string]string {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second) // Default is now 5s
	defer cancel()

	stats := make(map[string]string)

	// Check Redis health and populate the stats map
	stats = s.checkRedisHealth(ctx, stats)

	return stats
}

// checkRedisHealth checks the health of the Redis server and adds the relevant statistics to the stats map.
func (s *redisService) checkRedisHealth(ctx context.Context, stats map[string]string) map[string]string {
	// Ping the Redis server to check its availability.
	pong, err := s.client.Ping(ctx).Result()
	// Note: By extracting and simplifying like this, `log.Fatalf("db down: %v", err)`
	// can be changed into a standard error instead of a fatal error.
	if err != nil {
		log.Fatalf("db down: %v", err)
	}

	// Redis is up
	stats["redis_status"] = "up"
	stats["redis_message"] = "It's healthy"
	stats["redis_ping_response"] = pong

	// Retrieve Redis server information.
	info, err := s.client.Info(ctx).Result()
	if err != nil {
		stats["redis_message"] = fmt.Sprintf("Failed to retrieve Redis info: %v", err)
		return stats
	}

	// Parse the Redis info response.
	redisInfo := parseRedisInfo(info)

	// Get the pool stats of the Redis client.
	poolStats := s.client.PoolStats()

	// Prepare the stats map with Redis server information and pool statistics.
	// Note: The "stats" map in the code uses string keys and values,
	// which is suitable for structuring and serializing the data for the frontend (e.g., JSON, XML, HTMX).
	// Using string types allows for easy conversion and compatibility with various data formats,
	// making it convenient to create health stats for monitoring or other purposes.
	// Also note that any raw "memory" (e.g., used_memory) value here is in bytes and can be converted to megabytes or gigabytes as a float64.
	stats["redis_version"] = redisInfo["redis_version"]
	stats["redis_mode"] = redisInfo["redis_mode"]
	stats["redis_connected_clients"] = redisInfo["connected_clients"]
	stats["redis_used_memory"] = redisInfo["used_memory"]
	stats["redis_used_memory_peak"] = redisInfo["used_memory_peak"]
	stats["redis_uptime_in_seconds"] = redisInfo["uptime_in_seconds"]
	stats["redis_hits_connections"] = strconv.FormatUint(uint64(poolStats.Hits), 10)
	stats["redis_misses_connections"] = strconv.FormatUint(uint64(poolStats.Misses), 10)
	stats["redis_timeouts_connections"] = strconv.FormatUint(uint64(poolStats.Timeouts), 10)
	stats["redis_total_connections"] = strconv.FormatUint(uint64(poolStats.TotalConns), 10)
	stats["redis_idle_connections"] = strconv.FormatUint(uint64(poolStats.IdleConns), 10)
	stats["redis_stale_connections"] = strconv.FormatUint(uint64(poolStats.StaleConns), 10)
	stats["redis_max_memory"] = redisInfo["maxmemory"]

	// Calculate the number of active connections.
	// Note: We use math.Max to ensure that activeConns is always non-negative,
	// avoiding the need for an explicit check for negative values.
	// This prevents a potential underflow situation.
	activeConns := uint64(math.Max(float64(poolStats.TotalConns-poolStats.IdleConns), 0))
	stats["redis_active_connections"] = strconv.FormatUint(activeConns, 10)

	// Calculate the pool size percentage.
	poolSize := s.client.Options().PoolSize
	connectedClients, _ := strconv.Atoi(redisInfo["connected_clients"])
	poolSizePercentage := float64(connectedClients) / float64(poolSize) * 100
	stats["redis_pool_size_percentage"] = fmt.Sprintf("%.2f%%", poolSizePercentage)

	// Evaluate Redis stats and update the stats map with relevant messages.
	return s.evaluateRedisStats(redisInfo, stats)
}

// evaluateRedisStats evaluates the Redis server statistics and updates the stats map with relevant messages.
func (s *redisService) evaluateRedisStats(redisInfo, stats map[string]string) map[string]string {
	poolSize := s.client.Options().PoolSize
	poolStats := s.client.PoolStats()
	connectedClients, _ := strconv.Atoi(redisInfo["connected_clients"])
	highConnectionThreshold := int(float64(poolSize) * 0.8)

	// Check if the number of connected clients is high.
	if connectedClients > highConnectionThreshold {
		stats["redis_message"] = "Redis has a high number of connected clients"
	}

	// Check if the number of stale connections exceeds a threshold.
	minStaleConnectionsThreshold := 500
	if int(poolStats.StaleConns) > minStaleConnectionsThreshold {
		stats["redis_message"] = fmt.Sprintf("Redis has %d stale connections.", poolStats.StaleConns)
	}

	// Check if Redis is using a significant amount of memory.
	usedMemory, _ := strconv.ParseInt(redisInfo["used_memory"], 10, 64)
	maxMemory, _ := strconv.ParseInt(redisInfo["maxmemory"], 10, 64)
	if maxMemory > 0 {
		usedMemoryPercentage := float64(usedMemory) / float64(maxMemory) * 100
		if usedMemoryPercentage >= 90 {
			stats["redis_message"] = "Redis is using a significant amount of memory"
		}
	}

	// Check if Redis has been recently restarted.
	uptimeInSeconds, _ := strconv.ParseInt(redisInfo["uptime_in_seconds"], 10, 64)
	if uptimeInSeconds < 3600 {
		stats["redis_message"] = "Redis has been recently restarted"
	}

	// Check if the number of idle connections is high.
	idleConns := int(poolStats.IdleConns)
	highIdleConnectionThreshold := int(float64(poolSize) * 0.7)
	if idleConns > highIdleConnectionThreshold {
		stats["redis_message"] = "Redis has a high number of idle connections"
	}

	// Check if the connection pool utilization is high.
	poolUtilization := float64(poolStats.TotalConns-poolStats.IdleConns) / float64(poolSize) * 100
	highPoolUtilizationThreshold := 90.0
	if poolUtilization > highPoolUtilizationThreshold {
		stats["redis_message"] = "Redis connection pool utilization is high"
	}

	return stats
}

// parseRedisInfo parses the Redis info response and returns a map of key-value pairs.
func parseRedisInfo(info string) map[string]string {
	result := make(map[string]string)
	lines := strings.Split(info, "\r\n")
	for _, line := range lines {
		if strings.Contains(line, ":") {
			parts := strings.Split(line, ":")
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			result[key] = value
		}
	}
	return result
}
