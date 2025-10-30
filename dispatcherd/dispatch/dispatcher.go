package dispatch

import (
	"context"
	"errors"
)

var ErrUnknownDispatcherType = errors.New("unknown dispatcher type")

type Dispatcher interface {
	Dispatch(ctx context.Context, msg *Message) error
	ConfigSchema() map[string]interface{}
	SetConfig(config map[string]interface{})
}

type DispatcherConfig struct {
	Name      string                 `json:"name" validate:"required"`
	Type      string                 `json:"type" validate:"required"`
	IsDefault bool                   `json:"isDefault"`
	Config    map[string]interface{} `json:"config"`
}

func DispatcherFactory(dispatcherType string) (Dispatcher, error) {
	switch dispatcherType {
	case "log":
		return NewLogDispatcher(), nil
	case "counter":
		return NewCounterDispatcher(), nil
	case "mail":
		return NewMailDispatcher(), nil
	default:
		return nil, ErrUnknownDispatcherType
	}
}
