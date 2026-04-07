package utils

import (
	"context"
	"os"
	"time"

	customLogger "github.com/Zomato/espresso/lib/logger"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type ZeroLog struct {
	logger  zerolog.Logger
	enabled bool
}

var (
	Logger ZeroLog
)

func NewZeroLogger() ZeroLog {
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
	})

	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	// Check if logging is disabled via environment variable
	enabled := os.Getenv("DISABLE_LOGGING") != "true"

	zeroLog := ZeroLog{
		logger:  log.Logger,
		enabled: enabled,
	}

	Logger = zeroLog

	return zeroLog
}

func addFields(event *zerolog.Event, fields map[string]any) *zerolog.Event {
	for k, v := range fields {
		event = event.Interface(k, v)
	}

	return event
}

func (l ZeroLog) Info(ctx context.Context, msg string, fields customLogger.Fields) {
	if l.enabled {
		addFields(l.logger.Info(), fields).Msg(msg)
	}
}

func (l ZeroLog) Warn(ctx context.Context, msg string, fields customLogger.Fields) {
	if l.enabled {
		addFields(l.logger.Warn(), fields).Msg(msg)
	}
}

func (l ZeroLog) Error(ctx context.Context, msg string, err error, fields customLogger.Fields) {
	if l.enabled {
		addFields(l.logger.Err(err), fields).Msg(msg)
	}
}

func (l ZeroLog) Debug(ctx context.Context, msg string, fields customLogger.Fields) {
	if l.enabled {
		addFields(l.logger.Debug(), fields).Msg(msg)
	}
}

func (l ZeroLog) Enabled() bool {
	return l.enabled
}
