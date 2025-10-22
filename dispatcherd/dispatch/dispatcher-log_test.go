package dispatch

import (
	"bytes"
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogDispatcherDispatch(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))
	slog.SetDefault(logger)
	dispatcher := NewLogDispatcher()

	msg := &Message{
		Title:   "Test Title",
		Message: "Test Message",
	}

	err := dispatcher.Dispatch(context.Background(), msg)
	assert.NoError(t, err)

	assert.Contains(t, buf.String(), `"messageTitle":"Test Title"`)
	assert.Contains(t, buf.String(), `"messageMessage":"Test Message"`)
}
