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

func (f HandlerFunc) Handle(ctx context.Context, req *protokol.Request) (*protokol.Response, error) {
	return f(ctx, req)
}

func Chain(h Handler, middleware ...Middleware) Handler {
	for i := len(middleware) - 1; i >= 0; i-- {
		h = middleware[i].Wrap(h)
	}
	return h
}
