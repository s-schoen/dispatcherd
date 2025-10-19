package dispatch_test

import (
	"context"
	dispatcherdContext "dispatcherd/context"
	"dispatcherd/dispatch"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMessage(t *testing.T) {
	const (
		expectedTitle   string = "Test Title"
		expectedMessage string = "Test Message"
	)
	expectedTags := map[string]string{
		"tag1": "value1",
		"tag2": "value2",
	}

	t.Run("NewMessage without expectedTags", func(t *testing.T) {
		msg := dispatch.NewMessage(expectedTitle, expectedMessage, nil)
		assert.Equal(t, expectedTitle, msg.Title)
		assert.Equal(t, expectedMessage, msg.Message)
		assert.Len(t, msg.Tags, 0)
		assert.NotEmpty(t, msg.ID)
	})

	t.Run("NewMessage with expectedTags", func(t *testing.T) {
		msg := dispatch.NewMessage(expectedTitle, expectedMessage, expectedTags)
		assert.Equal(t, expectedTitle, msg.Title)
		assert.Equal(t, expectedMessage, msg.Message)
		assert.Len(t, msg.Tags, 2)
		assert.NotEmpty(t, msg.ID)
	})
}

func TestAnnotateContext(t *testing.T) {
	ctx := context.Background()
	msg := dispatch.NewMessage("Test Title", "Test Message", nil)

	msgCtx := msg.AnnotateContext(ctx)
	assert.NotEmpty(t, msgCtx.Value(dispatcherdContext.KeyMessageID))
	assert.Equal(t, msg.ID, msgCtx.Value(dispatcherdContext.KeyMessageID))
}
