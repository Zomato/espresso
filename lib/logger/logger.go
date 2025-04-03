package log

import (
	"context"
	"sync"
)

var (
	once   sync.Once
	logger ILogger
)

type Level uint8

const (
	DebugLevel Level = iota
	Info
	Warn
	Error
)

type ILogger interface {
	Info(ctx context.Context, msg string)
	Warn(ctx context.Context, msg string)
	Error(ctx context.Context, msg string, err error)
	Debug(ctx context.Context, msg string)
}

func InitLogger() {
	// Choose which logger to use.

	once.Do(func() {
		zeroLog := newZeroLog()
		logger = zeroLog

		logger.Info(context.Background(), "Logger: ZeroLog initialized")
	})
}
