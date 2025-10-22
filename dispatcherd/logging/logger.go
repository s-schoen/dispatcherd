package logging

import (
	"log/slog"
)

type LoggerComponent string

const (
	API               LoggerComponent = "api"
	Audit             LoggerComponent = "audit"
	MessageProcessing LoggerComponent = "processing"
	DataAccess        LoggerComponent = "dal"
)

func GetLogger(component LoggerComponent) *slog.Logger {
	return slog.Default().With("component", component)
}
