// Package protokol provides protocol translation capabilities.
// It exposes backend services through multiple API protocols
// (REST, gRPC, GraphQL, WebSocket) simultaneously.
package protokol

import (
	"context"
	"sync"

	"github.com/jekabolt/protokol/schema"
)

const Version = "0.1.0-dev"

// Adapter exposes the schema through a specific protocol.
type Adapter interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Name() string
}

// Protokol is the main orchestrator.
type Protokol struct {
	schema   *schema.Schema
	backends *BackendRegistry
	adapters []Adapter

	mu      sync.RWMutex
	running bool
}

func New() *Protokol {
	return &Protokol{
		schema:   schema.NewSchema(),
		backends: NewBackendRegistry(),
	}
}

func (p *Protokol) Schema() *schema.Schema {
	return p.schema
}

func (p *Protokol) Backends() *BackendRegistry {
	return p.backends
}

func (p *Protokol) AddAdapter(a Adapter) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.adapters = append(p.adapters, a)
}

func (p *Protokol) Run(ctx context.Context) error {
	p.mu.Lock()
	if p.running {
		p.mu.Unlock()
		return ErrAlreadyRunning
	}
	p.running = true
	p.mu.Unlock()

	errCh := make(chan error, len(p.adapters))
	for _, a := range p.adapters {
		go func(adapter Adapter) {
			errCh <- adapter.Start(ctx)
		}(a)
	}

	select {
	case err := <-errCh:
		p.Stop(context.Background())
		return err
	case <-ctx.Done():
		return p.Stop(context.Background())
	}
}

func (p *Protokol) Stop(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	var firstErr error
	for _, a := range p.adapters {
		if err := a.Stop(ctx); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	if err := p.backends.Close(); err != nil && firstErr == nil {
		firstErr = err
	}

	p.running = false
	return firstErr
}
