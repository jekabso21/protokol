package ratelimit

import (
	"context"
	"errors"
	"hash/fnv"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/jekabolt/protokol"
	"github.com/jekabolt/protokol/adapters"
)

var ErrRateLimited = errors.New("rate limit exceeded")

const (
	// defaultShards is the number of shards for the bucket map.
	defaultShards = 32
	// defaultCleanupInterval is how often to clean up stale buckets.
	defaultCleanupInterval = time.Minute
	// defaultMaxIdleTime is the maximum time a bucket can be idle before cleanup.
	defaultMaxIdleTime = 5 * time.Minute
)

// KeyFunc extracts the rate limit key from a request.
// Returns an empty string if no key can be determined, which bypasses rate limiting.
type KeyFunc func(req *protokol.Request) string

// ByService returns a key based on service name.
func ByService(req *protokol.Request) string {
	return req.Service
}

// ByMethod returns a key based on service and method.
func ByMethod(req *protokol.Request) string {
	return req.Service + "/" + req.Method
}

// ByIP returns a key based on client IP address.
// Checks headers in order: X-Forwarded-For, X-Real-IP, then falls back to
// RemoteAddr from the connection. If behind a reverse proxy, ensure the
// proxy sets X-Forwarded-For or X-Real-IP headers.
// Returns empty string if no IP can be determined (bypasses rate limiting).
func ByIP(req *protokol.Request) string {
	// Check X-Forwarded-For header (may contain comma-separated list)
	if xff, ok := req.Metadata["X-Forwarded-For"]; ok && len(xff) > 0 {
		// Take the first IP (client IP) from potentially comma-separated list
		ip := strings.TrimSpace(strings.Split(xff[0], ",")[0])
		if ip != "" {
			return ip
		}
	}

	// Check X-Real-IP header
	if xri, ok := req.Metadata["X-Real-Ip"]; ok && len(xri) > 0 {
		ip := strings.TrimSpace(xri[0])
		if ip != "" {
			return ip
		}
	}

	// Fall back to RemoteAddr from connection
	if req.RemoteAddr != "" {
		// RemoteAddr may be "ip:port", extract just the IP
		host, _, err := net.SplitHostPort(req.RemoteAddr)
		if err != nil {
			// No port, use as-is
			return req.RemoteAddr
		}
		return host
	}

	// No IP could be determined
	return ""
}

type bucket struct {
	mu        sync.Mutex
	tokens    float64
	lastCheck time.Time
	lastUsed  time.Time
}

// shard holds a subset of buckets.
type shard struct {
	mu      sync.RWMutex
	buckets map[string]*bucket
}

// Middleware implements token bucket rate limiting with sharding.
type Middleware struct {
	shards   []*shard
	rate     float64 // tokens per second
	capacity float64 // max tokens
	keyFunc  KeyFunc

	cleanupInterval time.Duration
	maxIdleTime     time.Duration
	stopCleanup     chan struct{}
	cleanupDone     chan struct{}
}

// Option configures the Middleware.
type Option func(*Middleware)

// WithCleanupInterval sets how often stale buckets are cleaned up.
func WithCleanupInterval(d time.Duration) Option {
	return func(m *Middleware) {
		m.cleanupInterval = d
	}
}

// WithMaxIdleTime sets how long a bucket can be idle before being removed.
func WithMaxIdleTime(d time.Duration) Option {
	return func(m *Middleware) {
		m.maxIdleTime = d
	}
}

// New creates a rate limiting middleware.
// requestsPerSecond is the sustained rate limit.
// burst is the maximum burst size (bucket capacity).
// keyFunc extracts the rate limit key from requests (nil defaults to ByIP).
func New(requestsPerSecond float64, burst int, keyFunc KeyFunc, opts ...Option) *Middleware {
	if keyFunc == nil {
		keyFunc = ByIP
	}

	m := &Middleware{
		shards:          make([]*shard, defaultShards),
		rate:            requestsPerSecond,
		capacity:        float64(burst),
		keyFunc:         keyFunc,
		cleanupInterval: defaultCleanupInterval,
		maxIdleTime:     defaultMaxIdleTime,
		stopCleanup:     make(chan struct{}),
		cleanupDone:     make(chan struct{}),
	}

	for i := range m.shards {
		m.shards[i] = &shard{
			buckets: make(map[string]*bucket),
		}
	}

	for _, opt := range opts {
		opt(m)
	}

	// Start cleanup goroutine
	go m.cleanupLoop()

	return m
}

// Stop stops the cleanup goroutine. Call this when the middleware is no longer needed.
func (m *Middleware) Stop() {
	close(m.stopCleanup)
	<-m.cleanupDone
}

func (m *Middleware) Wrap(next adapters.Handler) adapters.Handler {
	return adapters.HandlerFunc(func(ctx context.Context, req *protokol.Request) (*protokol.Response, error) {
		key := m.keyFunc(req)

		// Empty key bypasses rate limiting
		if key == "" {
			return next.Handle(ctx, req)
		}

		if !m.allow(key) {
			return nil, ErrRateLimited
		}

		return next.Handle(ctx, req)
	})
}

// getShard returns the shard for a given key.
func (m *Middleware) getShard(key string) *shard {
	h := fnv.New32a()
	h.Write([]byte(key))
	return m.shards[h.Sum32()%uint32(len(m.shards))]
}

func (m *Middleware) allow(key string) bool {
	s := m.getShard(key)
	now := time.Now()

	// Try to get existing bucket with read lock
	s.mu.RLock()
	b, ok := s.buckets[key]
	s.mu.RUnlock()

	if !ok {
		// Need to create bucket, acquire write lock
		s.mu.Lock()
		// Double-check after acquiring write lock
		b, ok = s.buckets[key]
		if !ok {
			b = &bucket{
				tokens:    m.capacity,
				lastCheck: now,
				lastUsed:  now,
			}
			s.buckets[key] = b
		}
		s.mu.Unlock()
	}

	// Lock only this bucket for token operations
	b.mu.Lock()
	defer b.mu.Unlock()

	elapsed := now.Sub(b.lastCheck).Seconds()
	b.tokens += elapsed * m.rate
	if b.tokens > m.capacity {
		b.tokens = m.capacity
	}
	b.lastCheck = now
	b.lastUsed = now

	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

// cleanupLoop periodically removes idle buckets.
func (m *Middleware) cleanupLoop() {
	defer close(m.cleanupDone)

	ticker := time.NewTicker(m.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-m.stopCleanup:
			return
		case <-ticker.C:
			m.cleanup()
		}
	}
}

// cleanup removes buckets that haven't been used within maxIdleTime.
func (m *Middleware) cleanup() {
	now := time.Now()
	cutoff := now.Add(-m.maxIdleTime)

	for _, s := range m.shards {
		var toDelete []string

		// Find stale buckets with read lock
		s.mu.RLock()
		for key, b := range s.buckets {
			b.mu.Lock()
			if b.lastUsed.Before(cutoff) {
				toDelete = append(toDelete, key)
			}
			b.mu.Unlock()
		}
		s.mu.RUnlock()

		// Delete stale buckets with write lock
		if len(toDelete) > 0 {
			s.mu.Lock()
			for _, key := range toDelete {
				// Re-check under write lock in case bucket was accessed
				if b, ok := s.buckets[key]; ok {
					b.mu.Lock()
					if b.lastUsed.Before(cutoff) {
						delete(s.buckets, key)
					}
					b.mu.Unlock()
				}
			}
			s.mu.Unlock()
		}
	}
}
