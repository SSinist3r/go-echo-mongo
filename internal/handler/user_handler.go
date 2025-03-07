package handler

import (
	"errors"
	"fmt"
	"go-echo-mongo/internal/dto"
	"go-echo-mongo/internal/model"
	"go-echo-mongo/internal/service"
	"go-echo-mongo/pkg/web/mwutil"
	"go-echo-mongo/pkg/web/response"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

// UserHandler defines the interface for user-related HTTP handlers
type UserHandler interface {
	Register(e *echo.Echo)
	Create(c echo.Context) error
	GetByID(c echo.Context) error
	GetAll(c echo.Context) error
	GetPaginated(c echo.Context) error
	Update(c echo.Context) error
	Delete(c echo.Context) error
	Login(c echo.Context) error

	// Batch operations
	CreateMany(c echo.Context) error
	FindByFilter(c echo.Context) error
	UpdateMany(c echo.Context) error
	DeleteMany(c echo.Context) error
}

// userHandler implements UserHandler interface
type userHandler struct {
	service service.UserService
}

// NewUserHandler creates a new UserHandler instance
func NewUserHandler(service service.UserService) UserHandler {
	return &userHandler{
		service: service,
	}
}

// Register registers all user routes
func (h *userHandler) Register(e *echo.Echo) {
	users := e.Group("/api/v1/users")
	users.Use(mwutil.NewFixedRateLimiter(3, 1*time.Minute))
	users.POST("", h.Create, mwutil.NewAPIKeyAuth(model.RoleAdmin))
	users.GET("", h.GetAll)
	users.GET("/paginated", h.GetPaginated)
	users.GET("/:id", h.GetByID)
	users.PUT("/:id", h.Update)
	users.DELETE("/:id", h.Delete)
	users.POST("/login", h.Login)

	// Batch operation routes
	users.POST("/batch", h.CreateMany, mwutil.NewAPIKeyAuth(model.RoleAdmin))
	users.POST("/filter", h.FindByFilter)
	users.PUT("/batch", h.UpdateMany)
	users.DELETE("/batch", h.DeleteMany)
}

// Create handles user creation
func (h *userHandler) Create(c echo.Context) error {
	req := new(dto.CreateUserRequest)
	if err := c.Bind(req); err != nil {
		return response.BadRequest(c, "Invalid request format")
	}

	if err := c.Validate(req); err != nil {
		return response.ValidationError(c, err)
	}

	user := req.ToModel()
	if err := h.service.Create(c.Request().Context(), user); err != nil {
		switch {
		case errors.Is(err, service.ErrEmailExists):
			return response.Conflict(c, "User with this email already exists")
		default:
			return response.InternalError(c, "Failed to create user")
		}
	}

	return response.Created(c, "User created successfully", dto.NewUserResponse(user))
}

// GetByID handles retrieving a user by ID
func (h *userHandler) GetByID(c echo.Context) error {
	user, err := h.service.GetByID(c.Request().Context(), c.Param("id"))
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserNotFound):
			return response.NotFound(c, "User not found")
		default:
			return response.InternalError(c, "Failed to retrieve user")
		}
	}

	return response.OK(c, "User retrieved successfully", dto.NewUserResponse(user))
}

// GetAll handles retrieving all users
func (h *userHandler) GetAll(c echo.Context) error {
	users, err := h.service.GetAll(c.Request().Context())
	if err != nil {
		return response.InternalError(c, "Failed to retrieve users")
	}

	return response.OK(c, "Users retrieved successfully", dto.NewUserResponseList(users))
}

// GetPaginated handles the request to get users with pagination
func (h *userHandler) GetPaginated(c echo.Context) error {
	// Parse pagination parameters from query
	page, err := strconv.ParseInt(c.QueryParam("page"), 10, 64)
	if err != nil || page < 1 {
		page = 1
	}

	itemsPerPage, err := strconv.ParseInt(c.QueryParam("items_per_page"), 10, 64)
	if err != nil || itemsPerPage < 1 {
		itemsPerPage = 10
	}

	// Get users with pagination directly using the base service method
	users, totalCount, err := h.service.GetPaginated(
		c.Request().Context(),
		nil,
		page,
		itemsPerPage,
	)
	if err != nil {
		return response.InternalError(c, "Failed to retrieve users")
	}

	// Convert to response DTOs
	var responses []dto.UserResponse
	for _, user := range users {
		responses = append(responses, dto.UserResponse{
			ID:        user.ID.Hex(),
			Name:      user.Name,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		})
	}

	// Return paginated response
	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": responses,
		"meta": map[string]interface{}{
			"current_page":   page,
			"items_per_page": itemsPerPage,
			"total_items":    totalCount,
			"total_pages":    (totalCount + itemsPerPage - 1) / itemsPerPage,
		},
	})
}

// Update handles updating a user
func (h *userHandler) Update(c echo.Context) error {
	req := new(dto.UpdateUserRequest)
	if err := c.Bind(req); err != nil {
		return response.BadRequest(c, "Invalid request format")
	}

	if err := c.Validate(req); err != nil {
		return response.ValidationError(c, err)
	}

	// Get existing user first
	existingUser, err := h.service.GetByID(c.Request().Context(), c.Param("id"))
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserNotFound):
			return response.NotFound(c, "User not found")
		default:
			return response.InternalError(c, "Failed to retrieve user")
		}
	}

	// Update only the fields that were provided
	updatedUser := req.ToModel(existingUser)
	if err := h.service.Update(c.Request().Context(), c.Param("id"), updatedUser); err != nil {
		switch {
		case errors.Is(err, service.ErrUserNotFound):
			return response.NotFound(c, "User not found")
		case errors.Is(err, service.ErrEmailExists):
			return response.Conflict(c, "Email is already taken")
		default:
			return response.InternalError(c, "Failed to update user")
		}
	}

	return response.OK(c, "User updated successfully", dto.NewUserResponse(updatedUser))
}

