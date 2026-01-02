package rest

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"sync"

	"github.com/go-chi/chi/v5"

	"github.com/jekabolt/protokol"
	"github.com/jekabolt/protokol/adapters"
	"github.com/jekabolt/protokol/middleware/auth"
	"github.com/jekabolt/protokol/middleware/ratelimit"
	"github.com/jekabolt/protokol/schema"
)

// Config for REST adapter.
type Config struct {
	adapters.Config
	Listen     string
	PathPrefix string
}

// Adapter implements REST/HTTP protocol.
type Adapter struct {
	config  Config
	server  *http.Server
	router  chi.Router
	reqPool sync.Pool
}

func New(cfg Config) *Adapter {
	if cfg.Schema == nil {
		panic("rest: schema is required")
	}
	if cfg.Backends == nil {
		panic("rest: backends registry is required")
	}

	a := &Adapter{
		config: cfg,
		router: chi.NewRouter(),
		reqPool: sync.Pool{
			New: func() any {
				return &protokol.Request{
					Input:    make(map[string]any),
					Metadata: make(map[string][]string),
				}
			},
		},
	}
	a.buildRoutes()
	return a
}

func (a *Adapter) Name() string {
	return "rest"
}

func (a *Adapter) Start(ctx context.Context) error {
	a.server = &http.Server{
		Addr:    a.config.Listen,
		Handler: a.router,
	}

	errCh := make(chan error, 1)
	go func() {
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		return a.Stop(context.Background())
	}
}

func (a *Adapter) Stop(ctx context.Context) error {
	if a.server != nil {
		return a.server.Shutdown(ctx)
	}
	return nil
}

func (a *Adapter) Router() chi.Router {
	return a.router
}

func (a *Adapter) buildRoutes() {
	prefix := a.config.PathPrefix
	for _, svc := range a.config.Schema.Services {
		for _, method := range svc.Methods {
			a.registerMethod(prefix, svc, method)
		}
	}
}

func (a *Adapter) registerMethod(prefix string, svc schema.Service, method schema.Method) {
	path := a.methodPath(prefix, svc, method)
	httpMethod := a.httpMethod(method)
	handler := a.makeHandler(svc, method)

	switch httpMethod {
	case http.MethodGet:
		a.router.Get(path, handler)
	case http.MethodPost:
		a.router.Post(path, handler)
	case http.MethodPut:
		a.router.Put(path, handler)
	case http.MethodDelete:
		a.router.Delete(path, handler)
	case http.MethodPatch:
		a.router.Patch(path, handler)
	default:
		a.router.Post(path, handler)
	}
}

func (a *Adapter) methodPath(prefix string, svc schema.Service, method schema.Method) string {
	if method.HTTPPath != "" {
		return prefix + method.HTTPPath
	}
	return prefix + "/" + svc.Name + "/" + method.Name
}

func (a *Adapter) httpMethod(method schema.Method) string {
	if method.HTTPMethod != "" {
		return method.HTTPMethod
	}
	name := method.Name
	switch {
	case hasPrefix(name, "Get", "List", "Find", "Search"):
		return http.MethodGet
	case hasPrefix(name, "Delete", "Remove"):
		return http.MethodDelete
	case hasPrefix(name, "Update", "Patch"):
		return http.MethodPut
	default:
		return http.MethodPost
	}
}

func (a *Adapter) makeHandler(svc schema.Service, method schema.Method) http.HandlerFunc {
	// Build the handler chain: middleware -> backend call
	var handler adapters.Handler = adapters.HandlerFunc(func(ctx context.Context, req *protokol.Request) (*protokol.Response, error) {
		backend, ok := a.config.Backends.Get(svc.Backend)
		if !ok {
			return nil, protokol.ErrBackendNotFound
		}
		return backend.Call(ctx, req)
	})

	// Apply middleware in reverse order
	handler = adapters.Chain(handler, a.config.Middleware...)

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		req := a.reqPool.Get().(*protokol.Request)
		defer func() {
			req.Service = ""
			req.Method = ""
			clear(req.Input)
			clear(req.Metadata)
			req.RawInput = nil
			req.RemoteAddr = ""
			a.reqPool.Put(req)
		}()

		req.Service = svc.Name
		req.Method = method.Name

		if r.Body != nil && r.ContentLength > 0 {
			if err := json.NewDecoder(r.Body).Decode(&req.Input); err != nil {
				a.writeError(w, http.StatusBadRequest, "invalid JSON body")
				return
			}
		}

		a.extractPathParams(r, req)

		if r.Method == http.MethodGet {
			a.extractQueryParams(r, req)
		}

		for k, v := range r.Header {
			req.Metadata[k] = v
		}

		// Set remote address from connection
		req.RemoteAddr = r.RemoteAddr

		resp, err := handler.Handle(ctx, req)
		if err != nil {
			status := a.errorStatus(err)
			a.writeError(w, status, err.Error())
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp.Output)
	}
}

func (a *Adapter) errorStatus(err error) int {
	switch {
	case errors.Is(err, protokol.ErrBackendNotFound):
		return http.StatusInternalServerError
	case errors.Is(err, auth.ErrUnauthorized),
		errors.Is(err, auth.ErrInvalidToken),
		errors.Is(err, auth.ErrMissingToken):
		return http.StatusUnauthorized
	case errors.Is(err, ratelimit.ErrRateLimited):
		return http.StatusTooManyRequests
	default:
		return http.StatusInternalServerError
	}
}

func (a *Adapter) extractPathParams(r *http.Request, req *protokol.Request) {
	rctx := chi.RouteContext(r.Context())
	if rctx == nil {
		return
	}
	for i, key := range rctx.URLParams.Keys {
		if i < len(rctx.URLParams.Values) {
			req.Input[key] = rctx.URLParams.Values[i]
		}
	}
}

func (a *Adapter) extractQueryParams(r *http.Request, req *protokol.Request) {
	for k, v := range r.URL.Query() {
		if len(v) == 1 {
			req.Input[k] = v[0]
		} else {
			req.Input[k] = v
		}
	}
}

func (a *Adapter) writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func hasPrefix(s string, prefixes ...string) bool {
	for _, p := range prefixes {
		if strings.HasPrefix(s, p) {
			return true
		}
	}
	return false
}
