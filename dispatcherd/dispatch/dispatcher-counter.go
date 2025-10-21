package dispatch

import "context"

type CounterDispatcher struct {
	CallsCount int
}

func (c *CounterDispatcher) Dispatch(ctx context.Context, msg *Message) error {
	c.CallsCount++
	return nil
}

func (c *CounterDispatcher) ConfigSchema() map[string]interface{} {
	return map[string]interface{}{}
}

func (c *CounterDispatcher) SetConfig(config map[string]interface{}) {
	// nothing to do
}

func NewCounterDispatcher() *CounterDispatcher {
	return &CounterDispatcher{}
}
