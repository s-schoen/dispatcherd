package service

import (
	"context"
	"dispatcherd/dispatch"
	"dispatcherd/logging"
	"errors"
	"fmt"
	"log/slog"

	"github.com/go-playground/validator/v10"
)

var ErrDispatcherNotFound = errors.New("unknown dispatcher")
var ErrDispatcherConfigInvalid = errors.New("invalid dispatcher config")

type MessageService interface {
	QueueMessage(ctx context.Context, message *dispatch.Message) error
	LoadDispatcherConfig(config dispatch.DispatcherConfig) error
}

type DispatcherFactoryFunc func(logger *slog.Logger, typeName string) (dispatch.Dispatcher, error)

type messageService struct {
	logger            *slog.Logger
	ruleEngine        dispatch.RuleEngine
	configs           map[string]dispatch.DispatcherConfig
	validator         *validator.Validate
	dispatcherFactory DispatcherFactoryFunc
}

func NewMessageService(logger *slog.Logger, ruleEngine dispatch.RuleEngine, factoryFunc DispatcherFactoryFunc) MessageService {
	return &messageService{
		logger:            logger,
		ruleEngine:        ruleEngine,
		configs:           make(map[string]dispatch.DispatcherConfig),
		validator:         validator.New(),
		dispatcherFactory: factoryFunc,
	}
}

func NewDefaultMessageService(logger *slog.Logger, ruleEngine dispatch.RuleEngine) MessageService {
	return NewMessageService(logger, ruleEngine, dispatch.DispatcherFactory)
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
		defaultDispatchers, err := s.getDefaultDispatchers()
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
			dispatcher, err := s.getDispatcherByName(dispatcherName)
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

func (s *messageService) LoadDispatcherConfig(config dispatch.DispatcherConfig) error {
	// check if dispatcher type exists
	dispatcher, err := s.dispatcherFactory(s.logger, config.Type)
	if err != nil {
		return ErrDispatcherNotFound
	}

	// validate config
	validateErrors := s.validator.ValidateMap(config.Config, dispatcher.ConfigSchema())
	if len(validateErrors) > 0 {
		return ErrDispatcherConfigInvalid
	}

	s.configs[config.Name] = config
	return nil
}

func (s *messageService) getDispatcherByName(name string) (dispatch.Dispatcher, error) {
	if config, ok := s.configs[name]; ok {
		dispatcher, err := s.dispatcherFactory(s.logger, config.Type)
		if err != nil {
			return nil, ErrDispatcherNotFound
		}

		// load config into dispatcher
		dispatcher.SetConfig(config.Config)

		return dispatcher, nil
	} else {
		return nil, ErrDispatcherNotFound
	}
}

func (s *messageService) getDefaultDispatchers() ([]dispatch.Dispatcher, error) {
	defaultDispatchers := make([]dispatch.Dispatcher, 0)
	for _, config := range s.configs {
		if config.IsDefault {
			dispatcher, err := s.getDispatcherByName(config.Name)
			if err != nil {
				return nil, err
			}
			defaultDispatchers = append(defaultDispatchers, dispatcher)
		}
	}

	return defaultDispatchers, nil
}
