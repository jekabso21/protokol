# Middleware

Middleware provides cross-cutting functionality like logging, authentication, and rate limiting. Middleware wraps handlers and can inspect/modify requests and responses.

## Middleware Interface

```go
// github.com/jekabolt/protokol/adapters

type Middleware interface {
    Wrap(next Handler) Handler
}

type Handler interface {
    Handle(ctx context.Context, req *protokol.Request) (*protokol.Response, error)
}
```

## Using Middleware

Add middleware to adapter configuration:

```go
import (
    "github.com/jekabolt/protokol/adapters"
    "github.com/jekabolt/protokol/middleware/logging"
    "github.com/jekabolt/protokol/middleware/recover"
    "github.com/jekabolt/protokol/middleware/ratelimit"
)

middleware := []adapters.Middleware{
    recover.New(logger),              // First: catch panics
    logging.New(logger),              // Second: log requests
    ratelimit.New(10, 20, ratelimit.ByIP), // Third: rate limit
}

adapter := rest.New(rest.Config{
    Config: adapters.Config{
        Schema:     p.Schema(),
        Backends:   p.Backends(),
        Middleware: middleware,
    },
    Listen: ":8080",
})
```

### Execution Order

Middleware executes in order for requests, reverse order for responses:

```
Request  → recover → logging → ratelimit → handler
Response ← recover ← logging ← ratelimit ← handler
```

## Built-in Middleware

### Logging

Logs request duration and errors using `log/slog`.

```go
import "github.com/jekabolt/protokol/middleware/logging"

logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
loggingMiddleware := logging.New(logger)
```

**Log Output:**

```
level=INFO msg="request completed" service=UserService method=GetUser duration=1.234ms
level=ERROR msg="request failed" service=UserService method=GetUser duration=0.5ms error="user not found"
```

**Custom Logger:**

```go
// Use default logger
logging.New(nil)

// Use custom logger
logging.New(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelDebug,
})))
```

### Recover

Catches panics and returns an error instead of crashing.

```go
import "github.com/jekabolt/protokol/middleware/recover"

recoverMiddleware := recover.New(logger)
```

**Features:**
- Catches panics in handlers
- Logs panic with stack trace
- Returns `ErrPanic` to client

**Log Output:**

```
level=ERROR msg="panic recovered" panic="runtime error: index out of range" service=UserService method=GetUser stack="goroutine 1 [running]:..."
```

### Rate Limiting

Token bucket rate limiter.

```go
import "github.com/jekabolt/protokol/middleware/ratelimit"

// 10 requests/second, burst of 20
rateLimiter := ratelimit.New(10, 20, ratelimit.ByIP)
```

**Parameters:**
- `requestsPerSecond`: Sustained rate limit
- `burst`: Maximum burst size
- `keyFunc`: Function to extract rate limit key

**Key Functions:**

```go
ratelimit.ByIP      // Rate limit by client IP (X-Forwarded-For or X-Real-IP)
ratelimit.ByService // Rate limit by service name
ratelimit.ByMethod  // Rate limit by service/method
```

**Custom Key Function:**

```go
// Rate limit by user ID from context
customKeyFunc := func(req *protokol.Request) string {
    if userID, ok := req.Metadata["X-User-ID"]; ok && len(userID) > 0 {
        return userID[0]
    }
    return "anonymous"
}

ratelimit.New(100, 200, customKeyFunc)
```

### Authentication

Token-based authentication with pluggable validators.

```go
import "github.com/jekabolt/protokol/middleware/auth"

// API Key authentication
validator := auth.APIKey("secret-key-1", "secret-key-2")
authMiddleware := auth.New(validator)
```

**Options:**

```go
// Custom header name (default: "Authorization")
auth.New(validator, auth.WithHeader("X-API-Key"))

// Custom scheme (default: "Bearer")
auth.New(validator, auth.WithScheme("Token"))

// No scheme (raw token)
auth.New(validator, auth.WithScheme(""))
```

**API Key Validator:**

```go
validator := auth.APIKey("key1", "key2", "key3")
authMiddleware := auth.New(validator, auth.WithScheme(""))

// Client sends: Authorization: key1
```

**Custom Validator:**

```go
validator := auth.ValidatorFunc(func(ctx context.Context, token string) (any, error) {
    // Validate JWT
    claims, err := validateJWT(token)
    if err != nil {
        return nil, err
    }
    return claims, nil
})

authMiddleware := auth.New(validator)

// Client sends: Authorization: Bearer eyJhbGc...
```

**Accessing User in Handler:**

```go
h.Register("UserService", "GetProfile", func(ctx context.Context, input map[string]any) (map[string]any, error) {
    user, ok := auth.UserFromContext(ctx)
    if !ok {
        return nil, errors.New("not authenticated")
    }

    claims := user.(*JWTClaims)
    return map[string]any{
        "userID": claims.UserID,
        "email":  claims.Email,
    }, nil
})
```

