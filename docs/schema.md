# Schema Definition

The schema package provides a fluent API for defining your API structure. Schemas are protocol-agnostic and used by all adapters.

## Types

### Primitive Types

```go
import "github.com/jekabolt/protokol/schema"

schema.String   // string
schema.Int32    // int32
schema.Int64    // int64
schema.Float32  // float32
schema.Float64  // float64
schema.Bool     // bool
schema.Bytes    // []byte
```

### Composite Types

#### Repeated (Arrays)

```go
// []string
schema.Repeated(schema.String)

// []User
schema.Repeated(schema.Message("User").
    Field("id", schema.String).
    Build())
```

#### Maps

```go
// map[string]string
schema.Map(schema.String, schema.String)

// map[string]int64
schema.Map(schema.String, schema.Int64)
```

### Message Types

Messages are structured types with named fields:

```go
userType := schema.Message("User").
    Field("id", schema.String).
    Field("name", schema.String).
    Field("email", schema.String).
    Field("age", schema.Int32).
    Build()
```

#### Required Fields

```go
schema.Message("CreateUserRequest").
    RequiredField("name", schema.String).   // Must be provided
    RequiredField("email", schema.String).  // Must be provided
    Field("nickname", schema.String).       // Optional
    Build()
```

#### Default Values

```go
schema.Message("ListRequest").
    FieldWithDefault("limit", schema.Int32, 10).   // Defaults to 10
    FieldWithDefault("offset", schema.Int32, 0).   // Defaults to 0
    Build()
```

#### Nested Messages

```go
addressType := schema.Message("Address").
    Field("street", schema.String).
    Field("city", schema.String).
    Build()

userType := schema.Message("User").
    Field("id", schema.String).
    Field("name", schema.String).
    Field("address", addressType).  // Nested message
    Build()
```

### Enum Types

```go
statusEnum := schema.Enum("Status").
    Value("PENDING", 0).
    Value("ACTIVE", 1).
    Value("INACTIVE", 2).
    Build()

userType := schema.Message("User").
    Field("id", schema.String).
    Field("status", statusEnum).
    Build()
```

## Methods

Methods define RPC operations on a service.

### Unary Methods

Request-response pattern:

```go
schema.Unary("GetUser").
    Input(schema.Message("GetUserRequest").
        RequiredField("id", schema.String).
        Build()).
    Output(schema.Message("User").
        Field("id", schema.String).
        Field("name", schema.String).
        Build()).
    Build()
```

### Streaming Methods

```go
// Server streaming: one request, multiple responses
schema.ServerStream("ListUsers").
    Input(schema.Message("ListRequest").Build()).
    Output(schema.Message("User").Build()).
    Build()

// Client streaming: multiple requests, one response
schema.ClientStream("UploadChunks").
    Input(schema.Message("Chunk").Build()).
    Output(schema.Message("UploadResult").Build()).
    Build()

// Bidirectional streaming
schema.Bidirectional("Chat").
    Input(schema.Message("ChatMessage").Build()).
    Output(schema.Message("ChatMessage").Build()).
    Build()
```

### HTTP Mapping

Map methods to REST endpoints:

```go
schema.Unary("GetUser").
    HTTP("GET", "/users/{id}").  // GET /users/123
    // ...

schema.Unary("CreateUser").
    HTTP("POST", "/users").      // POST /users
    // ...

schema.Unary("UpdateUser").
    HTTP("PUT", "/users/{id}").  // PUT /users/123
    // ...

schema.Unary("DeleteUser").
    HTTP("DELETE", "/users/{id}"). // DELETE /users/123
    // ...
```

Path parameters (like `{id}`) are automatically extracted and added to the request input.

### Method Options

Add custom metadata:

```go
schema.Unary("AdminOnly").
    Option("auth", "admin").
    Option("rateLimit", 100).
    // ...
```

## Services

Services group related methods:

```go
userService := schema.NewService("UserService").
    Package("users.v1").           // Optional namespace
    Backend("user-backend").       // Backend identifier
    Description("User management").
    Method(schema.Unary("GetUser")./* ... */.Build()).
    Method(schema.Unary("CreateUser")./* ... */.Build()).
    Method(schema.Unary("UpdateUser")./* ... */.Build()).
    Method(schema.Unary("DeleteUser")./* ... */.Build()).
    Method(schema.Unary("ListUsers")./* ... */.Build()).
    MustBuild()
```

### Registering Services

```go
p := protokol.New()
p.Schema().AddService(userService)
p.Schema().AddService(productService)
```

### Looking Up Services

```go
svc, ok := p.Schema().ServiceByName("UserService")
if ok {
    method, ok := svc.MethodByName("GetUser")
}
```

## Complete Example

```go
package main

import "github.com/jekabolt/protokol/schema"

// Define reusable types
var (
    userType = schema.Message("User").
        Field("id", schema.String).
        Field("name", schema.String).
        Field("email", schema.String).
        Field("createdAt", schema.Int64).
        Build()

    addressType = schema.Message("Address").
        Field("street", schema.String).
        Field("city", schema.String).
        Field("country", schema.String).
        Build()
)

// Define service
var UserService = schema.NewService("UserService").
    Backend("users").
    Method(schema.Unary("GetUser").
        Input(schema.Message("GetUserRequest").
            RequiredField("id", schema.String).
            Build()).
        Output(userType).
        HTTP("GET", "/users/{id}").
        Description("Get a user by ID").
        Build()).
    Method(schema.Unary("CreateUser").
        Input(schema.Message("CreateUserRequest").
            RequiredField("name", schema.String).
            RequiredField("email", schema.String).
            Field("address", addressType).
            Build()).
        Output(userType).
        HTTP("POST", "/users").
        Description("Create a new user").
        Build()).
    Method(schema.Unary("ListUsers").
        Input(schema.Message("ListUsersRequest").
            FieldWithDefault("limit", schema.Int32, 20).
            FieldWithDefault("offset", schema.Int32, 0).
            Build()).
        Output(schema.Message("ListUsersResponse").
            Field("users", schema.Repeated(userType)).
            Field("total", schema.Int32).
            Build()).
        HTTP("GET", "/users").
        Description("List all users").
        Build()).
    MustBuild()
```

## Type Reference

| Kind | Go Type | JSON Type |
|------|---------|-----------|
| `KindString` | `string` | `string` |
| `KindInt32` | `int32` | `number` |
| `KindInt64` | `int64` | `number` |
| `KindFloat32` | `float32` | `number` |
| `KindFloat64` | `float64` | `number` |
| `KindBool` | `bool` | `boolean` |
| `KindBytes` | `[]byte` | `string` (base64) |
| `KindMessage` | `struct` | `object` |
| `KindEnum` | `int` | `string` or `number` |
| `KindRepeated` | `[]T` | `array` |
| `KindMap` | `map[K]V` | `object` |
