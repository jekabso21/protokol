# Backends

Backends implement the business logic for your services. Each service in the schema maps to a backend by name.

## Backend Interface

```go
// github.com/jekabolt/protokol

type Backend interface {
    Call(ctx context.Context, req *Request) (*Response, error)
    Stream(ctx context.Context, req *Request) (Stream, error)
    Close() error
}
```

## Request and Response

```go
type Request struct {
    Service  string              // Service name (e.g., "UserService")
    Method   string              // Method name (e.g., "GetUser")
    Input    map[string]any      // Decoded request data
    RawInput []byte              // Raw bytes (for passthrough)
    Metadata map[string][]string // Headers/metadata
}

type Response struct {
    Output    map[string]any      // Response data
    RawOutput []byte              // Raw bytes (for passthrough)
    Metadata  map[string][]string // Response headers
}
```

## Handler Backend

The `backend.Handler` is the simplest way to implement a backend:

```go
import "github.com/jekabolt/protokol/backend"

h := backend.NewHandler()

h.Register("UserService", "GetUser", func(ctx context.Context, input map[string]any) (map[string]any, error) {
    id := input["id"].(string)

    user, err := db.GetUser(id)
    if err != nil {
        return nil, err
    }

    return map[string]any{
        "id":    user.ID,
        "name":  user.Name,
        "email": user.Email,
    }, nil
})
```

### Registering Handlers

```go
h.Register(serviceName, methodName, handlerFunc)
```

- `serviceName`: Must match the service name in your schema
- `methodName`: Must match the method name in your schema
- `handlerFunc`: Function that processes the request

### Handler Function Signature

```go
type HandlerFunc func(ctx context.Context, input map[string]any) (map[string]any, error)
```

- `ctx`: Request context (contains deadlines, cancellation, values)
- `input`: Decoded request data as a map
- Returns: Response data as a map, or an error

## Working with Input

### Accessing Fields

```go
h.Register("UserService", "GetUser", func(ctx context.Context, input map[string]any) (map[string]any, error) {
    // String field
    id, ok := input["id"].(string)
    if !ok {
        return nil, errors.New("id is required")
    }

    // Integer field
    limit, _ := input["limit"].(float64) // JSON numbers are float64
    limitInt := int(limit)

    // Boolean field
    active, _ := input["active"].(bool)

    // Nested object
    if address, ok := input["address"].(map[string]any); ok {
        city := address["city"].(string)
    }

    // Array
    if tags, ok := input["tags"].([]any); ok {
        for _, tag := range tags {
            tagStr := tag.(string)
        }
    }
})
```

### Type Assertions

JSON decoding produces these Go types:

| JSON Type | Go Type |
|-----------|---------|
| `string` | `string` |
| `number` | `float64` |
| `boolean` | `bool` |
| `null` | `nil` |
| `array` | `[]any` |
| `object` | `map[string]any` |

## Returning Responses

### Simple Response

```go
return map[string]any{
    "id":    "123",
    "name":  "John",
    "email": "john@example.com",
}, nil
```

### Nested Objects

```go
return map[string]any{
    "user": map[string]any{
        "id":   "123",
        "name": "John",
    },
    "metadata": map[string]any{
        "createdAt": time.Now().Unix(),
    },
}, nil
```

### Arrays

```go
return map[string]any{
    "users": []map[string]any{
        {"id": "1", "name": "Alice"},
        {"id": "2", "name": "Bob"},
    },
    "total": 2,
}, nil
```

## Error Handling

### Returning Errors

```go
h.Register("UserService", "GetUser", func(ctx context.Context, input map[string]any) (map[string]any, error) {
    id := input["id"].(string)

    user, err := db.GetUser(id)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, errors.New("user not found")
        }
        return nil, fmt.Errorf("database error: %w", err)
    }

    return map[string]any{"id": user.ID, "name": user.Name}, nil
})
```

### Custom Error Types

```go
type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// Usage
return nil, &ValidationError{Field: "email", Message: "invalid format"}
```

## Context Usage

### Accessing Metadata

```go
h.Register("UserService", "GetUser", func(ctx context.Context, input map[string]any) (map[string]any, error) {
    // Access user from auth middleware
    if user, ok := auth.UserFromContext(ctx); ok {
        // user is authenticated
    }
})
```

### Respecting Cancellation

```go
h.Register("UserService", "LongOperation", func(ctx context.Context, input map[string]any) (map[string]any, error) {
    for i := 0; i < 100; i++ {
        select {
        case <-ctx.Done():
            return nil, ctx.Err()
        default:
            // Do work
        }
    }
    return map[string]any{"done": true}, nil
})
```

## Backend Registry

Register backends with the protokol instance:

```go
p := protokol.New()

// Register multiple backends
p.Backends().Register("users", userHandler)
p.Backends().Register("products", productHandler)
p.Backends().Register("orders", orderHandler)
```

### Multi-Backend Routing

Each service can route to a different backend:

```go
// Schema
userService := schema.NewService("UserService").
    Backend("users").  // Routes to "users" backend
    // ...

productService := schema.NewService("ProductService").
    Backend("products").  // Routes to "products" backend
    // ...

// Backends
p.Backends().Register("users", userHandler)
p.Backends().Register("products", productHandler)
```

## Custom Backend Implementation

Implement the `Backend` interface for custom backends:

```go
type GRPCBackend struct {
    conn *grpc.ClientConn
}

func (b *GRPCBackend) Call(ctx context.Context, req *protokol.Request) (*protokol.Response, error) {
    // Forward request to gRPC service
    // ...
}

func (b *GRPCBackend) Stream(ctx context.Context, req *protokol.Request) (protokol.Stream, error) {
    // Handle streaming
    // ...
}

func (b *GRPCBackend) Close() error {
    return b.conn.Close()
}

// Usage
p.Backends().Register("external-service", &GRPCBackend{conn: conn})
```

## Complete Example

```go
package main

import (
    "context"
    "errors"
    "sync"

    "github.com/jekabolt/protokol/backend"
)

// In-memory user store
type UserStore struct {
    mu    sync.RWMutex
    users map[string]User
}

type User struct {
    ID    string
    Name  string
    Email string
}

func NewUserBackend(store *UserStore) *backend.Handler {
    h := backend.NewHandler()

    h.Register("UserService", "GetUser", func(ctx context.Context, input map[string]any) (map[string]any, error) {
        id := input["id"].(string)

        store.mu.RLock()
        user, ok := store.users[id]
        store.mu.RUnlock()

        if !ok {
            return nil, errors.New("user not found")
        }

        return map[string]any{
            "id":    user.ID,
            "name":  user.Name,
            "email": user.Email,
        }, nil
    })

    h.Register("UserService", "CreateUser", func(ctx context.Context, input map[string]any) (map[string]any, error) {
        user := User{
            ID:    generateID(),
            Name:  input["name"].(string),
            Email: input["email"].(string),
        }

        store.mu.Lock()
        store.users[user.ID] = user
        store.mu.Unlock()

        return map[string]any{
            "id":    user.ID,
            "name":  user.Name,
            "email": user.Email,
        }, nil
    })

    return h
}
```
