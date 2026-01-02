# Getting Started

## Installation

```bash
go get github.com/jekabolt/protokol
```

## Basic Setup

### 1. Create a Protokol Instance

```go
import "github.com/jekabolt/protokol"

p := protokol.New()
```

The `Protokol` instance is the central orchestrator that manages:
- Schema definitions
- Backend registry
- Protocol adapters

### 2. Define Your Schema

```go
import "github.com/jekabolt/protokol/schema"

userService := schema.NewService("UserService").
    Backend("users").                    // Backend identifier
    Description("User management").      // Optional description
    Method(schema.Unary("GetUser").
        Input(schema.Message("GetUserRequest").
            RequiredField("id", schema.String).
            Build()).
        Output(schema.Message("User").
            Field("id", schema.String).
            Field("name", schema.String).
            Field("email", schema.String).
            Build()).
        HTTP("GET", "/users/{id}").      // REST mapping
        Build()).
    MustBuild()

p.Schema().AddService(userService)
```

### 3. Implement Your Backend

```go
import "github.com/jekabolt/protokol/backend"

handler := backend.NewHandler()

handler.Register("UserService", "GetUser", func(ctx context.Context, input map[string]any) (map[string]any, error) {
    id := input["id"].(string)

    // Your business logic here
    user := getUserFromDB(id)

    return map[string]any{
        "id":    user.ID,
        "name":  user.Name,
        "email": user.Email,
    }, nil
})

p.Backends().Register("users", handler)
```

### 4. Add an Adapter

```go
import (
    "github.com/jekabolt/protokol/adapters"
    "github.com/jekabolt/protokol/adapters/rest"
)

restAdapter := rest.New(rest.Config{
    Config: adapters.Config{
        Schema:   p.Schema(),
        Backends: p.Backends(),
    },
    Listen:     ":8080",
    PathPrefix: "/api/v1",
})

p.AddAdapter(restAdapter)
```

### 5. Run the Server

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

// Handle shutdown signals
sigCh := make(chan os.Signal, 1)
signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
go func() {
    <-sigCh
    cancel()
}()

if err := p.Run(ctx); err != nil {
    log.Fatal(err)
}
```

## Complete Example

```go
package main

import (
    "context"
    "log"
    "os"
    "os/signal"
    "syscall"

    "github.com/jekabolt/protokol"
    "github.com/jekabolt/protokol/adapters"
    "github.com/jekabolt/protokol/adapters/rest"
    "github.com/jekabolt/protokol/backend"
    "github.com/jekabolt/protokol/schema"
)

func main() {
    p := protokol.New()

    // Schema
    svc := schema.NewService("UserService").
        Backend("users").
        Method(schema.Unary("GetUser").
            Input(schema.Message("Request").RequiredField("id", schema.String).Build()).
            Output(schema.Message("User").
                Field("id", schema.String).
                Field("name", schema.String).
                Build()).
            HTTP("GET", "/users/{id}").
            Build()).
        MustBuild()

    p.Schema().AddService(svc)

    // Backend
    h := backend.NewHandler()
    h.Register("UserService", "GetUser", func(ctx context.Context, in map[string]any) (map[string]any, error) {
        return map[string]any{
            "id":   in["id"],
            "name": "John Doe",
        }, nil
    })
    p.Backends().Register("users", h)

    // Adapter
    p.AddAdapter(rest.New(rest.Config{
        Config:     adapters.Config{Schema: p.Schema(), Backends: p.Backends()},
        Listen:     ":8080",
        PathPrefix: "/api",
    }))

    // Run
    ctx, cancel := context.WithCancel(context.Background())
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
    go func() { <-sigCh; cancel() }()

    log.Println("Starting on :8080")
    if err := p.Run(ctx); err != nil {
        log.Fatal(err)
    }
}
```

## Testing

```bash
# Start the server
go run main.go

# In another terminal
curl http://localhost:8080/api/users/123
# {"id":"123","name":"John Doe"}
```

## Next Steps

- [Schema Definition](schema.md) - Learn about types, methods, and services
- [Backends](backends.md) - Implement your business logic
- [Adapters](adapters.md) - Configure REST and other protocols
- [Middleware](middleware.md) - Add logging, auth, and rate limiting
