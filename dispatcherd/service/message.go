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
	RegisterDispatcher(dispatcher dispatch.Dispatcher, isDefault bool)
}

type messageService struct {
	logger            *slog.Logger
	ruleEngine        dispatch.RuleEngine
	dispatcher        []dispatch.Dispatcher
	defaultDispatcher dispatch.Dispatcher
}

func NewMessageService(logger *slog.Logger, ruleEngine dispatch.RuleEngine) MessageService {
	return &messageService{
		logger:     logger,
		ruleEngine: ruleEngine,
		dispatcher: make([]dispatch.Dispatcher, 0),
	}
}

func (s *messageService) RegisterDispatcher(dispatcher dispatch.Dispatcher, isDefault bool) {
	s.dispatcher = append(s.dispatcher, dispatcher)
	if isDefault {
		s.defaultDispatcher = dispatcher
		s.logger.Info(fmt.Sprintf("default dispatcher set to %s", dispatcher.Name()))
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
		if s.defaultDispatcher != nil {
			if err := s.defaultDispatcher.Dispatch(msgCtx, message); err != nil {
				s.logger.ErrorContext(msgCtx, "failed to dispatch message using default dispatcher", logging.LoggerFieldError, err)
				return fmt.Errorf("dispatching message: %w", err)
			}
			s.logger.InfoContext(msgCtx, "message dispatched using default dispatcher ("+s.defaultDispatcher.Name()+")")
		} else {
			s.logger.WarnContext(msgCtx, "no dispatchers matched, and no default dispatcher is configured")
		}
	} else {
		for _, dispatcherName := range dispatcherNames {
			for _, dispatcher := range s.dispatcher {
				if dispatcher.Name() == dispatcherName {
					s.logger.DebugContext(msgCtx, "dispatching message using "+dispatcherName)
					if err := dispatcher.Dispatch(msgCtx, message); err != nil {
						s.logger.ErrorContext(msgCtx, "failed to dispatch message using "+dispatcherName, logging.LoggerFieldError, err)
						break
					}
					s.logger.DebugContext(msgCtx, "message dispatched using "+dispatcherName)
				}
			}
		}
		s.logger.InfoContext(msgCtx, "message dispatched")
	}

	return nil
}