## Creating Custom Middleware

### Basic Structure

```go
package mymiddleware

import (
    "context"

    "github.com/jekabolt/protokol"
    "github.com/jekabolt/protokol/adapters"
)

type Middleware struct {
    // configuration
}

func New(/* params */) *Middleware {
    return &Middleware{}
}

func (m *Middleware) Wrap(next adapters.Handler) adapters.Handler {
    return adapters.HandlerFunc(func(ctx context.Context, req *protokol.Request) (*protokol.Response, error) {
        // Before handler
        // ...

        // Call next handler
        resp, err := next.Handle(ctx, req)

        // After handler
        // ...

        return resp, err
    })
}
```

### Example: Request ID

```go
package requestid

import (
    "context"

    "github.com/google/uuid"
    "github.com/jekabolt/protokol"
    "github.com/jekabolt/protokol/adapters"
)

type contextKey struct{}

func FromContext(ctx context.Context) string {
    if id, ok := ctx.Value(contextKey{}).(string); ok {
        return id
    }
    return ""
}

type Middleware struct{}

func New() *Middleware {
    return &Middleware{}
}

func (m *Middleware) Wrap(next adapters.Handler) adapters.Handler {
    return adapters.HandlerFunc(func(ctx context.Context, req *protokol.Request) (*protokol.Response, error) {
        // Check for existing request ID
        requestID := ""
        if ids, ok := req.Metadata["X-Request-ID"]; ok && len(ids) > 0 {
            requestID = ids[0]
        } else {
            requestID = uuid.New().String()
        }

        // Add to context
        ctx = context.WithValue(ctx, contextKey{}, requestID)

        // Call handler
        resp, err := next.Handle(ctx, req)

        // Add to response metadata
        if resp != nil {
            if resp.Metadata == nil {
                resp.Metadata = make(map[string][]string)
            }
            resp.Metadata["X-Request-ID"] = []string{requestID}
        }

        return resp, err
    })
}
```

### Example: Timeout

```go
package timeout

import (
    "context"
    "time"

    "github.com/jekabolt/protokol"
    "github.com/jekabolt/protokol/adapters"
)

type Middleware struct {
    timeout time.Duration
}

func New(timeout time.Duration) *Middleware {
    return &Middleware{timeout: timeout}
}

func (m *Middleware) Wrap(next adapters.Handler) adapters.Handler {
    return adapters.HandlerFunc(func(ctx context.Context, req *protokol.Request) (*protokol.Response, error) {
        ctx, cancel := context.WithTimeout(ctx, m.timeout)
        defer cancel()

        return next.Handle(ctx, req)
    })
}
```

### Example: Caching

```go
package cache

import (
    "context"
    "sync"
    "time"

    "github.com/jekabolt/protokol"
    "github.com/jekabolt/protokol/adapters"
)

type entry struct {
    resp      *protokol.Response
    expiresAt time.Time
}

type Middleware struct {
    mu    sync.RWMutex
    cache map[string]entry
    ttl   time.Duration
}

func New(ttl time.Duration) *Middleware {
    return &Middleware{
        cache: make(map[string]entry),
        ttl:   ttl,
    }
}

func (m *Middleware) Wrap(next adapters.Handler) adapters.Handler {
    return adapters.HandlerFunc(func(ctx context.Context, req *protokol.Request) (*protokol.Response, error) {
        // Only cache GET-like methods
        if !isReadMethod(req.Method) {
            return next.Handle(ctx, req)
        }

        key := req.Service + "/" + req.Method

        // Check cache
        m.mu.RLock()
        if e, ok := m.cache[key]; ok && time.Now().Before(e.expiresAt) {
            m.mu.RUnlock()
            return e.resp, nil
        }
        m.mu.RUnlock()

        // Call handler
        resp, err := next.Handle(ctx, req)
        if err != nil {
            return nil, err
        }

        // Store in cache
        m.mu.Lock()
        m.cache[key] = entry{
            resp:      resp,
            expiresAt: time.Now().Add(m.ttl),
        }
        m.mu.Unlock()

        return resp, nil
    })
}

func isReadMethod(name string) bool {
    return len(name) >= 3 && (name[:3] == "Get" || name[:4] == "List")
}
```

## Middleware Chaining

Use `adapters.Chain` to combine middleware:

```go
handler := adapters.Chain(
    baseHandler,
    recoverMiddleware,
    loggingMiddleware,
    authMiddleware,
)
```

Equivalent to:

```go
handler := recoverMiddleware.Wrap(
    loggingMiddleware.Wrap(
        authMiddleware.Wrap(
            baseHandler,
        ),
    ),
)
```
