# Validator Utilities

This package provides validation utilities for Echo applications using the go-playground/validator package.

## Features

- Custom validator implementation for Echo
- Human-readable error messages
- Support for custom validation rules
- JSON field name mapping

## Usage

### Basic Setup

```go
import (
    "github.com/yourusername/go-echo-mongo/pkg/web/validator"
    "github.com/labstack/echo/v4"
)

func main() {
    e := echo.New()
    
    // Set up the validator
    e.Validator = validator.New()
    
    // ... rest of your Echo setup
}
```

### Request Validation

```go
type CreateUserRequest struct {
    Name     string `json:"name" validate:"required,min=2,max=50"`
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=8"`
    Age      int    `json:"age" validate:"required,min=18"`
    Role     string `json:"role" validate:"required,oneof=admin user guest"`
}

func CreateUserHandler(c echo.Context) error {
    var req CreateUserRequest
    
    // Bind request body to struct
    if err := c.Bind(&req); err != nil {
        return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
    }
    
    // Validate the request
    if err := c.Validate(&req); err != nil {
        return err // The validator will return a properly formatted error
    }
    
    // Process the validated request
    // ...
    
    return c.JSON(http.StatusCreated, map[string]string{
        "message": "User created successfully",
    })
}
```

### Custom Validation Rules

```go
import (
    "github.com/yourusername/go-echo-mongo/pkg/web/validator"
    "github.com/go-playground/validator/v10"
    "github.com/labstack/echo/v4"
    "regexp"
)

func main() {
    e := echo.New()
    
    // Create a new validator
    v := validator.New()
    
    // Register a custom validation rule
    v.RegisterCustomValidation("phone", func(fl validator.FieldLevel) bool {
        // Simple phone number validation
        re := regexp.MustCompile(`^\+?[0-9]{10,15}$`)
        return re.MatchString(fl.Field().String())
    })
    
    e.Validator = v
    
    // ... rest of your Echo setup
}

// Then in your request struct:
type ContactRequest struct {
    Name  string `json:"name" validate:"required"`
    Phone string `json:"phone" validate:"required,phone"`
}
```

### Validation Error Response

The validator will return errors in the following format:

```json
{
  "message": {
    "name": "This field is required",
    "email": "Invalid email format",
    "password": "Must be at least 8 characters long",
    "age": "Must be at least 18",
    "role": "Must be one of: admin user guest"
  }
}
``` 