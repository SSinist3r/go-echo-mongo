// Package database provides database connection utilities.
package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// Config holds MongoDB connection configuration
type Config struct {
	URI             string
	Database        string
	ConnectTimeout  time.Duration
	MaxConnIdleTime time.Duration
	MaxPoolSize     uint64
	MinPoolSize     uint64
	RetryWrites     bool
	RetryReads      bool
	MaxRetries      uint64
}

// DefaultConfig returns a default MongoDB configuration
func DefaultConfig() Config {
	return Config{
		ConnectTimeout:  10 * time.Second,
		MaxConnIdleTime: 30 * time.Second,
		MaxPoolSize:     100,
		MinPoolSize:     10,
		RetryWrites:     true,
		RetryReads:      true,
		MaxRetries:      3,
	}
}

// MongoDBService defines the interface for MongoDB operations
type MongoDBService interface {
	GetClient() *mongo.Client
	GetDatabase() *mongo.Database
	IsConnected(ctx context.Context) bool
	Disconnect(ctx context.Context) error
	// Add other methods as needed
}

// mongoDBService implements the MongoDBService interface
type mongoDBService struct {
	client   *mongo.Client
	database *mongo.Database
}

// NewMongoDBService creates a new MongoDB service instance
func NewMongoDBService(config Config) (MongoDBService, error) {
	db, client, err := connect(config)
	if err != nil {
		return nil, err
	}

	return &mongoDBService{
		client:   client,
		database: db,
	}, nil
}

// GetClient returns the MongoDB client instance
func (s *mongoDBService) GetClient() *mongo.Client {
	return s.client
}

// GetDatabase returns the MongoDB database instance
func (s *mongoDBService) GetDatabase() *mongo.Database {
	return s.database
}

// Connect establishes connection to MongoDB and returns the database instance
func connect(config Config) (*mongo.Database, *mongo.Client, error) {
	slog.Info("Attempting to connect to MongoDB")

	// Set up context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), config.ConnectTimeout)
	defer cancel()

	// Create MongoDB client options
	clientOptions := options.Client().
		ApplyURI(config.URI).
		SetMaxConnIdleTime(config.MaxConnIdleTime).
		SetMaxPoolSize(config.MaxPoolSize).
		SetMinPoolSize(config.MinPoolSize).
		SetRetryWrites(config.RetryWrites).
		SetRetryReads(config.RetryReads).
		SetMaxConnecting(config.MaxRetries)

	// Connect to MongoDB with retry logic
	var client *mongo.Client
	var err error
	for i := uint64(0); i <= config.MaxRetries; i++ {
		client, err = mongo.Connect(ctx, clientOptions)
		if err == nil {
			break
		}
		if i < config.MaxRetries {
			slog.Warn("Failed to connect to MongoDB", "attempt", i+1, "maxRetries", config.MaxRetries, "error", err)
			time.Sleep(time.Second * time.Duration(i+1)) // Exponential backoff
		}
	}
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create MongoDB client after %d attempts: %w", config.MaxRetries, err)
	}

	// Ping the database with retry logic
	for i := uint64(0); i <= config.MaxRetries; i++ {
		err = client.Ping(ctx, readpref.Primary())
		if err == nil {
			break
		}
		if i < config.MaxRetries {
			slog.Warn("Failed to ping MongoDB", "attempt", i+1, "maxRetries", config.MaxRetries, "error", err)
			time.Sleep(time.Second * time.Duration(i+1)) // Exponential backoff
		}
	}
	if err != nil {
		return nil, nil, fmt.Errorf("failed to ping MongoDB after %d attempts: %w", config.MaxRetries, err)
	}

	slog.Info("Successfully connected to MongoDB")
	return client.Database(config.Database), client, nil
}

// Disconnect closes the MongoDB connection
func (s *mongoDBService) Disconnect(ctx context.Context) error {
	if s.client == nil {
		return nil
	}

	if err := s.client.Disconnect(ctx); err != nil {
		return fmt.Errorf("failed to disconnect from MongoDB: %w", err)
	}

	slog.Info("Successfully disconnected from MongoDB")
	return nil
}

// IsConnected checks if the MongoDB client is connected and healthy
func (s *mongoDBService) IsConnected(ctx context.Context) bool {
	if s.client == nil {
		return false
	}
	return s.client.Ping(ctx, readpref.Primary()) == nil
}
