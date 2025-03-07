# Response Utilities

This package provides standardized API response utilities for Echo applications.

## Features

- Consistent JSON response structure
- Helper functions for common HTTP status codes
- Separate modules for success and error responses
- Type-safe response generation

## Usage

### Basic Response Structure

All responses follow a consistent structure:

```json
{
  "status_code": 200,
  "message": "Success message",
  "data": { ... }
}
```

### Success Responses

```go
import (
    "github.com/yourusername/go-echo-mongo/pkg/web/response"
    "github.com/labstack/echo/v4"
)

func GetUserHandler(c echo.Context) error {
    user := User{ID: "123", Name: "John Doe", Email: "john@example.com"}
    return response.OK(c, "User retrieved successfully", user)
}

func CreateUserHandler(c echo.Context) error {
    user := User{ID: "123", Name: "John Doe", Email: "john@example.com"}
    return response.Created(c, "User created successfully", user)
}

func UpdateUserHandler(c echo.Context) error {
    user := User{ID: "123", Name: "John Doe", Email: "john@example.com"}
    return response.OK(c, "User updated successfully", user)
}

func DeleteUserHandler(c echo.Context) error {
    return response.NoContent(c)
}
```

### Error Responses

```go
import (
    "github.com/yourusername/go-echo-mongo/pkg/web/response"
    "github.com/labstack/echo/v4"
)

func GetUserHandler(c echo.Context) error {
    user, err := userService.FindByID(c.Param("id"))
    if err != nil {
        if errors.Is(err, service.ErrUserNotFound) {
            return response.NotFound(c, "User not found")
        }
        return response.InternalError(c, "Failed to retrieve user")
    }
    return response.OK(c, "User retrieved successfully", user)
}

func CreateUserHandler(c echo.Context) error {
    var req CreateUserRequest
    if err := c.Bind(&req); err != nil {
        return response.BadRequest(c, "Invalid request format")
    }
    
    if err := c.Validate(req); err != nil {
        return response.ValidationError(c, err)
    }
    
    if err := userService.Create(c.Request().Context(), req.ToModel()); err != nil {
        if errors.Is(err, service.ErrEmailExists) {
            return response.Conflict(c, "User with this email already exists")
        }
        return response.InternalError(c, "Failed to create user")
    }
    
    return response.Created(c, "User created successfully", user)
}
```

### Custom Status Codes

For status codes not covered by the helper functions:

```go
func CustomResponseHandler(c echo.Context) error {
    // Success response with custom status code
    return response.Success(c, 299, "Custom success message", data)
    
    // Error response with custom status code
    return response.Error(c, 499, "Custom error message")
}
``` 