// Package backend provides backend implementations for handling requests.
package backend

import (
	"context"

	"github.com/jekabolt/protokol"
)

// HandlerFunc is the signature for method handlers.
type HandlerFunc func(ctx context.Context, input map[string]any) (map[string]any, error)

// Handler is a backend that routes to Go function handlers.
type Handler struct {
	handlers map[string]map[string]HandlerFunc
}

// NewHandler creates a new Handler with an empty handler registry.
func NewHandler() *Handler {
	return &Handler{
		handlers: make(map[string]map[string]HandlerFunc),
	}
}

// Register adds a handler function for the given service and method.
func (h *Handler) Register(service, method string, fn HandlerFunc) {
	if h.handlers[service] == nil {
		h.handlers[service] = make(map[string]HandlerFunc)
	}
	h.handlers[service][method] = fn
}

// Call invokes the registered handler for the request's service and method.
func (h *Handler) Call(ctx context.Context, req *protokol.Request) (*protokol.Response, error) {
	methods, ok := h.handlers[req.Service]
	if !ok {
		return nil, protokol.ErrServiceNotFound
	}
	fn, ok := methods[req.Method]
	if !ok {
		return nil, protokol.ErrMethodNotFound
	}

	output, err := fn(ctx, req.Input)
	if err != nil {
		return nil, err
	}

	return &protokol.Response{Output: output}, nil
}

// Stream returns ErrStreamingNotSupported as Handler does not support streaming.
func (h *Handler) Stream(ctx context.Context, req *protokol.Request) (protokol.Stream, error) {
	return nil, protokol.ErrStreamingNotSupported
}

// Close is a no-op for Handler as there are no resources to release.
func (h *Handler) Close() error {
	return nil
}
