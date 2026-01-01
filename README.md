# protokol

A lightweight Go library for exposing backend services through multiple API protocols simultaneously.

Define your service once, serve it via REST, gRPC, GraphQL, or WebSocket.

## Status

🚧 **Early development** — not ready for use yet.

See [ROADMAP.md](ROADMAP.md) for planned features.

## Why

You have a backend service. Different clients want different protocols:
- Mobile app wants REST
- Web app wants GraphQL
- Internal services want gRPC
- Real-time features need WebSocket

Instead of building and maintaining separate API layers, protokol lets you expose one backend through all of them.

## How It Works

1. Point protokol at your existing service (via gRPC reflection or manual definition)
2. Pick which protocols you want to expose
3. Configure each adapter as needed
4. Run

The library handles request translation, type conversion, and error mapping between protocols.

## Idea

```go
package main

import (
    "github.com/jekabso21/protokol"
    "github.com/jekabso21/protokol/adapters/rest"
    "github.com/jekabso21/protokol/adapters/grpc"
    "github.com/jekabso21/protokol/adapters/graphql"
)

func main() {
    p := protokol.New()

    // Connect to your gRPC backend
    p.Source(grpc.Reflect("localhost:50051"))

    // Expose via multiple protocols
    p.Adapter(rest.New(":8080"))
    p.Adapter(graphql.New(":8081"))

    p.Run()
}
```

With config file:

```yaml
source:
  type: grpc
  endpoint: localhost:50051
  reflection: true

adapters:
  rest:
    enabled: true
    listen: :8080
    prefix: /api/v1
    
  graphql:
    enabled: true
    listen: :8081
```

This is the goal. Not there yet.

## Installation

```bash
go get github.com/jekabso21/protokol
```

> ⚠️ Don't actually run this yet — the library isn't functional.

## Features (Planned)

- **Modular** — only import the adapters you need
- **Low overhead** — minimal allocations, small footprint
- **Configurable** — YAML config or programmatic setup
- **Middleware support** — auth, logging, rate limiting
- **Schema generation** — OpenAPI for REST, SDL for GraphQL

## Contributing

The project is in early stages. If you're interested in contributing, check the roadmap and open an issue to discuss.

## License

[MIT](./LICENSE)