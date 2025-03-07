package server

import (
	"log"
	"log/slog"
	"os"

	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	slogecho "github.com/samber/slog-echo"
	slogzerolog "github.com/samber/slog-zerolog/v2"
	"go.mongodb.org/mongo-driver/mongo"

	"go-echo-mongo/internal/handler"
	"go-echo-mongo/internal/repository"
	"go-echo-mongo/internal/repository/redisrepo"
	"go-echo-mongo/internal/service"
	"go-echo-mongo/pkg/database"
	"go-echo-mongo/pkg/ratelimit"
	"go-echo-mongo/pkg/web/mwutil"
	"go-echo-mongo/pkg/web/response"
)

// Bootstrap initializes all dependencies and sets up the server
func bootstrap(e *echo.Echo, cfg *Config) (*mongo.Database, *redis.Client) {
	// Setup logger
	logger := setupLogger()

	// Setup echo logger
	setupEchoLogger(e, logger)

	// Setup prometheus
	setupPrometheus(e)

	// Setup database
	db := setupDatabase(cfg)

	// Setup Redis
	redisClient := setupRedis(e, cfg)

	// Setup Repositories, Services and Routes
	setupReposServicesRoutes(e, db, redisClient)

	slog.Info("Server initialized successfully")

	return db, redisClient
}

// setupLogger initializes and configures the logger
func setupLogger() *slog.Logger {
	zerologLogger := zerolog.New(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: "2006-01-02 15:04:05",
	}).With().Timestamp().Caller().Logger()

	logger := slog.New(slogzerolog.Option{
		Level:  slog.LevelDebug,
		Logger: &zerologLogger,
	}.NewZerologHandler())

	slog.SetDefault(logger)
	return logger
}

// setupEchoLogger sets up slog middleware for echo instance
func setupEchoLogger(e *echo.Echo, logger *slog.Logger) {
	e.Use(slogecho.New(logger))
}

// setupPrometheus sets up Prometheus middleware for echo instance and metrics endpoints
func setupPrometheus(e *echo.Echo) {
	// Add Prometheus middleware with custom configuration
	e.Use(echoprometheus.NewMiddlewareWithConfig(echoprometheus.MiddlewareConfig{
		// Skip metrics collection for the metrics endpoint itself
		Skipper: func(c echo.Context) bool {
			return c.Path() == "/metrics"
		},
	}))

	// Register Prometheus metrics endpoint
	e.GET("/metrics", echoprometheus.NewHandler())
}

// setupDatabase initializes the MongoDB connection
func setupDatabase(cfg *Config) *mongo.Database {
	dbConfig := database.DefaultConfig()
	dbConfig.URI = cfg.MongoDB.URI
	dbConfig.Database = cfg.MongoDB.Database

	mongoDBService, err := database.NewMongoDBService(dbConfig)
	if err != nil {
		slog.Error("Failed to connect to MongoDB", "error", err)
		log.Fatal(err)
	}

	return mongoDBService.GetDatabase()
}

// setupRedis initializes the Redis connection
func setupRedis(e *echo.Echo, cfg *Config) *redis.Client {
	redisConfig := database.DefaultRedisConfig()

	// Override defaults with config values if provided
	if cfg.Redis.Addr != "" {
		redisConfig.Addr = cfg.Redis.Addr
	}
	if cfg.Redis.Password != "" {
		redisConfig.Password = cfg.Redis.Password
	}
	if cfg.Redis.DB != 0 {
		redisConfig.DB = cfg.Redis.DB
	}

	redisService, err := database.NewRedisService(redisConfig)
	if err != nil {
		slog.Error("Failed to connect to Redis", "error", err)
		log.Fatal(err)
	}

	// Setup Redis health check endpoint
	e.GET("/redis/health", func(c echo.Context) error {
		if redisService.IsConnected(c.Request().Context()) {
			return response.OK(c, "Redis is healthy", redisService.Health())
		}
		return response.InternalError(c, "Redis is not healthy")
	})

	return redisService.GetClient()
}

// setupRedisRepositories initializes all Redis repositories
func setupRedisRepositories(redisClient *redis.Client) (redisrepo.Repository, redisrepo.CacheRepository, redisrepo.SessionRepository, redisrepo.RateLimitRepository) {
	// Create base Redis repository
	baseRepo := redisrepo.New(redisClient)

	// Create specialized repositories
	cacheRepo := redisrepo.NewCacheRepository(baseRepo)
	sessionRepo := redisrepo.NewSessionRepository(baseRepo)
	rateLimitRepo := redisrepo.NewRateLimitRepository(baseRepo)

	return baseRepo, cacheRepo, sessionRepo, rateLimitRepo
}

func setupReposServicesRoutes(e *echo.Echo, db *mongo.Database, redisClient *redis.Client) {
	// Initialize Redis repositories
	baseRedisRepo, cacheRepo, sessionRepo, rateLimitRepo := setupRedisRepositories(redisClient)

	// Log Redis repositories initialization
	slog.Info("Redis repositories initialized",
		"baseRepo", baseRedisRepo != nil,
		"cacheRepo", cacheRepo != nil,
		"sessionRepo", sessionRepo != nil,
		"rateLimitRepo", rateLimitRepo != nil)

	// Set the rate limit repo for rate limit middleware
	ratelimit.SetRateLimitRepo(rateLimitRepo)

	// Initialize MongoDB repositories
	userRepo := repository.NewUserRepository(db)
	productRepo := repository.NewProductRepository(db)

	// Initialize services
	userService := service.NewUserService(userRepo, baseRedisRepo)
	productService := service.NewProductService(productRepo, baseRedisRepo)
	// Add new services here as needed

	// Set API key validator
	mwutil.SetAPIKeyValidator(userService)

	// Initialize handlers and register routes
	routesRegistry := NewRegistry()
	routesRegistry.Add(handler.NewUserHandler(userService))
	routesRegistry.Add(handler.NewProductHandler(productService))
	// Add new handlers here as needed
	routesRegistry.RegisterAll(e)
}
