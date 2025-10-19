package dispatch

import (
	"context"
)

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
