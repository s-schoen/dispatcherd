package dispatch

import (
	"context"
)

type Dispatcher interface {
	Name() string
	Dispatch(ctx context.Context, msg *Message) error
}
