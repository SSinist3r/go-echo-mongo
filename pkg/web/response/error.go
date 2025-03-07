package response

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// Error sends an error response with the given message (HTTP 4XX, 5XX)
func Error(c echo.Context, statusCode int, message string) error {
	return Send(c, statusCode, message, nil)
}

// BadRequest sends a 400 Bad Request response with the given message
func BadRequest(c echo.Context, message string) error {
	return Error(c, http.StatusBadRequest, message)
}

// ValidationError sends a 400 Bad Request response for validation errors
func ValidationError(c echo.Context, err error) error {
	return BadRequest(c, err.Error())
}

// Unauthorized sends a 401 Unauthorized response with the given message
func Unauthorized(c echo.Context, message string) error {
	return Error(c, http.StatusUnauthorized, message)
}

// Forbidden sends a 403 Forbidden response with the given message
func Forbidden(c echo.Context, message string) error {
	return Error(c, http.StatusForbidden, message)
}

// NotFound sends a 404 Not Found response with the given message
func NotFound(c echo.Context, message string) error {
	return Error(c, http.StatusNotFound, message)
}

// Conflict sends a 409 Conflict response with the given message
func Conflict(c echo.Context, message string) error {
	return Error(c, http.StatusConflict, message)
}

// InternalError sends a 500 Internal Server Error response with the given message
func InternalError(c echo.Context, message string) error {
	return Error(c, http.StatusInternalServerError, message)
}

// NotImplemented sends a 501 Not Implemented response with the given message
func NotImplemented(c echo.Context, message string) error {
	return Error(c, http.StatusNotImplemented, message)
}

// ServiceUnavailable sends a 503 Service Unavailable response with the given message
func ServiceUnavailable(c echo.Context, message string) error {
	return Error(c, http.StatusServiceUnavailable, message)
}
