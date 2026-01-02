# Adapters

Adapters expose your schema through different protocols. Each adapter translates between its protocol and the internal request/response format.

## Adapter Interface

```go
// github.com/jekabolt/protokol

type Adapter interface {
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    Name() string
}
```

## REST Adapter

The REST adapter exposes your services as HTTP/JSON endpoints.

### Basic Usage

```go
import (
    "github.com/jekabolt/protokol/adapters"
    "github.com/jekabolt/protokol/adapters/rest"
)

adapter := rest.New(rest.Config{
    Config: adapters.Config{
        Schema:   p.Schema(),
        Backends: p.Backends(),
    },
    Listen:     ":8080",
    PathPrefix: "/api/v1",
})

p.AddAdapter(adapter)
```

### Configuration

```go
type Config struct {
    adapters.Config        // Common adapter config
    Listen     string      // Address to listen on (e.g., ":8080")
    PathPrefix string      // URL prefix (e.g., "/api/v1")
}
```

### URL Routing

#### Explicit HTTP Mapping

Methods with explicit HTTP mappings use those paths:

```go
schema.Unary("GetUser").
    HTTP("GET", "/users/{id}").
    Build()
// → GET /api/v1/users/{id}

schema.Unary("CreateUser").
    HTTP("POST", "/users").
    Build()
// → POST /api/v1/users
```

#### Automatic Routing

Methods without explicit HTTP mappings use the default pattern:

```
{prefix}/{ServiceName}/{MethodName}
```

HTTP method is inferred from the method name:
- `Get*`, `List*`, `Find*`, `Search*` → GET
- `Delete*`, `Remove*` → DELETE
- `Update*`, `Patch*` → PUT
- Everything else → POST

```go
schema.Unary("GetUser").Build()     // → GET  /api/v1/UserService/GetUser
schema.Unary("CreateUser").Build()  // → POST /api/v1/UserService/CreateUser
schema.Unary("DeleteUser").Build()  // → DELETE /api/v1/UserService/DeleteUser
```

### Path Parameters

Path parameters are extracted and added to the request input:

```go
// Schema
schema.Unary("GetUser").
    HTTP("GET", "/users/{id}").
    Input(schema.Message("Request").
        RequiredField("id", schema.String).
        Build()).
    Build()

// Request: GET /api/v1/users/123
// Handler receives: input["id"] = "123"
```

### Query Parameters

For GET requests, query parameters are added to the input:

```go
// Schema
schema.Unary("ListUsers").
    HTTP("GET", "/users").
    Input(schema.Message("Request").
        Field("limit", schema.Int32).
        Field("offset", schema.Int32).
        Build()).
    Build()

// Request: GET /api/v1/users?limit=10&offset=20
// Handler receives: input["limit"] = "10", input["offset"] = "20"
```

### Request Body

POST, PUT, and PATCH requests parse JSON body:

```go
// Request: POST /api/v1/users
// Body: {"name": "John", "email": "john@example.com"}
// Handler receives: input["name"] = "John", input["email"] = "john@example.com"
```

### Response Format

Responses are returned as JSON:

```go
// Handler returns:
return map[string]any{
    "id": "123",
    "name": "John",
}, nil

// HTTP Response:
// Content-Type: application/json
// {"id":"123","name":"John"}
```

### Error Responses

Errors are returned with appropriate HTTP status codes:

```go
// 400 Bad Request - Invalid JSON body
{"error": "invalid JSON body"}

// 401 Unauthorized - Auth middleware rejection
{"error": "unauthorized"}

// 429 Too Many Requests - Rate limit exceeded
{"error": "rate limit exceeded"}

// 500 Internal Server Error - Backend errors
{"error": "user not found"}
```

### Accessing Headers

Headers are available in request metadata:

```go
h.Register("UserService", "GetUser", func(ctx context.Context, input map[string]any) (map[string]any, error) {
    // Access via auth middleware context
    // Or check input metadata through a custom middleware
})
```

### Custom Router Access

Access the underlying chi router for custom middleware:

```go
adapter := rest.New(cfg)

// Add chi middleware
adapter.Router().Use(cors.Handler(cors.Options{
    AllowedOrigins: []string{"*"},
}))

p.AddAdapter(adapter)
```

## Adding Middleware

Apply middleware to all requests:

```go
import (
    "github.com/jekabolt/protokol/adapters"
    "github.com/jekabolt/protokol/middleware/logging"
    "github.com/jekabolt/protokol/middleware/recover"
)

adapter := rest.New(rest.Config{
    Config: adapters.Config{
        Schema:   p.Schema(),
        Backends: p.Backends(),
        Middleware: []adapters.Middleware{
            recover.New(logger),
            logging.New(logger),
        },
    },
    Listen: ":8080",
})
```

See [Middleware](middleware.md) for details.

## Running Multiple Adapters

```go
p := protokol.New()

// REST on :8080
p.AddAdapter(rest.New(rest.Config{
    Config:     adapters.Config{Schema: p.Schema(), Backends: p.Backends()},
    Listen:     ":8080",
    PathPrefix: "/api",
}))

// Another REST on :8081 (e.g., admin API)
p.AddAdapter(rest.New(rest.Config{
    Config:     adapters.Config{Schema: p.Schema(), Backends: p.Backends()},
    Listen:     ":8081",
    PathPrefix: "/admin",
}))

// All adapters start when Run is called
p.Run(ctx)
```

## Lifecycle

```go
// Start all adapters
err := p.Run(ctx)

// Stop all adapters (called automatically on context cancellation)
err := p.Stop(ctx)
```

## Future Adapters

Planned adapters:
- **gRPC** - Protocol Buffers over HTTP/2
- **GraphQL** - Query language for APIs
- **WebSocket** - Real-time bidirectional communication
