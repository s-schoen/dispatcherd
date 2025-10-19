package service

import (
	"context"
	"dispatcherd/dispatch"
	"dispatcherd/logging"
	"fmt"
	"log/slog"
)

type MessageService interface {
	QueueMessage(ctx context.Context, message *dispatch.Message) error
}

type messageService struct {
	logger            *slog.Logger
	ruleEngine        dispatch.RuleEngine
	dispatcherService DispatcherService
}

func NewMessageService(logger *slog.Logger, ruleEngine dispatch.RuleEngine, dispatcherService DispatcherService) MessageService {
	return &messageService{
		logger:            logger,
		ruleEngine:        ruleEngine,
		dispatcherService: dispatcherService,
	}
}

func (s *messageService) QueueMessage(ctx context.Context, message *dispatch.Message) error {
	msgCtx := message.AnnotateContext(ctx)

	s.logger.DebugContext(msgCtx, fmt.Sprintf("received message: %s", message.String()))

	dispatcherNames, err := s.ruleEngine.ProcessMessage(msgCtx, message)
	if err != nil {
		return fmt.Errorf("processing message: %w", err)
	}

	if len(dispatcherNames) == 0 {
		// use default dispatcher
		defaultDispatchers, err := s.dispatcherService.GetDefaultDispatchers()
		if err != nil {
			s.logger.ErrorContext(msgCtx, "failed to get default dispatchers", logging.LoggerFieldError, err)
			return fmt.Errorf("getting default dispatchers: %w", err)
		}

		if len(defaultDispatchers) > 0 {
			s.invokeDispatchers(msgCtx, message, defaultDispatchers)
			s.logger.InfoContext(msgCtx, "message dispatched using default dispatchers")
		} else {
			s.logger.WarnContext(msgCtx, "no dispatchers matched, and no default dispatcher is configured")
		}
	} else {
		for _, dispatcherName := range dispatcherNames {
			dispatcher, err := s.dispatcherService.GetDispatcher(dispatcherName)
			if err != nil {
				s.logger.ErrorContext(msgCtx, "failed to get dispatcher "+dispatcherName, logging.LoggerFieldError, err)
				return fmt.Errorf("getting dispatcher '%s': %w", dispatcherName, err)
			}
			s.invokeDispatchers(msgCtx, message, []dispatch.Dispatcher{dispatcher})
		}
		s.logger.InfoContext(msgCtx, "message dispatched")
	}

	return nil
}

func (s *messageService) invokeDispatchers(ctx context.Context, message *dispatch.Message, dispatchers []dispatch.Dispatcher) {
	for _, dispatcher := range dispatchers {
		if err := dispatcher.Dispatch(ctx, message); err != nil {
			s.logger.ErrorContext(ctx, "failed to dispatch message", logging.LoggerFieldError, err)
			break
		}
	}
}
