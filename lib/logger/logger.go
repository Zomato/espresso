package log

import (
	"context"
)

var (

	// Assigning default no-op logger. Later it will be replaced by actual logger(if provided) inside init function.
	Logger ILogger = newNoOpLogger()
)

type Level uint8

const (
	DebugLevel Level = iota
	Info
	Warn
	Error
)

type Fields map[string]any

type ILogger interface {
	Info(ctx context.Context, msg string, fields Fields)
	Warn(ctx context.Context, msg string, fields Fields)
	Error(ctx context.Context, msg string, err error, fields Fields)
	Debug(ctx context.Context, msg string, fields Fields)
}

func Initialize(loggerInstance ILogger) {
	Logger = loggerInstance
}
