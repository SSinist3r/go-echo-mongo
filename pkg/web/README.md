# Web Utilities

This package provides a collection of utilities for building web applications with the Echo framework.

## Packages

- **response**: Standardized API response utilities
- **middleware**: Custom middleware functions
- **validator**: Request validation utilities

## Usage

### Setting Up an Echo Application

```go
package main

import (
	"github.com/labstack/echo/v4"
	"github.com/yourusername/go-echo-mongo/pkg/web/middleware"
	"github.com/yourusername/go-echo-mongo/pkg/web/response"
	"github.com/yourusername/go-echo-mongo/pkg/web/validator"
)

func main() {
	// Create a new Echo instance
	e := echo.New()
	
	// Set up validator
	e.Validator = validator.New()
	
	// Set up middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recovery())
	e.Use(middleware.CORS())
	
	// Define routes
	e.GET("/", func(c echo.Context) error {
		return response.OK(c, "Welcome to the API", nil)
	})
	
	// Protected routes
	g := e.Group("/api")
	g.Use(middleware.JWT("your-secret-key"))
	
	g.GET("/users", GetUsersHandler)
	g.POST("/users", CreateUserHandler)
	
	// Start the server
	e.Logger.Fatal(e.Start(":8080"))
}

func GetUsersHandler(c echo.Context) error {
	// Get users from database
	users := []map[string]interface{}{
		{"id": "1", "name": "John Doe", "email": "john@example.com"},
		{"id": "2", "name": "Jane Smith", "email": "jane@example.com"},
	}
	
	return response.OK(c, "Users retrieved successfully", users)
}

func CreateUserHandler(c echo.Context) error {
	// Bind and validate request
	var req struct {
		Name  string `json:"name" validate:"required"`
		Email string `json:"email" validate:"required,email"`
	}
	
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "Invalid request format")
	}
	
	if err := c.Validate(&req); err != nil {
		return response.ValidationError(c, err)
	}
	
	// Create user in database
	user := map[string]interface{}{
		"id":    "3",
		"name":  req.Name,
		"email": req.Email,
	}
	
	return response.Created(c, "User created successfully", user)
}
```

## Best Practices

1. **Consistent Response Format**: Use the response package to ensure all API responses follow a consistent format.

2. **Proper Error Handling**: Use appropriate error response functions for different error scenarios.

3. **Request Validation**: Always validate incoming requests using the validator package.

4. **Middleware Usage**: Apply middleware in the correct order:
   - Recovery (first to catch panics)
   - Logger (to log all requests)
   - CORS (for cross-origin requests)
   - JWT (for protected routes)

5. **Route Organization**: Group related routes together and apply middleware at the group level when appropriate.

6. **Context Usage**: Use Echo's context to pass data between middleware and handlers.

7. **Dependency Injection**: Pass dependencies to handlers through closures or a dependency container.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request. 