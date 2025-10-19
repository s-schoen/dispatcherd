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

type DispatcherConfigRepository interface {
	ListDispatcherConfigs(ctx context.Context) ([]dispatch.DispatcherConfig, error)
}

type FileSystemDispatcherConfigRepository struct {
	logger          *slog.Logger
	configDirectory string
}

func (f FileSystemDispatcherConfigRepository) ListDispatcherConfigs(ctx context.Context) ([]dispatch.DispatcherConfig, error) {
	var configs []dispatch.DispatcherConfig

	f.logger.DebugContext(ctx, "loading dispatcher configs from "+f.configDirectory)

	files, err := os.ReadDir(f.configDirectory)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".json" {
			filePath := filepath.Join(f.configDirectory, file.Name())
			fileContent, err := os.ReadFile(filePath)
			if err != nil {
				f.logger.ErrorContext(ctx, "failed to read dispatcher config file", logging.LoggerFieldError, err, "file", filePath)
				continue
			}

			var conf dispatch.DispatcherConfig
			if err := json.Unmarshal(fileContent, &conf); err != nil {
				f.logger.ErrorContext(ctx, "failed to unmarshal dispatcher config file", logging.LoggerFieldError, err, "file", filePath)
				continue
			}
			configs = append(configs, conf)
		}
	}

	f.logger.InfoContext(ctx, "loaded "+strconv.Itoa(len(configs))+" dispatcher configs")

	return configs, nil
}

func NewFileSystemDispatcherConfigRepository(logger *slog.Logger, configDirectory string) DispatcherConfigRepository {
	return FileSystemDispatcherConfigRepository{
		logger:          logger,
		configDirectory: configDirectory,
	}
}
