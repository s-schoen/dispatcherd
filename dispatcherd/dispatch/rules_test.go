package dispatch_test

import (
	"bytes"
	"context"
	"dispatcherd/dispatch"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func setupEngine(t *testing.T) *dispatch.DefaultRuleEngine {
	t.Helper()

	var logBuffer bytes.Buffer
	mockLogger := slog.New(slog.NewJSONHandler(&logBuffer, nil))

	return dispatch.NewRuleEngine(mockLogger)
}

func createTestMessage(t *testing.T, tags map[string]string) *dispatch.Message {
	t.Helper()
	return dispatch.NewMessage("Test Title", "Test Message", tags)
}

func TestMessageWithNoTags(t *testing.T) {
	engine := setupEngine(t)
	msg := createTestMessage(t, nil)
	matched, err := engine.ProcessMessage(context.Background(), msg)
	assert.NoError(t, err)
	assert.Len(t, matched, 0)
}

func TestRuleEquals(t *testing.T) {
	engine := setupEngine(t)

	rules := []dispatch.Rule{
		{
			ID:             "Test tag1",
			DispatcherName: "test",
			Match: []dispatch.RuleMatch{
				{
					TagName:  "tag",
					Operator: dispatch.EQUALS,
					Value:    "value",
				},
			},
		},
	}
	engine.SetRules(rules)

	t.Run("Should match with with one message tag", func(t *testing.T) {
		msg := createTestMessage(t, map[string]string{
			"tag": "value",
		})
		matched, err := engine.ProcessMessage(context.Background(), msg)

		assert.NoError(t, err)
		assert.Len(t, matched, 1)
		assert.Equal(t, "test", matched[0])
	})

	t.Run("Should match with additional message tags", func(t *testing.T) {
		msg := createTestMessage(t, map[string]string{
			"tag":  "value",
			"tag2": "value",
		})
		matched, err := engine.ProcessMessage(context.Background(), msg)

		assert.NoError(t, err)
		assert.Len(t, matched, 1)
		assert.Equal(t, "test", matched[0])
	})

	t.Run("Should not match with not matching value", func(t *testing.T) {
		msg := createTestMessage(t, map[string]string{
			"tag": "not-value",
		})
		matched, err := engine.ProcessMessage(context.Background(), msg)

		assert.NoError(t, err)
		assert.Len(t, matched, 0)
	})
}

func TestRuleMultiTag(t *testing.T) {
	engine := setupEngine(t)

	rules := []dispatch.Rule{
		{
			ID:             "Test tag1",
			DispatcherName: "test",
			Match: []dispatch.RuleMatch{
				{
					TagName:  "tag1",
					Operator: dispatch.EQUALS,
					Value:    "value1",
				},
				{
					TagName:  "tag2",
					Operator: dispatch.EQUALS,
					Value:    "value2",
				},
			},
		},
	}
	engine.SetRules(rules)

	t.Run("Should match message with both tags", func(t *testing.T) {
		msg := createTestMessage(t, map[string]string{
			"tag1": "value1",
			"tag2": "value2",
		})

		matched, err := engine.ProcessMessage(context.Background(), msg)

		assert.NoError(t, err)
		assert.Len(t, matched, 1)
		assert.Equal(t, "test", matched[0])
	})

	t.Run("Should not match message with only one tag", func(t *testing.T) {
		msg := createTestMessage(t, map[string]string{
			"tag1": "value1",
		})

		matched, err := engine.ProcessMessage(context.Background(), msg)

		assert.NoError(t, err)
		assert.Len(t, matched, 0)
	})

	t.Run("Should match message with additional tags", func(t *testing.T) {
		msg := createTestMessage(t, map[string]string{
			"tag1": "value1",
			"tag2": "value2",
		})

		matched, err := engine.ProcessMessage(context.Background(), msg)

		assert.NoError(t, err)
		assert.Len(t, matched, 1)
		assert.Equal(t, "test", matched[0])
	})
}

func TestMultipleRules(t *testing.T) {
	engine := setupEngine(t)

	rules := []dispatch.Rule{
		{
			ID:             "Test tag1",
			DispatcherName: "test1",
			Match: []dispatch.RuleMatch{
				{
					TagName:  "tag",
					Operator: dispatch.EQUALS,
					Value:    "value",
				},
			},
		},
		{
			ID:             "Test tag2",
			DispatcherName: "test2",
			Match: []dispatch.RuleMatch{
				{
					TagName:  "tag2",
					Operator: dispatch.EQUALS,
					Value:    "value",
				},
			},
		},
	}
	engine.SetRules(rules)

	t.Run("Should match both rules", func(t *testing.T) {
		msg := createTestMessage(t, map[string]string{
			"tag":  "value",
			"tag2": "value",
		})
		matched, err := engine.ProcessMessage(context.Background(), msg)

		assert.NoError(t, err)
		assert.Len(t, matched, 2)
		assert.Equal(t, "test1", matched[0])
		assert.Equal(t, "test2", matched[1])
	})

	t.Run("Should match only one rule", func(t *testing.T) {
		msg := createTestMessage(t, map[string]string{
			"tag":  "not-value",
			"tag2": "value",
		})
		matched, err := engine.ProcessMessage(context.Background(), msg)

		assert.NoError(t, err)
		assert.Len(t, matched, 1)
		assert.Equal(t, "test2", matched[0])
	})
}
