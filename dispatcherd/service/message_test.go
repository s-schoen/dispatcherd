package service_test

import (
	"context"
	"dispatcherd/dispatch"
	"dispatcherd/service"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TODO: tests are not really meaningful this way, we need to mock the factory in each test individually to improve

func setupMessageService(t *testing.T, re dispatch.RuleEngine, failFactory bool) (service.MessageService, *dispatch.CounterDispatcher) {
	t.Helper()

	dispatcher := dispatch.NewCounterDispatcher()
	factory := func(typeName string) (dispatch.Dispatcher, error) {
		if failFactory {
			return nil, dispatch.ErrUnknownDispatcherType
		}
		return dispatcher, nil
	}

	return service.NewMessageService(re, factory), dispatcher
}

func TestCallDefaultDispatcher(t *testing.T) {
	// mock rule engine to return no dispatchers
	mre := &MockRuleEngine{
		ProcessMessageFunc: func(ctx context.Context, msg *dispatch.Message) ([]string, error) {
			return []string{}, nil
		},
	}

	defaultDispatcher := dispatch.DispatcherConfig{
		Name:      "Test",
		Type:      "mock",
		IsDefault: true,
		Config:    nil,
	}

	messageService, dispatcher := setupMessageService(t, mre, false)
	err := messageService.LoadDispatcherConfig(defaultDispatcher)
	require.NoError(t, err)

	err = messageService.QueueMessage(context.Background(), &dispatch.Message{})

	assert.NoError(t, err)
	assert.Equal(t, 1, dispatcher.CallsCount)
}

func TestCallNonDefaultDispatcher(t *testing.T) {
	mre := &MockRuleEngine{
		ProcessMessageFunc: func(ctx context.Context, msg *dispatch.Message) ([]string, error) {
			return []string{"non-default"}, nil
		},
	}

	defaultDispatcher := dispatch.DispatcherConfig{
		Name:      "non-default",
		Type:      "mock",
		IsDefault: false,
		Config:    nil,
	}

	messageService, dispatcher := setupMessageService(t, mre, false)
	err := messageService.LoadDispatcherConfig(defaultDispatcher)
	require.NoError(t, err)

	err = messageService.QueueMessage(context.Background(), &dispatch.Message{})

	assert.NoError(t, err)
	assert.Equal(t, 1, dispatcher.CallsCount)
}

func TestNoDispatchersFound(t *testing.T) {
	mre := &MockRuleEngine{
		ProcessMessageFunc: func(ctx context.Context, msg *dispatch.Message) ([]string, error) {
			return []string{"test"}, nil
		},
	}

	messageService, _ := setupMessageService(t, mre, true)

	err := messageService.QueueMessage(context.Background(), &dispatch.Message{})
	assert.Error(t, err)
}
