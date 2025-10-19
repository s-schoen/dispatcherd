package service_test

import (
	"bytes"
	"context"
	"dispatcherd/dispatch"
	"dispatcherd/service"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func setupMessageService(t *testing.T, re dispatch.RuleEngine) service.MessageService {
	t.Helper()

	var logBuffer bytes.Buffer
	mockLogger := slog.New(slog.NewJSONHandler(&logBuffer, nil))

	return service.NewMessageService(mockLogger, re)
}

func TestCallDefaultDispatcher(t *testing.T) {
	// mock rule engine to return no dispatchers
	mre := &MockRuleEngine{
		ProcessMessageFunc: func(ctx context.Context, msg *dispatch.Message) ([]string, error) {
			return []string{}, nil
		},
	}

	defaultDispatcher := &MockDispatcher{
		DispatchFunc: func(ctx context.Context, msg *dispatch.Message) error {
			return nil
		},
		NameFunc: func() string {
			return "default"
		},
	}

	sut := setupMessageService(t, mre)
	sut.RegisterDispatcher(defaultDispatcher, true)
	err := sut.QueueMessage(context.Background(), &dispatch.Message{})

	assert.NoError(t, err)
	assert.Equal(t, 1, len(defaultDispatcher.DispatchCalls()))
}

func TestCallNonDefaultDispatcher(t *testing.T) {
	mre := &MockRuleEngine{
		ProcessMessageFunc: func(ctx context.Context, msg *dispatch.Message) ([]string, error) {
			return []string{"non-default"}, nil
		},
	}

	defaultDispatcher := &MockDispatcher{
		DispatchFunc: func(ctx context.Context, msg *dispatch.Message) error {
			return nil
		},
		NameFunc: func() string {
			return "default"
		},
	}

	nonDefaultDispatcher := &MockDispatcher{
		DispatchFunc: func(ctx context.Context, msg *dispatch.Message) error {
			return nil
		},
		NameFunc: func() string {
			return "non-default"
		},
	}

	sut := setupMessageService(t, mre)
	sut.RegisterDispatcher(defaultDispatcher, true)
	sut.RegisterDispatcher(nonDefaultDispatcher, false)

	err := sut.QueueMessage(context.Background(), &dispatch.Message{})

	assert.NoError(t, err)
	assert.Equal(t, 0, len(defaultDispatcher.DispatchCalls()))
	assert.Equal(t, 1, len(nonDefaultDispatcher.DispatchCalls()))
}

func TestCallMultipleNonDefaultDispatchers(t *testing.T) {
	mre := &MockRuleEngine{
		ProcessMessageFunc: func(ctx context.Context, msg *dispatch.Message) ([]string, error) {
			return []string{"non-default-1", "non-default-2"}, nil
		},
	}

	dispatchers := make(map[string]*MockDispatcher)
	dispatchers["non-default-1"] = &MockDispatcher{
		DispatchFunc: func(ctx context.Context, msg *dispatch.Message) error {
			return nil
		},
		NameFunc: func() string {
			return "non-default-1"
		},
	}
	dispatchers["non-default-2"] = &MockDispatcher{
		DispatchFunc: func(ctx context.Context, msg *dispatch.Message) error {
			return nil
		},
		NameFunc: func() string {
			return "non-default-2"
		},
	}

	sut := setupMessageService(t, mre)
	for _, d := range dispatchers {
		sut.RegisterDispatcher(d, false)
	}

	err := sut.QueueMessage(context.Background(), &dispatch.Message{})

	assert.NoError(t, err)
	for _, d := range dispatchers {
		assert.Equal(t, 1, len(d.DispatchCalls()))
	}
}

func TestNoDispatchersFound(t *testing.T) {
	mre := &MockRuleEngine{
		ProcessMessageFunc: func(ctx context.Context, msg *dispatch.Message) ([]string, error) {
			return []string{}, nil
		},
	}

	sut := setupMessageService(t, mre)

	err := sut.QueueMessage(context.Background(), &dispatch.Message{})
	assert.NoError(t, err)
}
