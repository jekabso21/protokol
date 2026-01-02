package logging

import (
	"context"
	"log/slog"
	"time"

	"github.com/jekabolt/protokol"
	"github.com/jekabolt/protokol/adapters"
)

type Middleware struct {
	logger *slog.Logger
}

func New(logger *slog.Logger) *Middleware {
	if logger == nil {
		logger = slog.Default()
	}
	return &Middleware{logger: logger}
}

func (m *Middleware) Wrap(next adapters.Handler) adapters.Handler {
	return adapters.HandlerFunc(func(ctx context.Context, req *protokol.Request) (*protokol.Response, error) {
		start := time.Now()

		resp, err := next.Handle(ctx, req)

		duration := time.Since(start)

		attrs := []slog.Attr{
			slog.String("service", req.Service),
			slog.String("method", req.Method),
			slog.Duration("duration", duration),
		}

		if err != nil {
			attrs = append(attrs, slog.String("error", err.Error()))
			m.logger.LogAttrs(ctx, slog.LevelError, "request failed", attrs...)
		} else {
			m.logger.LogAttrs(ctx, slog.LevelInfo, "request completed", attrs...)
		}

		return resp, err
	})
}