// Delete handles deleting a user
func (h *userHandler) Delete(c echo.Context) error {
	if err := h.service.Delete(c.Request().Context(), c.Param("id")); err != nil {
		switch {
		case errors.Is(err, service.ErrUserNotFound):
			return response.NotFound(c, "User not found")
		default:
			return response.InternalError(c, "Failed to delete user")
		}
	}

	return response.NoContent(c)
}

// Login handles user authentication
func (h *userHandler) Login(c echo.Context) error {
	req := new(dto.LoginRequest)
	if err := c.Bind(req); err != nil {
		return response.BadRequest(c, "Invalid request format")
	}

	if err := c.Validate(req); err != nil {
		return response.ValidationError(c, err)
	}

	user, err := h.service.ValidateCredentials(c.Request().Context(), req.Email, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidCredentials):
			return response.Unauthorized(c, "Invalid email or password")
		default:
			return response.InternalError(c, "Failed to authenticate user")
		}
	}

	return response.OK(c, "Login successful", dto.NewUserResponse(user))
}

// CreateMany handles batch creation of users
func (h *userHandler) CreateMany(c echo.Context) error {
	req := new(dto.BatchCreateUsersRequest)
	if err := c.Bind(req); err != nil {
		return response.BadRequest(c, "Invalid request format")
	}

	if err := c.Validate(req); err != nil {
		return response.ValidationError(c, err)
	}

	users := req.ToModels()
	if err := h.service.CreateUsers(c.Request().Context(), users); err != nil {
		switch {
		case errors.Is(err, service.ErrEmailExists):
			return response.Conflict(c, "One or more users with the provided emails already exist")
		case errors.Is(err, service.ErrEmptyBatch):
			return response.BadRequest(c, "No users provided")
		default:
			return response.InternalError(c, "Failed to create users")
		}
	}

	return response.Created(c, "Users created successfully", dto.NewUserResponseList(users))
}

// FindByFilter handles finding users by filter criteria
func (h *userHandler) FindByFilter(c echo.Context) error {
	req := new(dto.UserFilterRequest)
	if err := c.Bind(req); err != nil {
		return response.BadRequest(c, "Invalid request format")
	}

	// Convert DTO to filter map
	filter := make(map[string]interface{})
	if req.Name != "" {
		filter["name"] = req.Name
	}
	if req.Email != "" {
		filter["email"] = req.Email
	}

	users, err := h.service.FindUsersByFilter(c.Request().Context(), filter, req.Limit, req.Skip)
	if err != nil {
		return response.InternalError(c, "Failed to find users")
	}

	return response.OK(c, "Users found successfully", dto.NewUserResponseList(users))
}

// UpdateMany handles batch update of users
func (h *userHandler) UpdateMany(c echo.Context) error {
	req := new(dto.BatchUpdateUsersRequest)
	if err := c.Bind(req); err != nil {
		return response.BadRequest(c, "Invalid request format")
	}

	if err := c.Validate(req); err != nil {
		return response.ValidationError(c, err)
	}

	if len(req.Updates) == 0 {
		return response.BadRequest(c, "No updates provided")
	}

	// Map to store individual updates for each user
	userUpdates := make(map[string]map[string]interface{})

	// Process each user's updates separately
	for id, updateReq := range req.Updates {
		// Validate the ID
		if _, err := primitive.ObjectIDFromHex(id); err != nil {
			return response.BadRequest(c, fmt.Sprintf("Invalid user ID format: %s", id))
		}

		// Create a map for this user's updates
		updates := make(map[string]interface{})

		// Add each field if it's not empty
		if updateReq.Name != "" {
			updates["name"] = updateReq.Name
		}
		if updateReq.Email != "" {
			updates["email"] = updateReq.Email
		}
		if updateReq.Password != "" {
			// Hash password before updating
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(updateReq.Password), bcrypt.DefaultCost)
			if err != nil {
				return response.InternalError(c, "Failed to process password")
			}
			updates["password"] = string(hashedPassword)
		}

		// Add updated_at timestamp
		updates["updated_at"] = time.Now().UTC()

		// Only add to userUpdates if we have actual updates
		if len(updates) > 0 {
			userUpdates[id] = updates
		}
	}

	// Call service to perform bulk update with individual user updates
	// This uses the Case 1 approach in the service method
	count, err := h.service.UpdateUsersByFilter(c.Request().Context(), userUpdates, nil)
	if err != nil {
		return response.InternalError(c, "Failed to update users")
	}

	return response.OK(c, fmt.Sprintf("Successfully updated %d users", count), map[string]int64{"updated_count": count})
}

// DeleteMany handles batch deletion of users
func (h *userHandler) DeleteMany(c echo.Context) error {
	req := new(dto.BatchDeleteUsersRequest)
	if err := c.Bind(req); err != nil {
		return response.BadRequest(c, "Invalid request format")
	}

	if err := c.Validate(req); err != nil {
		return response.ValidationError(c, err)
	}

	count, err := h.service.DeleteUsersByIDs(c.Request().Context(), req.IDs)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrEmptyBatch):
			return response.BadRequest(c, "No valid user IDs provided")
		default:
			return response.InternalError(c, "Failed to delete users")
		}
	}

	return response.OK(c, "Users deleted successfully", map[string]int64{"deleted_count": count})
}
