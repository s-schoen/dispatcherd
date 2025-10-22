package repository

import (
	"context"
	"dispatcherd/dispatch"
	"dispatcherd/logging"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
)

type RuleRepository interface {
	ListRules(ctx context.Context) ([]dispatch.Rule, error)
}

type FilesystemRuleRepository struct {
	logger        *slog.Logger
	ruleDirectory string
}

func NewFilesystemRuleRepository(ruleDirectory string) *FilesystemRuleRepository {
	return &FilesystemRuleRepository{
		logger:        logging.GetLogger(logging.DataAccess),
		ruleDirectory: ruleDirectory,
	}
}

func (r *FilesystemRuleRepository) ListRules(ctx context.Context) ([]dispatch.Rule, error) {
	var rules []dispatch.Rule

	r.logger.DebugContext(ctx, "loading rules from "+r.ruleDirectory)

	files, err := os.ReadDir(r.ruleDirectory)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".json" {
			filePath := filepath.Join(r.ruleDirectory, file.Name())
			fileContent, err := os.ReadFile(filePath)
			if err != nil {
				r.logger.ErrorContext(ctx, "failed to read rule file", logging.FieldError, err, "file", filePath)
				continue
			}

			var rule dispatch.Rule
			if err := json.Unmarshal(fileContent, &rule); err != nil {
				r.logger.ErrorContext(ctx, "failed to unmarshal rule file", logging.FieldError, err, "file", filePath)
				continue
			}
			rules = append(rules, rule)
		}
	}

	r.logger.InfoContext(ctx, "loaded "+strconv.Itoa(len(rules))+" rules")

	return rules, nil
}
