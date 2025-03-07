package response

import (
	"github.com/labstack/echo/v4"
)

// Response represents a standardized API response structure
type Response struct {
	StatusCode int         `json:"status_code"`
	Message    string      `json:"message" example:"Operation completed successfully"`
	Data       interface{} `json:"data,omitempty"`
}

// PaginatedResponse represents a paginated API response
type PaginatedResponse struct {
	Data interface{} `json:"data"`
	Meta struct {
		CurrentPage  int64 `json:"current_page" example:"1"`
		ItemsPerPage int64 `json:"items_per_page" example:"10"`
		TotalItems   int64 `json:"total_items" example:"100"`
		TotalPages   int64 `json:"total_pages" example:"10"`
	} `json:"meta"`
}

// New creates a new JSON response instance
func New(statusCode int, message string, data interface{}) *Response {
	return &Response{
		StatusCode: statusCode,
		Message:    message,
		Data:       data,
	}
}

// Send sends a JSON response with the given status code, message, and data
func Send(c echo.Context, statusCode int, message string, data interface{}) error {
	return c.JSON(statusCode, New(statusCode, message, data))
}
