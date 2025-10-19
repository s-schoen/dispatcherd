package service

import (
	"dispatcherd/dispatch"
	"errors"
	"log/slog"

	"github.com/go-playground/validator/v10"
)

var ErrDispatcherNotFound = errors.New("unknown dispatcher")
var ErrDispatcherConfigInvalid = errors.New("invalid dispatcher config")

type DispatcherService interface {
	LoadDispatcherConfig(config dispatch.DispatcherConfig) error
	GetDispatcher(name string) (dispatch.Dispatcher, error)
	GetDefaultDispatchers() ([]dispatch.Dispatcher, error)
}

type dispatcherService struct {
	logger    *slog.Logger
	configs   map[string]dispatch.DispatcherConfig
	validator *validator.Validate
}

func (s dispatcherService) GetDefaultDispatchers() ([]dispatch.Dispatcher, error) {
	defaultDispatchers := make([]dispatch.Dispatcher, 0)
	for _, config := range s.configs {
		if config.IsDefault {
			dispatcher, err := s.GetDispatcher(config.Name)
			if err != nil {
				return nil, err
			}
			defaultDispatchers = append(defaultDispatchers, dispatcher)
		}
	}

	return defaultDispatchers, nil
}

func NewDispatcherService(logger *slog.Logger) DispatcherService {
	svc := dispatcherService{
		logger:    logger,
		configs:   make(map[string]dispatch.DispatcherConfig),
		validator: validator.New(),
	}
	return svc
}

func (s dispatcherService) LoadDispatcherConfig(config dispatch.DispatcherConfig) error {
	// check if dispatcher type exists
	dispatcher, err := s.instantiateDispatcher(config.Type)
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

func (s dispatcherService) GetDispatcher(name string) (dispatch.Dispatcher, error) {
	if config, ok := s.configs[name]; ok {
		dispatcher, err := s.instantiateDispatcher(config.Type)
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

func (s dispatcherService) instantiateDispatcher(dispatcherType string) (dispatch.Dispatcher, error) {
	switch dispatcherType {
	case "log":
		return dispatch.NewLogDispatcher(s.logger), nil
	default:
		return nil, ErrDispatcherNotFound
	}
}
