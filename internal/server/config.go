package server

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// MongoDBCfg holds MongoDB connection configuration
type MongoDBCfg struct {
	URI      string
	Database string
}

// RedisCfg holds Redis connection configuration
type RedisCfg struct {
	Addr     string
	Password string
	DB       int
}

// Config holds server configuration
type Config struct {
	Port            string
	MongoDB         MongoDBCfg
	Redis           RedisCfg
	ShutdownTimeout time.Duration
}

// NewConfig creates a new Config instance with values from environment variables
func NewConfig() *Config {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		slog.Warn("Warning: .env file not found or error loading it", "error", err)
	}

	// Handle port configuration
	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil || port <= 0 {
		port = 8080 // Default port if not specified
	}

	mongoURI := getEnv("MONGODB_URI", "")
	if mongoURI == "" {
		mongoURI = fmt.Sprintf("mongodb://%s:%s@%s:%s",
			getEnv("DB_USER", "admin"),
			getEnv("DB_PASSWORD", "password123"),
			getEnv("DB_HOST", "localhost"),
			getEnv("DB_PORT", "27017"),
		)
	}

	// Parse Redis DB index
	redisDB, err := strconv.Atoi(getEnv("REDIS_DB", "0"))
	if err != nil {
		redisDB = 0
	}

	return &Config{
		Port: fmt.Sprintf(":%d", port),
		MongoDB: MongoDBCfg{
			URI:      mongoURI,
			Database: getEnv("DB_NAME", "development_db"),
		},
		Redis: RedisCfg{
			Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       redisDB,
		},
		ShutdownTimeout: 10 * time.Second,
	}
}

// getEnv gets environment variable or returns default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
