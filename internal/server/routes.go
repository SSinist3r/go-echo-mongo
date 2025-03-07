package server

import (
	"go-echo-mongo/internal/handler"

	"github.com/labstack/echo/v4"
)

// Registry holds all HTTP handlers
type Registry struct {
	handlers []handler.BaseHandler
}

// NewRegistry creates a new handler registry
func NewRegistry() *Registry {
	return &Registry{
		handlers: make([]handler.BaseHandler, 0),
	}
}

// Add registers a new handler in the registry
func (r *Registry) Add(h handler.BaseHandler) {
	r.handlers = append(r.handlers, h)
}

// RegisterAll registers all handlers with Echo
func (r *Registry) RegisterAll(e *echo.Echo) {
	for _, h := range r.handlers {
		h.Register(e)
	}
}
