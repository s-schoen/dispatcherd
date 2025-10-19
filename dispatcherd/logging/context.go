package logging

import (
	"context"
	dispatcherdContext "dispatcherd/context"
	"log/slog"
)

const (
	LoggerFieldRequestID string = "requestId"
	LoggerFieldMessageID string = "messageId"
	LoggerFieldError     string = "error"
)

type ContextHandler struct {
	slog.Handler
}

func (h ContextHandler) Handle(ctx context.Context, r slog.Record) error {
	if val, ok := ctx.Value(dispatcherdContext.KeyRequestID).(string); ok {
		r.AddAttrs(slog.String(LoggerFieldRequestID, val))
	}

	if val, ok := ctx.Value(dispatcherdContext.KeyMessageID).(string); ok {
		r.AddAttrs(slog.String(LoggerFieldMessageID, val))
	}

	return h.Handler.Handle(ctx, r)
}
