package mwutil

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

// JWTConfig defines the config for JWT middleware.
type JWTConfig struct {
	// Skipper defines a function to skip middleware.
	Skipper func(c echo.Context) bool

	// Secret is the key used for validating the JWT token.
	Secret string

	// TokenLookup is a string in the form of "<source>:<name>" that is used
	// to extract token from the request.
	// Default is "header:Authorization"
	TokenLookup string

	// AuthScheme is a string that defines the authorization scheme.
	// Default is "Bearer"
	AuthScheme string

	// ContextKey is the key used to store user information from the JWT token
	// in the echo.Context.
	// Default is "user"
	ContextKey string
}

// DefaultJWTConfig is the default JWT middleware config.
var DefaultJWTConfig = JWTConfig{
	Skipper:     func(c echo.Context) bool { return false },
	TokenLookup: "header:Authorization",
	AuthScheme:  "Bearer",
	ContextKey:  "user",
}

// JWTWithConfig returns a JWT middleware with config.
func JWTWithConfig(config JWTConfig) echo.MiddlewareFunc {
	// Return a middleware handler
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}

			// Extract token
			parts := strings.Split(config.TokenLookup, ":")
			if len(parts) != 2 {
				return echo.NewHTTPError(http.StatusInternalServerError, "invalid token lookup format")
			}

			var token string
			switch parts[0] {
			case "header":
				auth := c.Request().Header.Get(parts[1])
				if auth == "" {
					return echo.NewHTTPError(http.StatusUnauthorized, "missing or malformed jwt")
				}
				if config.AuthScheme != "" {
					l := len(config.AuthScheme)
					if len(auth) > l+1 && auth[:l] == config.AuthScheme {
						token = auth[l+1:]
					} else {
						return echo.NewHTTPError(http.StatusUnauthorized, "invalid auth scheme")
					}
				} else {
					token = auth
				}
			case "query":
				token = c.QueryParam(parts[1])
			case "cookie":
				cookie, err := c.Cookie(parts[1])
				if err != nil {
					return echo.NewHTTPError(http.StatusUnauthorized, "missing or malformed jwt")
				}
				token = cookie.Value
			default:
				return echo.NewHTTPError(http.StatusInternalServerError, "invalid token lookup source")
			}

			// Validate token
			// This is a placeholder for actual JWT validation
			// In a real implementation, you would use a JWT library to validate the token
			if token == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "missing or malformed jwt")
			}

			// For demonstration purposes, we're just checking if the token is not empty
			// and setting a dummy user in the context
			// In a real implementation, you would decode and validate the token
			c.Set(config.ContextKey, map[string]interface{}{
				"id":    "user-123",
				"email": "user@example.com",
				"roles": []string{"user"},
			})

			return next(c)
		}
	}
}

// JWT returns a middleware that validates JWT tokens.
func JWT(secret string) echo.MiddlewareFunc {
	config := DefaultJWTConfig
	config.Secret = secret
	return JWTWithConfig(config)
}
