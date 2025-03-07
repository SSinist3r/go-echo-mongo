package server

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
)

// Server represents the HTTP server
type Server struct {
	config *Config
	echo   *echo.Echo
	db     *mongo.Database
	redis  *redis.Client
}

// NewServer creates and initializes a new server instance
func NewServer() *Server {
	e := echo.New()
	config := NewConfig()

	// Setup middleware
	setupMiddleware(e)

	// Setup validator
	setupValidator(e)

	return &Server{
		config: config,
		echo:   e,
	}
}

// Start initializes the server, sets up routes and starts listening
func (s *Server) Start() error {
	// Initialize all dependencies
	s.db, s.redis = bootstrap(s.echo, s.config)

	// Start server
	go s.startServer()

	return s.gracefulShutdown()
}

func (s *Server) startServer() {
	if err := s.echo.Start(s.config.Port); err != nil && err != http.ErrServerClosed {
		slog.Error("Shutting down the server", "error", err)
		log.Fatalf("http server error: %s", err)
	}
}

// gracefulShutdown handles the graceful shutdown process of the server
func (s *Server) gracefulShutdown() error {
	// Create a context that will be canceled when shutdown signal is received
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Wait for interrupt signal
	<-ctx.Done()

	slog.Info("shutting down gracefully, press Ctrl+C again to force")

	// Create a timeout context for shutdown operations
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create a channel to track shutdown completion
	done := make(chan bool, 1)
	go func() {
		// Close MongoDB connection
		if err := s.db.Client().Disconnect(shutdownCtx); err != nil {
			slog.Error("Error disconnecting from MongoDB", "error", err)
		}
		// Close Redis connection
		if err := s.redis.Close(); err != nil {
			slog.Error("Error disconnecting from Redis", "error", err)
		}
		// Shutdown server
		if err := s.echo.Shutdown(shutdownCtx); err != nil {
			slog.Error("Error during server shutdown", "error", err)
		}

		done <- true
	}()

	// Wait for shutdown to complete or timeout
	select {
	case <-done:
		slog.Info("Server shutdown complete")
		return nil
	case <-shutdownCtx.Done():
		return fmt.Errorf("shutdown timed out")
	}
}
