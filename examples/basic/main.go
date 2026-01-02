package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/jekabolt/protokol"
	"github.com/jekabolt/protokol/adapters"
	"github.com/jekabolt/protokol/adapters/rest"
	"github.com/jekabolt/protokol/backend"
	"github.com/jekabolt/protokol/middleware/logging"
	"github.com/jekabolt/protokol/middleware/ratelimit"
	"github.com/jekabolt/protokol/middleware/recover"
	"github.com/jekabolt/protokol/schema"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	p := protokol.New()

	userService := schema.NewService("UserService").
		Backend("users").
		Description("User management service").
		Method(schema.Unary("GetUser").
			Input(schema.Message("GetUserRequest").
				RequiredField("id", schema.String).
				Build()).
			Output(schema.Message("User").
				Field("id", schema.String).
				Field("name", schema.String).
				Field("email", schema.String).
				Build()).
			HTTP("GET", "/users/{id}").
			Build()).
		Method(schema.Unary("CreateUser").
			Input(schema.Message("CreateUserRequest").
				RequiredField("name", schema.String).
				RequiredField("email", schema.String).
				Build()).
			Output(schema.Message("User").
				Field("id", schema.String).
				Field("name", schema.String).
				Field("email", schema.String).
				Build()).
			HTTP("POST", "/users").
			Build()).
		Method(schema.Unary("ListUsers").
			Input(schema.Message("ListUsersRequest").
				FieldWithDefault("limit", schema.Int32, 10).
				Field("offset", schema.Int32).
				Build()).
			Output(schema.Message("ListUsersResponse").
				Field("users", schema.Repeated(schema.Message("User").
					Field("id", schema.String).
					Field("name", schema.String).
					Field("email", schema.String).
					Build())).
				Field("total", schema.Int32).
				Build()).
			HTTP("GET", "/users").
			Build()).
		MustBuild()

	p.Schema().AddService(userService)

	userBackend := backend.NewHandler()

	userBackend.Register("UserService", "GetUser", func(ctx context.Context, input map[string]any) (map[string]any, error) {
		id, _ := input["id"].(string)
		return map[string]any{
			"id":    id,
			"name":  "John Doe",
			"email": "john@example.com",
		}, nil
	})

	userBackend.Register("UserService", "CreateUser", func(ctx context.Context, input map[string]any) (map[string]any, error) {
		return map[string]any{
			"id":    "new-user-123",
			"name":  input["name"],
			"email": input["email"],
		}, nil
	})

	userBackend.Register("UserService", "ListUsers", func(ctx context.Context, input map[string]any) (map[string]any, error) {
		return map[string]any{
			"users": []map[string]any{
				{"id": "1", "name": "Alice", "email": "alice@example.com"},
				{"id": "2", "name": "Bob", "email": "bob@example.com"},
			},
			"total": 2,
		}, nil
	})

	p.Backends().Register("users", userBackend)

	// Configure middleware
	middleware := []adapters.Middleware{
		recover.New(logger),
		logging.New(logger),
		ratelimit.New(10, 20, ratelimit.ByIP), // 10 req/s, burst of 20
	}

	restAdapter := rest.New(rest.Config{
		Config: adapters.Config{
			Schema:     p.Schema(),
			Backends:   p.Backends(),
			Middleware: middleware,
		},
		Listen:     ":8080",
		PathPrefix: "/api/v1",
	})
	p.AddAdapter(restAdapter)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		log.Println("Shutting down...")
		cancel()
	}()

	log.Println("Starting server on :8080")
	log.Println("Try: curl http://localhost:8080/api/v1/users/123")

	if err := p.Run(ctx); err != nil {
		log.Fatal(err)
	}
}
