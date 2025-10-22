package logging

import (
	"context"
	dispatcherdContext "dispatcherd/context"
	"log/slog"
)

const (
	FieldRequestID string = "requestId"
	FieldMessageID string = "messageId"
	FieldError     string = "error"
)

type ContextHandler struct {
	slog.Handler
}

func (h ContextHandler) Handle(ctx context.Context, r slog.Record) error {
	if val, ok := ctx.Value(dispatcherdContext.KeyRequestID).(string); ok {
		r.AddAttrs(slog.String(FieldRequestID, val))
	}

	if val, ok := ctx.Value(dispatcherdContext.KeyMessageID).(string); ok {
		r.AddAttrs(slog.String(FieldMessageID, val))
	}

	return h.Handler.Handle(ctx, r)
}
