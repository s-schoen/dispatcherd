package main

import (
	"context"
	"dispatcherd/dispatch"
	"dispatcherd/logging"
	"dispatcherd/repository"
	"dispatcherd/service"
	"fmt"
	"log/slog"
	"os"

	"github.com/caarlos0/env/v11"
	"github.com/lmittmann/tint"
)

const (
	EnvDev = "dev"
)

type AppConfig struct {
	ListenAddress             string     `env:"DISPATCHERD_LISTEN_ADDRESS"`
	LogLevel                  slog.Level `env:"DISPATCHERD_LOG_LEVEL"`
	Environment               string     `env:"DISPATCHERD_ENVIRONMENT"`
	CORSOrigin                string     `env:"DISPATCHERD_CORS_ALLOWED_ORIGIN"`
	RuleDirectory             string     `env:"DISPATCHERD_RULE_DIRECTORY"`
	DispatcherConfigDirectory string     `env:"DISPATCHERD_DISPATCHER_CONFIG_DIRECTORY"`
}

func main() {
	// load environment variables
	var appConfig = AppConfig{
		ListenAddress:             ":3001",
		LogLevel:                  slog.LevelDebug,
		Environment:               EnvDev,
		CORSOrigin:                "*",
		RuleDirectory:             "/data/rules",
		DispatcherConfigDirectory: "/data/dispatchers",
	}
	if err := env.Parse(&appConfig); err != nil {
		fmt.Println(err)
		panic("Error loading environment variables")
	}

	// setup logging
	w := os.Stdout
	var logger *slog.Logger
	if appConfig.Environment == EnvDev {
		// pretty log to console
		//nolint:exhaustruct // pkg defaults are fine
		loggerOptions := &tint.Options{
			Level: appConfig.LogLevel,
		}
		logger = slog.New(&logging.ContextHandler{Handler: tint.NewHandler(w, loggerOptions)})
	} else {
		// log json
		//nolint:exhaustruct // pkg defaults are fine
		loggerOptions := &slog.HandlerOptions{
			Level: appConfig.LogLevel,
		}
		logger = slog.New(slog.NewJSONHandler(w, loggerOptions))
	}

	// setup services
	ruleRepo := repository.NewFilesystemRuleRepository(logger, appConfig.RuleDirectory)
	dispatcherConfigRepo := repository.NewFileSystemDispatcherConfigRepository(logger, appConfig.DispatcherConfigDirectory)
	ruleEngine := dispatch.NewRuleEngine(logger)

	// load rules from fs
	rules, err := ruleRepo.ListRules(context.Background())
	if err != nil {
		logger.Error("failed to load rules", logging.LoggerFieldError, err)
		os.Exit(1)
	}
	ruleEngine.SetRules(rules)

	// load dispatcher configs from fs
	dispatcherConfigs, err := dispatcherConfigRepo.ListDispatcherConfigs(context.Background())
	if err != nil {
		logger.Error("failed to load dispatcher configs", logging.LoggerFieldError, err)
		os.Exit(1)
	}

	dispatcherService := service.NewDispatcherService(logger)
	for _, config := range dispatcherConfigs {
		if err := dispatcherService.LoadDispatcherConfig(config); err != nil {
			logger.Error("failed to load dispatcher config "+config.Name, logging.LoggerFieldError, err)
		}
	}

	messageService := service.NewMessageService(logger, ruleEngine, dispatcherService)

	// start api server
	serverOptions := ServerOptions{
		ListenAddress:  appConfig.ListenAddress,
		Logger:         logger,
		CorsOrigin:     appConfig.CORSOrigin,
		MessageService: messageService,
	}

	logger.Debug("allowed CORS origin: " + appConfig.CORSOrigin)

	server := NewServer(serverOptions)
	server.Start()
}
