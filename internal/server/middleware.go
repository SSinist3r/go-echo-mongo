package server

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// setupMiddleware configures all middleware for the server
func setupMiddleware(e *echo.Echo) {
	// Logger middleware logs HTTP requests
	e.Use(middleware.Logger())

	// Recover middleware recovers from panics
	e.Use(middleware.Recover())

	// CORS middleware handles Cross-Origin Resource Sharing
	e.Use(middleware.CORS())
}
