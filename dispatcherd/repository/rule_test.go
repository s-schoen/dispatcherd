package repository

import (
	"context"
	"dispatcherd/dispatch"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilesystemRuleRepositoryListRules(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "rules")
	assert.NoError(t, err)
	defer func(path string) {
		_ = os.RemoveAll(path)
	}(tempDir)

	rule1JSON := `{"id":"rule1","dispatcherName":"dispatcher1","match":[{"tagName":"tag1","operator":"eq","value":"value1"}]}`
	rule2JSON := `{"id":"rule2","dispatcherName":"dispatcher2","match":[{"tagName":"tag2","operator":"eq","value":"value2"}]}`

	err = os.WriteFile(filepath.Join(tempDir, "rule1.json"), []byte(rule1JSON), 0644)
	assert.NoError(t, err)

	err = os.WriteFile(filepath.Join(tempDir, "rule2.json"), []byte(rule2JSON), 0644)
	assert.NoError(t, err)

	// also create a non-json file to make sure it's ignored
	err = os.WriteFile(filepath.Join(tempDir, "ignore.txt"), []byte("ignore me"), 0644)
	assert.NoError(t, err)

	repo := NewFilesystemRuleRepository(tempDir)

	rules, err := repo.ListRules(context.Background())
	assert.NoError(t, err)

	expectedRules := []dispatch.Rule{
		{
			ID:             "rule1",
			DispatcherName: "dispatcher1",
			Match: []dispatch.RuleMatch{
				{
					TagName:  "tag1",
					Operator: dispatch.EQUALS,
					Value:    "value1",
				},
			},
		},
		{
			ID:             "rule2",
			DispatcherName: "dispatcher2",
			Match: []dispatch.RuleMatch{
				{
					TagName:  "tag2",
					Operator: dispatch.EQUALS,
					Value:    "value2",
				},
			},
		},
	}

	assert.ElementsMatch(t, expectedRules, rules)
}

func TestFilesystemRuleRepositoryListRulesFailed(t *testing.T) {
	t.Run("non-existent directory", func(t *testing.T) {
		repo := NewFilesystemRuleRepository("/non-existent-dir")

		_, err := repo.ListRules(context.Background())
		assert.Error(t, err)
	})

	t.Run("malformed json", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "rules")
		assert.NoError(t, err)
		defer func(path string) {
			_ = os.RemoveAll(path)
		}(tempDir)

		malformedJSON := `{"id":"rule1","dispatcherName":"dispatcher1","match":[`
		err = os.WriteFile(filepath.Join(tempDir, "rule1.json"), []byte(malformedJSON), 0644)
		assert.NoError(t, err)

		repo := NewFilesystemRuleRepository(tempDir)

		rules, err := repo.ListRules(context.Background())
		assert.NoError(t, err)
		assert.Len(t, rules, 0)
	})

	t.Run("unreadable file", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "rules")
		assert.NoError(t, err)
		defer func(path string) {
			_ = os.RemoveAll(path)
		}(tempDir)

		ruleFile := filepath.Join(tempDir, "rule1.json")
		err = os.WriteFile(ruleFile, []byte(`{}`), 0644)
		assert.NoError(t, err)

		err = os.Chmod(ruleFile, 0222)
		assert.NoError(t, err)

		repo := NewFilesystemRuleRepository(tempDir)

		rules, err := repo.ListRules(context.Background())
		assert.NoError(t, err)
		assert.Len(t, rules, 0)
	})
}
