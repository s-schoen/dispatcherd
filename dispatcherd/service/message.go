package service

import (
	"context"
	"dispatcherd/dispatch"
	"fmt"
	"log/slog"
)

type MessageService interface {
	QueueMessage(ctx context.Context, message *dispatch.Message) error
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

func (s *messageService) QueueMessage(ctx context.Context, message *dispatch.Message) error {
	msgCtx := message.AnnotateContext(ctx)

	s.logger.DebugContext(msgCtx, fmt.Sprintf("received message: %s", message.String()))

	dispatcherNames, err := s.ruleEngine.ProcessMessage(msgCtx, message)
	if err != nil {
		return fmt.Errorf("processing message: %w", err)
	}

	s.logger.DebugContext(msgCtx, fmt.Sprintf("matched: %v", dispatcherNames))

	return nil
}
