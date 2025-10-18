package service

import (
	"context"
	"dispatcherd/dispatch"
	"fmt"
	"log/slog"
)

type MessageService interface {
	QueueMessage(ctx context.Context, message dispatch.Message) error
}

type messageService struct {
	logger     *slog.Logger
	ruleEngine dispatch.RuleEngine
}

func NewMessageService(logger *slog.Logger, ruleEngine dispatch.RuleEngine) MessageService {
	return &messageService{
		logger:     logger,
		ruleEngine: ruleEngine,
	}
}

func (s *messageService) QueueMessage(ctx context.Context, message dispatch.Message) error {
	tags := "None"
	if message.Tags != nil {
		tags = ""
		for k, v := range message.Tags {
			tags += fmt.Sprintf("%s:%s;", k, v)
		}
	}
	s.logger.DebugContext(ctx, fmt.Sprintf("received message: title='%s' message='%s' tags='%s'",
		message.Title, message.Message, tags))

	dispatcherNames, err := s.ruleEngine.ProcessMessage(ctx, &message)
	if err != nil {
		return fmt.Errorf("processing message: %w", err)
	}

	s.logger.DebugContext(ctx, fmt.Sprintf("matched: %v", dispatcherNames))

	return nil
}
