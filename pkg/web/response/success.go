package response

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// Success sends a success response with the given message and data (HTTP 2XX)
func Success(c echo.Context, statusCode int, message string, data interface{}) error {
	return Send(c, statusCode, message, data)
}

// OK sends a 200 OK response with the given message and data
func OK(c echo.Context, message string, data interface{}) error {
	return Success(c, http.StatusOK, message, data)
}

// Created sends a 201 Created response with the given message and data
func Created(c echo.Context, message string, data interface{}) error {
	return Success(c, http.StatusCreated, message, data)
}

// Accepted sends a 202 Accepted response with the given message and data
func Accepted(c echo.Context, message string, data interface{}) error {
	return Success(c, http.StatusAccepted, message, data)
}

// NoContent sends a 204 No Content response
func NoContent(c echo.Context) error {
	return c.NoContent(http.StatusNoContent)
}
