package ratelimit

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/jekabolt/protokol"
	"github.com/jekabolt/protokol/adapters"
)

var ErrRateLimited = errors.New("rate limit exceeded")

// KeyFunc extracts the rate limit key from a request.
type KeyFunc func(req *protokol.Request) string

// ByService returns a key based on service name.
func ByService(req *protokol.Request) string {
	return req.Service
}

// ByMethod returns a key based on service and method.
func ByMethod(req *protokol.Request) string {
	return req.Service + "/" + req.Method
}

// ByIP returns a key based on X-Forwarded-For or X-Real-IP header.
func ByIP(req *protokol.Request) string {
	if xff, ok := req.Metadata["X-Forwarded-For"]; ok && len(xff) > 0 {
		return xff[0]
	}
	if xri, ok := req.Metadata["X-Real-Ip"]; ok && len(xri) > 0 {
		return xri[0]
	}
	return "unknown"
}

type bucket struct {
	tokens    float64
	lastCheck time.Time
}

type Middleware struct {
	mu       sync.Mutex
	buckets  map[string]*bucket
	rate     float64 // tokens per second
	capacity float64 // max tokens
	keyFunc  KeyFunc
}

func New(requestsPerSecond float64, burst int, keyFunc KeyFunc) *Middleware {
	if keyFunc == nil {
		keyFunc = ByIP
	}
	return &Middleware{
		buckets:  make(map[string]*bucket),
		rate:     requestsPerSecond,
		capacity: float64(burst),
		keyFunc:  keyFunc,
	}
}

func (m *Middleware) Wrap(next adapters.Handler) adapters.Handler {
	return adapters.HandlerFunc(func(ctx context.Context, req *protokol.Request) (*protokol.Response, error) {
		key := m.keyFunc(req)

		if !m.allow(key) {
			return nil, ErrRateLimited
		}

		return next.Handle(ctx, req)
	})
}

func (m *Middleware) allow(key string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	b, ok := m.buckets[key]
	if !ok {
		b = &bucket{tokens: m.capacity, lastCheck: now}
		m.buckets[key] = b
	}

	elapsed := now.Sub(b.lastCheck).Seconds()
	b.tokens += elapsed * m.rate
	if b.tokens > m.capacity {
		b.tokens = m.capacity
	}
	b.lastCheck = now

	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}
