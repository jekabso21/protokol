# protokol Documentation

protokol is a Go library for exposing backend services through multiple API protocols simultaneously.

## Table of Contents

1. [Getting Started](getting-started.md)
2. [Schema Definition](schema.md)
3. [Backends](backends.md)
4. [Adapters](adapters.md)
5. [Middleware](middleware.md)

## Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                          protokol                               │
│                                                                 │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────────────┐ │
│  │   Schema    │───▶│  Middleware │───▶│      Adapters       │ │
│  │ (Services)  │    │   (Chain)   │    │  (REST/gRPC/GQL/WS) │ │
│  └─────────────┘    └─────────────┘    └─────────────────────┘ │
│         │                                        │              │
│         ▼                                        ▼              │
│  ┌─────────────┐                         ┌─────────────┐       │
│  │  Backends   │◀────────────────────────│   Clients   │       │
│  │ (Handlers)  │                         │             │       │
│  └─────────────┘                         └─────────────┘       │
└─────────────────────────────────────────────────────────────────┘
```

## Core Concepts

### Schema

Define your API structure using the fluent builder API:

```go
service := schema.NewService("UserService").
    Backend("users").
    Method(schema.Unary("GetUser").
        Input(schema.Message("GetUserRequest").
            RequiredField("id", schema.String).
            Build()).
        Output(schema.Message("User").
            Field("id", schema.String).
            Field("name", schema.String).
            Build()).
        HTTP("GET", "/users/{id}").
        Build()).
    MustBuild()
```

### Backends

Implement your business logic:

```go
handler := backend.NewHandler()
handler.Register("UserService", "GetUser", func(ctx context.Context, input map[string]any) (map[string]any, error) {
    id := input["id"].(string)
    return map[string]any{"id": id, "name": "John"}, nil
})
```

### Adapters

Expose your service through different protocols:

```go
restAdapter := rest.New(rest.Config{
    Config: adapters.Config{
        Schema:   p.Schema(),
        Backends: p.Backends(),
    },
    Listen:     ":8080",
    PathPrefix: "/api/v1",
})
```

### Middleware

Add cross-cutting concerns:

```go
middleware := []adapters.Middleware{
    recover.New(logger),
    logging.New(logger),
    ratelimit.New(10, 20, ratelimit.ByIP),
}
```

## Quick Example

```go
package main

import (
    "context"
    "log"

    "github.com/jekabolt/protokol"
    "github.com/jekabolt/protokol/adapters"
    "github.com/jekabolt/protokol/adapters/rest"
    "github.com/jekabolt/protokol/backend"
    "github.com/jekabolt/protokol/schema"
)

func main() {
    p := protokol.New()

    // Define schema
    svc := schema.NewService("HelloService").
        Backend("hello").
        Method(schema.Unary("SayHello").
            Input(schema.Message("HelloRequest").Field("name", schema.String).Build()).
            Output(schema.Message("HelloResponse").Field("message", schema.String).Build()).
            HTTP("GET", "/hello/{name}").
            Build()).
        MustBuild()

    p.Schema().AddService(svc)

    // Implement handler
    h := backend.NewHandler()
    h.Register("HelloService", "SayHello", func(ctx context.Context, in map[string]any) (map[string]any, error) {
        return map[string]any{"message": "Hello, " + in["name"].(string)}, nil
    })
    p.Backends().Register("hello", h)

    // Start REST server
    p.AddAdapter(rest.New(rest.Config{
        Config:     adapters.Config{Schema: p.Schema(), Backends: p.Backends()},
        Listen:     ":8080",
        PathPrefix: "/api",
    }))

    log.Println("Server starting on :8080")
    p.Run(context.Background())
}
```

```bash
curl http://localhost:8080/api/hello/World
# {"message":"Hello, World"}
```
