# Roadmap

Current status: **early development**

---

## v0.1 — Core Library

The foundation. Get the basics working so you can import the library and expose a single backend through REST and gRPC.

- [ ] Define core types
    - [ ] Service definition struct
    - [ ] Method signatures
    - [ ] Field types (primitives, nested messages, maps, repeated)
    - [ ] Type conversion rules between protocols
- [ ] Configuration
    - [ ] YAML config parsing
    - [ ] Programmatic configuration via Go structs
    - [ ] Validation with clear error messages
- [ ] Adapter interface
    - [ ] Common interface that all protocol adapters implement
    - [ ] Lifecycle hooks (start, stop, health)
    - [ ] Middleware chain support
- [ ] gRPC adapter
    - [ ] Connect to existing gRPC backends
    - [ ] Reflection-based service discovery
    - [ ] Support for unary calls
    - [ ] Support for streaming (server, client, bidirectional)
- [ ] REST adapter
    - [ ] HTTP server with JSON serialization
    - [ ] Automatic path mapping from service methods
    - [ ] Path parameters, query parameters, request body handling
    - [ ] Configurable field naming (snake_case, camelCase)
- [ ] Transformation layer
    - [ ] Convert requests between protocol formats
    - [ ] Handle type mismatches gracefully
    - [ ] Error mapping (gRPC status codes ↔ HTTP status codes)

---

## v0.2 — GraphQL & WebSocket

Add two more protocols. At this point the library covers most common use cases.

- [ ] GraphQL adapter
    - [ ] Generate GraphQL schema from service definitions
    - [ ] Query and mutation support
    - [ ] Subscription support (ties into WebSocket)
    - [ ] Field resolvers mapped to backend methods
    - [ ] Introspection toggle
    - [ ] Query depth and complexity limits
- [ ] WebSocket adapter
    - [ ] Connection management
    - [ ] JSON message framing
    - [ ] Request/response correlation
    - [ ] Subscription delivery
    - [ ] Ping/pong and connection keepalive
- [ ] Schema generation
    - [ ] Export OpenAPI 3.x spec from REST adapter
    - [ ] Export GraphQL SDL from GraphQL adapter
- [ ] Error handling improvements
    - [ ] Consistent error format across all protocols
    - [ ] Custom error mapping hooks
    - [ ] Detailed errors in debug mode, sanitized in production

---

## v0.3 — Production Ready

Make the library suitable for real workloads. Focus on reliability, observability, and security.

- [ ] Middleware system
    - [ ] Authentication middleware (JWT, API key)
    - [ ] Logging middleware with request/response capture
    - [ ] Recovery middleware for panic handling
    - [ ] Custom middleware support
- [ ] Rate limiting
    - [ ] Token bucket implementation
    - [ ] Per-client and per-method limits
    - [ ] Configurable limit exceeded response
- [ ] Observability
    - [ ] Prometheus metrics (request count, latency, errors)
    - [ ] OpenTelemetry tracing integration
    - [ ] Structured logging interface
- [ ] Health checks
    - [ ] Liveness endpoint
    - [ ] Readiness endpoint with backend checks
    - [ ] Exposed via all enabled protocols
- [ ] Connection management
    - [ ] Connection pooling to backends
    - [ ] Configurable timeouts
    - [ ] Graceful shutdown
- [ ] Security
    - [ ] TLS configuration
    - [ ] CORS settings for REST/GraphQL
    - [ ] Request size limits

---

## v0.4 — Advanced Features

Features for more complex setups and developer convenience.

- [ ] Caching
    - [ ] In-memory cache with TTL
    - [ ] Per-method cache policies
    - [ ] Cache key customization
    - [ ] Optional Redis backend
- [ ] Service routing
    - [ ] Route to multiple backends
    - [ ] Method-level backend override
    - [ ] Load balancing between backends
- [ ] Request/response hooks
    - [ ] Transform requests before forwarding
    - [ ] Transform responses before returning
    - [ ] Access to full request context
- [ ] Mock mode
    - [ ] Return fake responses based on schema types
    - [ ] Configurable delays
    - [ ] Useful for testing and development

---

## Future

Ideas that might make it in eventually, depending on demand.

- Additional adapters (tRPC, JSON-RPC, SOAP)
- Schema versioning and compatibility checks
- Traffic recording and replay for debugging
- Code generation from proto files
- Plugin system for custom adapters

---

## Not in Scope

Things this library is not trying to be:

- A full API gateway (use Kong, Envoy for that)
- A service mesh
- A standalone proxy server (it's a library you embed)

---

Contributions and feedback welcome.