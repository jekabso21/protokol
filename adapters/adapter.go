// Package adapters provides common types for protocol adapters.
package adapters

import (
	"context"

	"github.com/jekabolt/protokol"
	"github.com/jekabolt/protokol/schema"
)

// Config is the common adapter configuration.
type Config struct {
	Schema     *schema.Schema
	Backends   *protokol.BackendRegistry
	Middleware []Middleware
}

// Middleware wraps handler logic.
type Middleware interface {
	Wrap(next Handler) Handler
}

// Handler processes a request.
type Handler interface {
	Handle(ctx context.Context, req *protokol.Request) (*protokol.Response, error)
}

// HandlerFunc adapts a function to Handler.
type HandlerFunc func(ctx context.Context, req *protokol.Request) (*protokol.Response, error)

// Handle implements the Handler interface.
func (f HandlerFunc) Handle(ctx context.Context, req *protokol.Request) (*protokol.Response, error) {
	return f(ctx, req)
}

// Chain wraps a handler with middleware in reverse order (last middleware wraps first).
func Chain(h Handler, middleware ...Middleware) Handler {
	for i := len(middleware) - 1; i >= 0; i-- {
		h = middleware[i].Wrap(h)
	}
	return h
}
