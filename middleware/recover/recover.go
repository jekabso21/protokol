package recover

import (
	"context"
	"errors"
	"log/slog"
	"runtime/debug"

	"github.com/jekabolt/protokol"
	"github.com/jekabolt/protokol/adapters"
)

var ErrPanic = errors.New("internal server error")

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
	return adapters.HandlerFunc(func(ctx context.Context, req *protokol.Request) (resp *protokol.Response, err error) {
		defer func() {
			if r := recover(); r != nil {
				m.logger.ErrorContext(ctx, "panic recovered",
					slog.Any("panic", r),
					slog.String("service", req.Service),
					slog.String("method", req.Method),
					slog.String("stack", string(debug.Stack())),
				)
				resp = nil
				err = ErrPanic
			}
		}()

		return next.Handle(ctx, req)
	})
}
