package log

import (
	"context"
	"sync"
)

var (
	once   sync.Once
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

func init() {
	// Choose which logger to use.

	once.Do(func() {
		zeroLog := newZeroLog()
		Logger = zeroLog

		Logger.Info(context.Background(), "Logger: ZeroLog initialized", nil)
	})
}
