// Package mwutil provides custom middleware utilities for Echo applications.
package mwutil

import (
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// LoggerConfig defines the config for Logger middleware.
type LoggerConfig struct {
	// Skipper defines a function to skip middleware.
	Skipper middleware.Skipper

	// Format is the log format which can be empty string or JSON
	Format string

	// CustomTimeFormat is the time format for the logs
	CustomTimeFormat string
}

// DefaultLoggerConfig is the default Logger middleware config.
var DefaultLoggerConfig = LoggerConfig{
	Skipper:          middleware.DefaultSkipper,
	Format:           "time=${time_rfc3339}, method=${method}, uri=${uri}, status=${status}, latency=${latency_human}\n",
	CustomTimeFormat: time.RFC3339,
}

// Logger returns a middleware that logs HTTP requests.
func Logger() echo.MiddlewareFunc {
	return LoggerWithConfig(DefaultLoggerConfig)
}

// LoggerWithConfig returns a Logger middleware with config.
func LoggerWithConfig(config LoggerConfig) echo.MiddlewareFunc {
	// Use Echo's built-in logger middleware with our custom config
	return middleware.LoggerWithConfig(middleware.LoggerConfig{
		Skipper:          config.Skipper,
		Format:           config.Format,
		CustomTimeFormat: config.CustomTimeFormat,
	})
}
