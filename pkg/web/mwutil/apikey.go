package mwutil

import (
	"context"
	"go-echo-mongo/internal/model"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
)

// APIKeyValidator is an interface for validating API keys and retrieving associated users
type APIKeyValidator interface {
	GetByApiKey(ctx context.Context, apiKey string) (*model.User, error)
}

// Global validator instance
var validator APIKeyValidator

// SetAPIKeyValidator sets the API key validator implementation
func SetAPIKeyValidator(v APIKeyValidator) {
	validator = v
}

// GetAPIKeyValidator returns the current API key validator implementation
func GetAPIKeyValidator() APIKeyValidator {
	return validator
}

// APIKeyAuthConfig defines the config for API key middleware
type APIKeyAuthConfig struct {
	// Skipper defines a function to skip middleware
	Skipper func(c echo.Context) bool

	// KeyLookup is a string in the form of "<source>:<name>" that is used
	// to extract the API key from the request.
	// Default is "header:X-API-Key"
	// Possible values:
	// - "header:<name>"
	// - "query:<name>"
	KeyLookup string

	// Validator is the interface for validating API keys
	Validator APIKeyValidator

	// ErrorHandler is a function to handle API key validation errors
	// If not set, default error handler is used
	ErrorHandler func(c echo.Context, err error) error

	// ContextKey is the key used to store user information in the context
	// Default is "user"
	ContextKey string

	// RequiredRoles specifies which roles are required to access the route
	// If empty, DefaultRole will be used
	RequiredRoles []string
}

// DefaultAPIKeyAuthConfig is the default API key middleware config
var DefaultAPIKeyAuthConfig = APIKeyAuthConfig{
	Skipper:       func(c echo.Context) bool { return false },
	KeyLookup:     "header:X-API-Key",
	ContextKey:    "user",
	RequiredRoles: []string{model.RoleUser},
}

// NewAPIKeyAuth returns a middleware that validates API keys using the global validator
func NewAPIKeyAuth(roles ...string) echo.MiddlewareFunc {
	if GetAPIKeyValidator() == nil {
		log.Fatal("echo: API key validator is not set")
	}
	c := DefaultAPIKeyAuthConfig
	c.Validator = GetAPIKeyValidator()
	if len(roles) > 0 {
		c.RequiredRoles = roles
	} else {
		c.RequiredRoles = []string{model.RoleUser}
	}
	return NewAPIKeyAuthWithConfig(c)
}

// NewAPIKeyAuthWithConfig returns an API key middleware with config
func NewAPIKeyAuthWithConfig(config APIKeyAuthConfig) echo.MiddlewareFunc {
	// Defaults
	if config.Skipper == nil {
		config.Skipper = DefaultAPIKeyAuthConfig.Skipper
	}
	if config.KeyLookup == "" {
		config.KeyLookup = DefaultAPIKeyAuthConfig.KeyLookup
	}
	if config.ContextKey == "" {
		config.ContextKey = DefaultAPIKeyAuthConfig.ContextKey
	}
	if config.Validator == nil {
		log.Fatal("echo: API key validator is required")
	}

	// Initialize
	parts := splitKeyLookup(config.KeyLookup)
	extractKey := extractKeyFromHeader
	switch parts[0] {
	case "header":
		extractKey = extractKeyFromHeader
	case "query":
		extractKey = extractKeyFromQuery
	default:
		panic("echo: invalid API key lookup")
	}

	// Return middleware handler
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}

			// Extract API key
			key, err := extractKey(c, parts[1])
			if err != nil {
				if config.ErrorHandler != nil {
					return config.ErrorHandler(c, err)
				}
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid or missing api key")
			}

			// Validate API key and get user
			user, err := config.Validator.GetByApiKey(c.Request().Context(), key)
			if err != nil {
				if config.ErrorHandler != nil {
					return config.ErrorHandler(c, err)
				}
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid api key")
			}
			// Check if user has any of the required roles
			if len(config.RequiredRoles) > 0 {
				if !user.HasAnyRole(config.RequiredRoles...) {
					return config.ErrorHandler(c, echo.ErrForbidden)
				}
			} else {
				// Default to requiring at least the basic user role
				if !user.HasRole(model.RoleUser) {
					return config.ErrorHandler(c, echo.ErrForbidden)
				}
			}

			// Store user in context
			c.Set(config.ContextKey, user)

			return next(c)
		}
	}
}

// Helper functions

func splitKeyLookup(lookup string) []string {
	parts := make([]string, 2)
	i := 0
	for i < len(lookup) {
		if lookup[i] == ':' {
			parts[0] = lookup[:i]
			parts[1] = lookup[i+1:]
			break
		}
		i++
	}
	return parts
}

func extractKeyFromHeader(c echo.Context, header string) (string, error) {
	key := c.Request().Header.Get(header)
	if key == "" {
		return "", echo.NewHTTPError(http.StatusUnauthorized, "missing api key in header")
	}
	return key, nil
}

func extractKeyFromQuery(c echo.Context, param string) (string, error) {
	key := c.QueryParam(param)
	if key == "" {
		return "", echo.NewHTTPError(http.StatusUnauthorized, "missing api key in query")
	}
	return key, nil
}
