package protokol

import (
	"context"
	"sync"
)

// Backend represents a service implementation.
type Backend interface {
	Call(ctx context.Context, req *Request) (*Response, error)
	Stream(ctx context.Context, req *Request) (Stream, error)
	Close() error
}

// Request represents an incoming request to a backend.
type Request struct {
	Service    string
	Method     string
	Input      map[string]any
	RawInput   []byte
	Metadata   map[string][]string
	RemoteAddr string // Client IP address from connection
}

// Response represents a backend response.
type Response struct {
	Output    map[string]any
	RawOutput []byte
	Metadata  map[string][]string
}

// Stream represents a bidirectional stream.
type Stream interface {
	Send(msg map[string]any) error
	Recv() (map[string]any, error)
	Close() error
}

// BackendRegistry manages backend instances.
type BackendRegistry struct {
	mu       sync.RWMutex
	backends map[string]Backend
}

// NewBackendRegistry creates a new empty backend registry.
func NewBackendRegistry() *BackendRegistry {
	return &BackendRegistry{
		backends: make(map[string]Backend),
	}
}

// Register adds a backend with the given name to the registry.
func (r *BackendRegistry) Register(name string, b Backend) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.backends[name] = b
}

// Get retrieves a backend by name. Returns false if not found.
func (r *BackendRegistry) Get(name string) (Backend, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	b, ok := r.backends[name]
	return b, ok
}

// Close closes all registered backends and clears the registry.
// Returns the first error encountered. Safe to call multiple times.
func (r *BackendRegistry) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	var firstErr error
	for _, b := range r.backends {
		if err := b.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	clear(r.backends)
	return firstErr
}
