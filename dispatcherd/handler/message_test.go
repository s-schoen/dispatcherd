package handler_test

import (
	"bytes"
	"context"
	"dispatcherd/dispatch"
	"dispatcherd/handler"
	"dispatcherd/test"
	"errors"
	"log/slog"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func setupMessageHandler(t *testing.T, svc *MockMessageService) *handler.MessageHandler {
	t.Helper()
	var logBuffer bytes.Buffer
	mockLogger := slog.New(slog.NewJSONHandler(&logBuffer, nil))

	return handler.NewDispatchHandler(mockLogger, svc)
}

func TestPostMessageSuccess(t *testing.T) {
	mockSvc := &MockMessageService{
		QueueMessageFunc: func(ctx context.Context, msg *dispatch.Message) error {
			return nil
		},
	}
	h := setupMessageHandler(t, mockSvc)

	body := `{"title": "Test Title", "message": "Test Message"}`
	runner := test.NewTestRunner(h.HandlePost)
	runner.WithBodyString(body).Run(t).ExpectNoError().ExpectStatusCode(http.StatusOK)
	assert.Len(t, mockSvc.QueueMessageCalls(), 1)
}

func TestPostMessageInvalidBody(t *testing.T) {
	mockSvc := &MockMessageService{
		QueueMessageFunc: func(ctx context.Context, msg *dispatch.Message) error {
			return nil
		},
	}
	h := setupMessageHandler(t, mockSvc)

	body := `{"title": "Test Title"}`
	runner := test.NewTestRunner(h.HandlePost)
	runner.WithBodyString(body).Run(t).ExpectAPIError(http.StatusBadRequest)
	assert.Len(t, mockSvc.QueueMessageCalls(), 0)
}

func TestPostMessageInternalError(t *testing.T) {
	mockSvc := &MockMessageService{
		QueueMessageFunc: func(ctx context.Context, msg *dispatch.Message) error {
			return errors.New("test")
		},
	}
	h := setupMessageHandler(t, mockSvc)

	body := `{"title": "Test Title", "message": "Test Message"}`
	runner := test.NewTestRunner(h.HandlePost)
	runner.WithBodyString(body).Run(t).ExpectAPIError(http.StatusInternalServerError)
	assert.Len(t, mockSvc.QueueMessageCalls(), 1)
}

func TestPostMessageAPIError(t *testing.T) {
	mockSvc := &MockMessageService{
		QueueMessageFunc: func(ctx context.Context, msg *dispatch.Message) error {
			return handler.NotFound("message", "")
		},
	}
	h := setupMessageHandler(t, mockSvc)

	body := `{"title": "Test Title", "message": "Test Message"}`
	runner := test.NewTestRunner(h.HandlePost)
	runner.WithBodyString(body).Run(t).ExpectAPIError(http.StatusNotFound)
	assert.Len(t, mockSvc.QueueMessageCalls(), 1)
}
