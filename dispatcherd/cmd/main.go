package main

import (
	"dispatcherd/dispatch"
	"dispatcherd/logging"
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
	ListenAddress string     `env:"DISPATCHERD_LISTEN_ADDRESS"`
	LogLevel      slog.Level `env:"DISPATCHERD_LOG_LEVEL"`
	Environment   string     `env:"DISPATCHERD_ENVIRONMENT"`
	CORSOrigin    string     `env:"DISPATCHERD_CORS_ALLOWED_ORIGIN"`
}

func main() {
	// load environment variables
	var appConfig = AppConfig{
		ListenAddress: ":3001",
		LogLevel:      slog.LevelDebug,
		Environment:   EnvDev,
		CORSOrigin:    "*",
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
	ruleEngine := dispatch.NewRuleEngine(logger)
	// set rules manually for now
	ruleEngine.SetRules([]dispatch.Rule{
		{
			ID:             "Test tag1",
			DispatcherName: "d1",
			Match: []dispatch.RuleMatch{
				{
					TagName:  "tag1",
					Operator: dispatch.EQUALS,
					Value:    "value1",
				},
			},
		},
	})

	messageService := service.NewMessageService(logger, ruleEngine)
	messageService.RegisterDispatcher(dispatch.NewLogDispatcher(logger), true)

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
