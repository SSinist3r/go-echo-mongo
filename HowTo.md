# How To Guide

This guide provides detailed instructions on how to extend and customize the Go Echo Mongo application, including adding new routes, implementing rate limiting, and using API key authentication.

## Table of Contents

- [Adding New Routes](#adding-new-routes)
  - [Step 1: Create Model](#step-1-create-model)
  - [Step 2: Create Repository](#step-2-create-repository)
  - [Step 3: Create Service](#step-3-create-service)
  - [Step 4: Create DTOs](#step-4-create-dtos)
  - [Step 5: Create Handler](#step-5-create-handler)
  - [Step 6: Register in Bootstrap](#step-6-register-in-bootstrap)
- [Rate Limiting Strategies](#rate-limiting-strategies)
  - [Fixed Window Rate Limiting](#fixed-window-rate-limiting)
  - [Sliding Window Rate Limiting](#sliding-window-rate-limiting)
  - [Token Bucket Rate Limiting](#token-bucket-rate-limiting)
  - [Leaky Bucket Rate Limiting](#leaky-bucket-rate-limiting)
  - [Custom Rate Limiting](#custom-rate-limiting)
- [API Key Authentication](#api-key-authentication)
  - [Basic Authentication](#basic-authentication)
  - [Role-Based Authentication](#role-based-authentication)
  - [Custom Authentication Rules](#custom-authentication-rules)

## Adding New Routes

Follow these steps to add a new entity (e.g., "Admin") to the application:

### Step 1: Create Model

Create a new model in `internal/model/admin.go`:

```go
package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Admin represents an admin user in the system
type Admin struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name      string             `bson:"name" json:"name"`
	Email     string             `bson:"email" json:"email"`
	Password  string             `bson:"password" json:"-"` // Password is not exported in JSON
	ApiKey    string             `bson:"api_key" json:"-"`  // API key is not exported in JSON
	Roles     []string           `bson:"roles" json:"roles"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}

// GetID returns the ID of the admin
func (a *Admin) GetID() primitive.ObjectID {
	return a.ID
}

// SetID sets the ID of the admin
func (a *Admin) SetID(id primitive.ObjectID) {
	a.ID = id
}

// GetCreatedAt returns the creation time of the admin
func (a *Admin) GetCreatedAt() time.Time {
	return a.CreatedAt
}

// SetCreatedAt sets the creation time of the admin
func (a *Admin) SetCreatedAt(t time.Time) {
	a.CreatedAt = t
}

// GetUpdatedAt returns the last update time of the admin
func (a *Admin) GetUpdatedAt() time.Time {
	return a.UpdatedAt
}

// SetUpdatedAt sets the last update time of the admin
func (a *Admin) SetUpdatedAt(t time.Time) {
	a.UpdatedAt = t
}

// HasRole checks if the admin has a specific role
func (a *Admin) HasRole(role string) bool {
	for _, r := range a.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// HasAnyRole checks if the admin has any of the specified roles
func (a *Admin) HasAnyRole(roles ...string) bool {
	for _, role := range roles {
		if a.HasRole(role) {
			return true
		}
	}
	return false
}
```

### Step 2: Create Repository

Create a new repository in `internal/repository/admin_repository.go`:

```go
package repository

import (
	"context"
	"go-echo-mongo/internal/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// AdminRepository defines the interface for admin-related database operations
type AdminRepository interface {
	BaseRepository[*model.Admin]
	FindByEmail(context.Context, string) (*model.Admin, error)
	FindByApiKey(context.Context, string) (*model.Admin, error)
}

// adminRepository implements AdminRepository interface
type adminRepository struct {
	BaseRepository[*model.Admin]
}

// NewAdminRepository creates a new AdminRepository instance
func NewAdminRepository(db *mongo.Database) AdminRepository {
	return &adminRepository{
		BaseRepository: newBaseRepository[*model.Admin](db.Collection("admins")),
	}
}

// FindByEmail finds an admin by email
func (r *adminRepository) FindByEmail(ctx context.Context, email string) (*model.Admin, error) {
	var admin *model.Admin
	err := r.GetCollection().FindOne(ctx, bson.M{"email": email}).Decode(&admin)
	return admin, err
}

// FindByApiKey finds an admin by API key
func (r *adminRepository) FindByApiKey(ctx context.Context, apiKey string) (*model.Admin, error) {
	var admin *model.Admin
	err := r.GetCollection().FindOne(ctx, bson.M{"api_key": apiKey}).Decode(&admin)
	return admin, err
}
```

### Step 3: Create Service

Create a new service in `internal/service/admin_service.go`:

```go
package service

import (
	"context"
	"log"

	"go-echo-mongo/internal/model"
	"go-echo-mongo/internal/repository"
	"go-echo-mongo/internal/repository/redisrepo"
	"go-echo-mongo/pkg/secutil"
	"go-echo-mongo/pkg/strutil"
)

// AdminService defines the interface for admin-related business logic
type AdminService interface {
	BaseService[*model.Admin]
	GetByEmail(ctx context.Context, email string) (*model.Admin, error)
	GetByApiKey(ctx context.Context, apiKey string) (*model.Admin, error)
	ValidateCredentials(ctx context.Context, email, password string) (*model.Admin, error)
}

type adminService struct {
	BaseService[*model.Admin]
	repo  repository.AdminRepository
	redis redisrepo.Repository
}

// NewAdminService creates a new AdminService instance
func NewAdminService(repo repository.AdminRepository, redis redisrepo.Repository) AdminService {
	if repo == nil {
		log.Fatal(ErrNilRepository)
	}
	return &adminService{
		BaseService: newBaseService(repo),
		repo:        repo,
		redis:       redis,
	}
}

// Create overrides base Create to add email check and password hashing
func (s *adminService) Create(ctx context.Context, admin *model.Admin) error {
	if err := validateContext(ctx); err != nil {
		return err
	}

	// Check if email already exists
	existingAdmin, err := s.GetByEmail(ctx, admin.Email)
	if err == nil && existingAdmin != nil {
		return ErrEmailExists
	}

	// Hash password
	hashedPassword, err := secutil.HashPassword(admin.Password)
	if err != nil {
		return err
	}
	admin.Password = hashedPassword

	// Generate API key
	apiKey, err := strutil.GenerateRandom(32, false, true, true, false)
	if err != nil {
		return err
	}
	admin.ApiKey = apiKey

	// Ensure admin has at least the basic admin role
	if len(admin.Roles) == 0 {
		admin.Roles = []string{model.RoleAdmin}
	}

	return s.BaseService.Create(ctx, admin)
}

// GetByEmail retrieves an admin by email
func (s *adminService) GetByEmail(ctx context.Context, email string) (*model.Admin, error) {
	if err := validateContext(ctx); err != nil {
		return nil, err
	}
	return s.repo.FindByEmail(ctx, email)
}

// GetByApiKey retrieves an admin by API key
func (s *adminService) GetByApiKey(ctx context.Context, apiKey string) (*model.Admin, error) {
	if err := validateContext(ctx); err != nil {
		return nil, err
	}
	return s.repo.FindByApiKey(ctx, apiKey)
}

// ValidateCredentials validates admin credentials and returns the admin if valid
func (s *adminService) ValidateCredentials(ctx context.Context, email, password string) (*model.Admin, error) {
	if err := validateContext(ctx); err != nil {
		return nil, err
	}

	admin, err := s.GetByEmail(ctx, email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if err := secutil.VerifyPassword(admin.Password, password); err != nil {
		return nil, ErrInvalidCredentials
	}

	return admin, nil
}
```

### Step 4: Create DTOs

Create DTOs in `internal/dto/admin_dto.go`:

```go
package dto

import (
	"go-echo-mongo/internal/model"
	"time"

	"github.com/go-playground/validator/v10"
)

// CreateAdminRequest represents the request body for creating a new admin
type CreateAdminRequest struct {
	Name     string   `json:"name" validate:"required,min=2,max=100"`
	Email    string   `json:"email" validate:"required,email"`
	Password string   `json:"password" validate:"required,min=8"`
	Roles    []string `json:"roles"`
}

// ToModel converts the request DTO to a model
func (r *CreateAdminRequest) ToModel() *model.Admin {
	return &model.Admin{
		Name:     r.Name,
		Email:    r.Email,
		Password: r.Password,
		Roles:    r.Roles,
	}
}

// UpdateAdminRequest represents the request body for updating an admin
type UpdateAdminRequest struct {
	Name  string   `json:"name" validate:"omitempty,min=2,max=100"`
	Email string   `json:"email" validate:"omitempty,email"`
	Roles []string `json:"roles"`
}

// ToModel converts the request DTO to a model, preserving existing data
func (r *UpdateAdminRequest) ToModel(existingAdmin *model.Admin) *model.Admin {
	// Only update fields that were provided
	if r.Name != "" {
		existingAdmin.Name = r.Name
	}
	if r.Email != "" {
		existingAdmin.Email = r.Email
	}
	if r.Roles != nil {
		existingAdmin.Roles = r.Roles
	}
	return existingAdmin
}

// AdminResponse represents the response body for admin operations
type AdminResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Roles     []string  `json:"roles"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NewAdminResponse creates a new AdminResponse from a model
func NewAdminResponse(admin *model.Admin) AdminResponse {
	return AdminResponse{
		ID:        admin.ID.Hex(),
		Name:      admin.Name,
		Email:     admin.Email,
		Roles:     admin.Roles,
		CreatedAt: admin.CreatedAt,
		UpdatedAt: admin.UpdatedAt,
	}
}

// NewAdminResponseList creates a list of AdminResponse from models
func NewAdminResponseList(admins []*model.Admin) []AdminResponse {
	var responses []AdminResponse
	for _, admin := range admins {
		responses = append(responses, NewAdminResponse(admin))
	}
	return responses
}

// AdminLoginRequest represents the request body for admin login
type AdminLoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// Validate validates the AdminLoginRequest
func (r *AdminLoginRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}
```

### Step 5: Create Handler

Create a new handler in `internal/handler/admin_handler.go`:

```go
package handler

import (
	"errors"
	"go-echo-mongo/internal/dto"
	"go-echo-mongo/internal/model"
	"go-echo-mongo/internal/service"
	"go-echo-mongo/pkg/web/mwutil"
	"go-echo-mongo/pkg/web/response"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
)

// AdminHandler defines the interface for admin-related HTTP handlers
type AdminHandler interface {
	Register(e *echo.Echo)
	Create(c echo.Context) error
	GetByID(c echo.Context) error
	GetAll(c echo.Context) error
	GetPaginated(c echo.Context) error
	Update(c echo.Context) error
	Delete(c echo.Context) error
	Login(c echo.Context) error
}

// adminHandler implements AdminHandler interface
type adminHandler struct {
	service service.AdminService
}

// NewAdminHandler creates a new AdminHandler instance
func NewAdminHandler(service service.AdminService) AdminHandler {
	return &adminHandler{
		service: service,
	}
}

// Register registers all admin routes
func (h *adminHandler) Register(e *echo.Echo) {
	admins := e.Group("/api/v1/admins")
	
	// Apply rate limiting to admin routes - 2 requests per minute
	admins.Use(mwutil.NewFixedRateLimiter(2, 1*time.Minute))
	
	// Only admins can access admin routes
	admins.Use(mwutil.NewAPIKeyAuth(model.RoleAdmin))
	
	admins.POST("", h.Create)
	admins.GET("", h.GetAll)
	admins.GET("/paginated", h.GetPaginated)
	admins.GET("/:id", h.GetByID)
	admins.PUT("/:id", h.Update)
	admins.DELETE("/:id", h.Delete)
	
	// Login endpoint - no authentication needed
	e.POST("/api/v1/admin/login", h.Login)
}

// Create handles admin creation
func (h *adminHandler) Create(c echo.Context) error {
	req := new(dto.CreateAdminRequest)
	if err := c.Bind(req); err != nil {
		return response.BadRequest(c, "Invalid request format")
	}

	if err := c.Validate(req); err != nil {
		return response.ValidationError(c, err)
	}

	admin := req.ToModel()
	if err := h.service.Create(c.Request().Context(), admin); err != nil {
		switch {
		case errors.Is(err, service.ErrEmailExists):
			return response.Conflict(c, "Admin with this email already exists")
		default:
			return response.InternalError(c, "Failed to create admin")
		}
	}

	return response.Created(c, "Admin created successfully", dto.NewAdminResponse(admin))
}

// GetByID handles retrieving an admin by ID
func (h *adminHandler) GetByID(c echo.Context) error {
	admin, err := h.service.GetByID(c.Request().Context(), c.Param("id"))
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserNotFound):
			return response.NotFound(c, "Admin not found")
		default:
			return response.InternalError(c, "Failed to retrieve admin")
		}
	}

	return response.OK(c, "Admin retrieved successfully", dto.NewAdminResponse(admin))
}

// GetAll handles retrieving all admins
func (h *adminHandler) GetAll(c echo.Context) error {
	admins, err := h.service.GetAll(c.Request().Context())
	if err != nil {
		return response.InternalError(c, "Failed to retrieve admins")
	}

	return response.OK(c, "Admins retrieved successfully", dto.NewAdminResponseList(admins))
}

// GetPaginated handles retrieving admins with pagination
func (h *adminHandler) GetPaginated(c echo.Context) error {
	// Parse pagination parameters from query
	page, err := strconv.ParseInt(c.QueryParam("page"), 10, 64)
	if err != nil || page < 1 {
		page = 1
	}

	itemsPerPage, err := strconv.ParseInt(c.QueryParam("items_per_page"), 10, 64)
	if err != nil || itemsPerPage < 1 {
		itemsPerPage = 10
	}

	// Get admins with pagination
	admins, totalCount, err := h.service.GetPaginated(
		c.Request().Context(),
		nil,
		page,
		itemsPerPage,
	)
	if err != nil {
		return response.InternalError(c, "Failed to retrieve admins")
	}

	// Return paginated response
	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": dto.NewAdminResponseList(admins),
		"meta": map[string]interface{}{
			"current_page":   page,
			"items_per_page": itemsPerPage,
			"total_items":    totalCount,
			"total_pages":    (totalCount + itemsPerPage - 1) / itemsPerPage,
		},
	})
}

// Update handles updating an admin
func (h *adminHandler) Update(c echo.Context) error {
	req := new(dto.UpdateAdminRequest)
	if err := c.Bind(req); err != nil {
		return response.BadRequest(c, "Invalid request format")
	}

	if err := c.Validate(req); err != nil {
		return response.ValidationError(c, err)
	}

	// Get existing admin first
	existingAdmin, err := h.service.GetByID(c.Request().Context(), c.Param("id"))
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserNotFound):
			return response.NotFound(c, "Admin not found")
		default:
			return response.InternalError(c, "Failed to retrieve admin")
		}
	}

	// Update only the fields that were provided
	updatedAdmin := req.ToModel(existingAdmin)
	if err := h.service.Update(c.Request().Context(), c.Param("id"), updatedAdmin); err != nil {
		switch {
		case errors.Is(err, service.ErrUserNotFound):
			return response.NotFound(c, "Admin not found")
		case errors.Is(err, service.ErrEmailExists):
			return response.Conflict(c, "Email is already taken")
		default:
			return response.InternalError(c, "Failed to update admin")
		}
	}

	return response.OK(c, "Admin updated successfully", dto.NewAdminResponse(updatedAdmin))
}

// Delete handles deleting an admin
func (h *adminHandler) Delete(c echo.Context) error {
	if err := h.service.Delete(c.Request().Context(), c.Param("id")); err != nil {
		switch {
		case errors.Is(err, service.ErrUserNotFound):
			return response.NotFound(c, "Admin not found")
		default:
			return response.InternalError(c, "Failed to delete admin")
		}
	}

	return response.NoContent(c)
}

// Login handles admin authentication
func (h *adminHandler) Login(c echo.Context) error {
	req := new(dto.AdminLoginRequest)
	if err := c.Bind(req); err != nil {
		return response.BadRequest(c, "Invalid request format")
	}

	if err := c.Validate(req); err != nil {
		return response.ValidationError(c, err)
	}

	admin, err := h.service.ValidateCredentials(c.Request().Context(), req.Email, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidCredentials):
			return response.Unauthorized(c, "Invalid email or password")
		default:
			return response.InternalError(c, "Failed to authenticate admin")
		}
	}

	return response.OK(c, "Login successful", dto.NewAdminResponse(admin))
}
```

### Step 6: Register in Bootstrap

Update `internal/server/bootstrap.go` to register your new repository, service, and handler:

```go
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
    adminRepo := repository.NewAdminRepository(db) // New admin repository

    // Initialize services
    userService := service.NewUserService(userRepo, baseRedisRepo)
    productService := service.NewProductService(productRepo, baseRedisRepo)
    adminService := service.NewAdminService(adminRepo, baseRedisRepo) // New admin service
    
    // Set API key validator
    mwutil.SetAPIKeyValidator(userService)

    // Initialize handlers and register routes
    routesRegistry := NewRegistry()
    routesRegistry.Add(handler.NewUserHandler(userService))
    routesRegistry.Add(handler.NewProductHandler(productService))
    routesRegistry.Add(handler.NewAdminHandler(adminService)) // New admin handler
    
    routesRegistry.RegisterAll(e)
}
```

## Rate Limiting Strategies

The application provides multiple rate limiting strategies that you can use to protect your API endpoints.

### Fixed Window Rate Limiting

Fixed window rate limiting restricts the number of requests within a fixed time window (e.g., 10 requests per minute).

```go
// Apply rate limiting to a group of routes
usersGroup := e.Group("/api/v1/users")
usersGroup.Use(mwutil.NewFixedRateLimiter(10, 1*time.Minute))

// Apply to a specific route
e.POST("/api/v1/login", loginHandler, mwutil.NewFixedRateLimiter(5, 1*time.Minute))

// Apply per-path rate limiting (different limits for different paths)
e.GET("/api/v1/products", getProductsHandler, mwutil.NewFixedRateLimiterPerPath(100, 1*time.Minute))
```

### Sliding Window Rate Limiting

Sliding window rate limiting provides smoother limits by considering requests within a sliding time window.

```go
// 20 requests per minute with a sliding window
apiGroup := e.Group("/api/v1")
apiGroup.Use(mwutil.NewSlidingRateLimiter(20, 1*time.Minute))

// Per-path sliding window rate limiting
e.GET("/api/v1/users/:id", getUserHandler, mwutil.NewSlidingRateLimiterPerPath(5, 30*time.Second))
```

### Token Bucket Rate Limiting

Token bucket rate limiting allows bursts of traffic while maintaining a steady average rate.

```go
// 5 tokens per second with maximum burst of 20
loginGroup := e.Group("/api/v1/auth")
loginGroup.Use(mwutil.NewTokenBucketLimiter(5, 20, 10*time.Minute))

// Per-path token bucket rate limiting
e.POST("/api/v1/orders", createOrderHandler, mwutil.NewTokenBucketLimiterPerPath(2, 10, 5*time.Minute))
```

### Leaky Bucket Rate Limiting

Leaky bucket rate limiting controls the flow of requests at a constant rate.

```go
// 10 capacity bucket with 2 requests/second leak rate
contactGroup := e.Group("/api/v1/contact")
contactGroup.Use(mwutil.NewLeakyBucketLimiter(10, 2, 5*time.Minute))

// Per-path leaky bucket rate limiting
e.POST("/api/v1/uploads", uploadHandler, mwutil.NewLeakyBucketLimiterPerPath(5, 1, 5*time.Minute))
```

### Custom Rate Limiting

You can configure custom rate limiting using the general configuration:

```go
// Configure custom rate limiting
config := mwutil.RateLimitConfig{
    Strategy: mwutil.TokenBucket,
    Rate:     3.0,  // 3 tokens per second
    Burst:    15,   // Maximum bucket size
    Window:   5 * time.Minute,
}

// Apply the rate limiter
adminGroup := e.Group("/api/v1/admin")
adminGroup.Use(mwutil.NewRateLimiter(config))
```

## API Key Authentication

The application uses API key authentication to protect API endpoints.

### Basic Authentication

To protect a route with API key authentication (any valid API key):

```go
// Apply API key authentication to a group of routes
privateGroup := e.Group("/api/v1/private")
privateGroup.Use(mwutil.NewAPIKeyAuth())

// Apply to a specific route
e.POST("/api/v1/resource", createResourceHandler, mwutil.NewAPIKeyAuth())
```

### Role-Based Authentication

To restrict access to users with specific roles:

```go
// Only admins can access
adminGroup := e.Group("/api/v1/admin")
adminGroup.Use(mwutil.NewAPIKeyAuth(model.RoleAdmin))

// Either admin or manager role is required
managementGroup := e.Group("/api/v1/management")
managementGroup.Use(mwutil.NewAPIKeyAuth(model.RoleAdmin, model.RoleManager))

// Protecting a specific route with role-based auth
e.DELETE("/api/v1/users/:id", deleteUserHandler, mwutil.NewAPIKeyAuth(model.RoleAdmin))
```

### Custom Authentication Rules

For more complex authentication scenarios, use the config-based API key authentication:

```go
// Custom auth config
config := mwutil.APIKeyAuthConfig{
    Skipper: func(c echo.Context) bool {
        // Skip authentication for specific paths or methods
        return c.Path() == "/api/v1/public" || c.Request().Method == "GET"
    },
    KeyLookup: "header:Authorization", // Use Authorization header instead of X-API-Key
    ContextKey: "current_user",        // Custom context key for the user
    RequiredRoles: []string{model.RoleAdmin, model.RoleManager},
    ErrorHandler: func(c echo.Context, err error) error {
        // Custom error handling
        return c.JSON(http.StatusForbidden, map[string]string{
            "error": "You don't have permission to access this resource",
        })
    },
}

// Apply the custom auth middleware
e.POST("/api/v1/sensitive", sensitiveHandler, mwutil.NewAPIKeyAuthWithConfig(config))
```

Remember to set the API key validator in the bootstrap process:

```go
// In bootstrap.go
mwutil.SetAPIKeyValidator(userService)
```

This setup allows the middleware to validate API keys using your user service. You can change this to use another service (like the admin service) if needed. 