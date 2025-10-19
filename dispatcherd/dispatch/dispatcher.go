package dispatch

import (
	"context"
	"log/slog"
)

type Dispatcher interface {
	Name() string
	Dispatch(ctx context.Context, msg *Message) error
}

type LogDispatcher struct {
	logger *slog.Logger
}

func (l LogDispatcher) Name() string {
	return "log"
}

func (l LogDispatcher) Dispatch(ctx context.Context, msg *Message) error {
	l.logger.InfoContext(ctx, "new message", "messageTitle", msg.Title, "messageMessage", msg.Message)
	return nil
}

func NewLogDispatcher(logger *slog.Logger) *LogDispatcher {
	return &LogDispatcher{
		logger: logger,
	}
}
