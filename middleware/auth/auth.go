package auth

import (
	"context"
	"errors"
	"strings"

	"github.com/jekabolt/protokol"
	"github.com/jekabolt/protokol/adapters"
)

var (
	ErrUnauthorized   = errors.New("unauthorized")
	ErrInvalidToken   = errors.New("invalid token")
	ErrMissingToken   = errors.New("missing authorization token")
)

type contextKey struct{}

// UserFromContext retrieves user info from context.
func UserFromContext(ctx context.Context) (any, bool) {
	user := ctx.Value(contextKey{})
	return user, user != nil
}

// Validator validates tokens and returns user info.
type Validator interface {
	Validate(ctx context.Context, token string) (user any, err error)
}

// ValidatorFunc adapts a function to Validator.
type ValidatorFunc func(ctx context.Context, token string) (any, error)

func (f ValidatorFunc) Validate(ctx context.Context, token string) (any, error) {
	return f(ctx, token)
}

type Middleware struct {
	validator  Validator
	headerName string
	scheme     string
}

type Option func(*Middleware)

func WithHeader(name string) Option {
	return func(m *Middleware) {
		m.headerName = name
	}
}

func WithScheme(scheme string) Option {
	return func(m *Middleware) {
		m.scheme = scheme
	}
}

func New(validator Validator, opts ...Option) *Middleware {
	m := &Middleware{
		validator:  validator,
		headerName: "Authorization",
		scheme:     "Bearer",
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

func (m *Middleware) Wrap(next adapters.Handler) adapters.Handler {
	return adapters.HandlerFunc(func(ctx context.Context, req *protokol.Request) (*protokol.Response, error) {
		values, ok := req.Metadata[m.headerName]
		if !ok || len(values) == 0 {
			return nil, ErrMissingToken
		}

		token := values[0]
		if m.scheme != "" {
			prefix := m.scheme + " "
			if !strings.HasPrefix(token, prefix) {
				return nil, ErrInvalidToken
			}
			token = strings.TrimPrefix(token, prefix)
		}

		user, err := m.validator.Validate(ctx, token)
		if err != nil {
			return nil, ErrUnauthorized
		}

		ctx = context.WithValue(ctx, contextKey{}, user)
		return next.Handle(ctx, req)
	})
}

// APIKey creates a simple API key validator.
func APIKey(validKeys ...string) Validator {
	keySet := make(map[string]struct{}, len(validKeys))
	for _, k := range validKeys {
		keySet[k] = struct{}{}
	}
	return ValidatorFunc(func(ctx context.Context, token string) (any, error) {
		if _, ok := keySet[token]; ok {
			return token, nil
		}
		return nil, ErrInvalidToken
	})
}
