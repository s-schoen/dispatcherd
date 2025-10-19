package dispatch

import (
	"context"
	"log/slog"
	"math"
)

type LogDispatcher struct {
	logger   *slog.Logger
	logLevel slog.Level
}

func (l *LogDispatcher) SetConfig(config map[string]interface{}) {
	if level, ok := config["level"]; ok {
		// type will be float64 because of json
		levelInt := int(math.Round(level.(float64)))
		l.logLevel = slog.Level(levelInt)
	}
}

func (l *LogDispatcher) ConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"level": "omitempty,min=-4,max=8",
	}
}

func (l *LogDispatcher) Dispatch(ctx context.Context, msg *Message) error {
	l.logger.Log(ctx, l.logLevel, "new message", "messageTitle", msg.Title, "messageMessage", msg.Message)
	return nil
}

func NewLogDispatcher(logger *slog.Logger) *LogDispatcher {
	return &LogDispatcher{
		logger: logger,
	}
}
