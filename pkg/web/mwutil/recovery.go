package mwutil

import (
	"fmt"
	"net/http"
	"runtime"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// RecoveryConfig defines the config for Recovery middleware.
type RecoveryConfig struct {
	// Skipper defines a function to skip middleware.
	Skipper func(c echo.Context) bool

	// LogLevel is the log level for the error.
	// Default is error.
	LogLevel string

	// LogErrorFunc is a function which is called when a panic occurs.
	// Default is to log the stack trace at error level.
	LogErrorFunc func(c echo.Context, err error, stack []byte)
}

// DefaultRecoveryConfig is the default Recovery middleware config.
var DefaultRecoveryConfig = RecoveryConfig{
	Skipper:  func(c echo.Context) bool { return false },
	LogLevel: "error",
	LogErrorFunc: func(c echo.Context, err error, stack []byte) {
		fmt.Printf("[PANIC RECOVERED] %v %s\n", err, stack)
	},
}

// Recovery returns a middleware which recovers from panics anywhere in the chain
// and handles the control to the centralized HTTPErrorHandler.
func Recovery() echo.MiddlewareFunc {
	return RecoveryWithConfig(DefaultRecoveryConfig)
}

// RecoveryWithConfig returns a Recovery middleware with config.
func RecoveryWithConfig(config RecoveryConfig) echo.MiddlewareFunc {
	// Use Echo's built-in recovery middleware
	return middleware.RecoverWithConfig(middleware.RecoverConfig{
		Skipper: config.Skipper,
		LogErrorFunc: func(c echo.Context, err error, stack []byte) error {
			if config.LogErrorFunc != nil {
				config.LogErrorFunc(c, err, stack)
			}
			return nil
		},
	})
}

// CustomRecovery returns a middleware which recovers from panics anywhere in the chain
// and handles the control to the centralized HTTPErrorHandler.
// This is a custom implementation that doesn't rely on Echo's built-in middleware.
func CustomRecovery() echo.MiddlewareFunc {
	return CustomRecoveryWithConfig(DefaultRecoveryConfig)
}

// CustomRecoveryWithConfig returns a custom Recovery middleware with config.
func CustomRecoveryWithConfig(config RecoveryConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}

			defer func() {
				if r := recover(); r != nil {
					err, ok := r.(error)
					if !ok {
						err = fmt.Errorf("%v", r)
					}

					// Get stack trace
					stack := make([]byte, 4<<10) // 4KB
					length := runtime.Stack(stack, false)
					stack = stack[:length]

					// Log the error
					if config.LogErrorFunc != nil {
						config.LogErrorFunc(c, err, stack)
					}

					// Send error response
					c.Error(echo.NewHTTPError(http.StatusInternalServerError, "Internal Server Error"))
				}
			}()

			return next(c)
		}
	}
}
