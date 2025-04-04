/*
	Default no operation logger for no overhead. It is assigned default to ILogger instance in logger.go file.
*/

package log

import (
	"context"
)

type NoOpLogger struct{}

func newNoOpLogger() NoOpLogger {
	return NoOpLogger{}
}

func (n NoOpLogger) Info(ctx context.Context, msg string, fields Fields) {}

func (n NoOpLogger) Warn(ctx context.Context, msg string, fields Fields) {}

func (n NoOpLogger) Error(ctx context.Context, msg string, err error, fields Fields) {}

func (n NoOpLogger) Debug(ctx context.Context, msg string, fields Fields) {}
