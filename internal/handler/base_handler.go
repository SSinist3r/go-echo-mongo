package handler

import "github.com/labstack/echo/v4"

// Handler interface that all handlers must implement
type BaseHandler interface {
	Register(e *echo.Echo)
}
