package dto

import (
	"time"

	"go-echo-mongo/internal/model"
)

// UserResponse represents the user response without sensitive data
type UserResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateUserRequest represents the request body for creating a user
type CreateUserRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=100"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

// UpdateUserRequest represents the request body for updating a user
type UpdateUserRequest struct {
	Name     string `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
	Email    string `json:"email,omitempty" validate:"omitempty,email"`
	Password string `json:"password,omitempty" validate:"omitempty,min=6"`
}

// LoginRequest represents the request body for user login
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

// BatchCreateUsersRequest represents the request body for creating multiple users
type BatchCreateUsersRequest struct {
	Users []CreateUserRequest `json:"users" validate:"required,min=1,dive"`
}

// BatchUpdateUsersRequest represents the request body for updating multiple users
type BatchUpdateUsersRequest struct {
	Updates map[string]UpdateUserRequest `json:"updates" validate:"required,min=1"`
}

// BatchDeleteUsersRequest represents the request body for deleting multiple users
type BatchDeleteUsersRequest struct {
	IDs []string `json:"ids" validate:"required,min=1"`
}

// UserFilterRequest represents the request body for filtering users
type UserFilterRequest struct {
	Name  string `json:"name,omitempty"`
	Email string `json:"email,omitempty"`
	Limit int64  `json:"limit,omitempty"`
	Skip  int64  `json:"skip,omitempty"`
}

// ToModel converts CreateUserRequest to model.User
func (r *CreateUserRequest) ToModel() *model.User {
	return &model.User{
		Name:     r.Name,
		Email:    r.Email,
		Password: r.Password,
	}
}

// ToModel converts UpdateUserRequest to model.User
func (r *UpdateUserRequest) ToModel(existing *model.User) *model.User {
	if r.Name != "" {
		existing.Name = r.Name
	}
	if r.Email != "" {
		existing.Email = r.Email
	}
	if r.Password != "" {
		existing.Password = r.Password
	}
	return existing
}

// FromModel creates a UserResponse from model.User
func (r *UserResponse) FromModel(user *model.User) *UserResponse {
	r.ID = user.ID.Hex()
	r.Name = user.Name
	r.Email = user.Email
	r.CreatedAt = user.CreatedAt
	r.UpdatedAt = user.UpdatedAt
	return r
}

// NewUserResponse creates a new UserResponse from model.User
func NewUserResponse(user *model.User) *UserResponse {
	return new(UserResponse).FromModel(user)
}

// NewUserResponseList creates a slice of UserResponse from a slice of model.User
func NewUserResponseList(users []*model.User) []*UserResponse {
	result := make([]*UserResponse, len(users))
	for i, user := range users {
		result[i] = NewUserResponse(user)
	}
	return result
}

// ToModels converts BatchCreateUsersRequest to a slice of model.User
func (r *BatchCreateUsersRequest) ToModels() []*model.User {
	users := make([]*model.User, len(r.Users))
	for i, userReq := range r.Users {
		users[i] = userReq.ToModel()
	}
	return users
}
