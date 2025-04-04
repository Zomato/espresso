package log

import (
	"context"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type ZeroLog struct {
	logger zerolog.Logger
}

func newZeroLog() ZeroLog {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	return ZeroLog{
		logger: log.Logger,
	}
}

func addFields(event *zerolog.Event, fields map[string]any) *zerolog.Event {
	for k, v := range fields {
		event.Interface(k, v)
	}

	return event
}

func (l ZeroLog) Info(ctx context.Context, msg string, fields Fields) {
	addFields(l.logger.Info(), fields).Msg(msg)
}

func (l ZeroLog) Warn(ctx context.Context, msg string, fields Fields) {
	addFields(l.logger.Warn(), fields).Msg(msg)
}

func (l ZeroLog) Error(ctx context.Context, msg string, err error, fields Fields) {
	addFields(l.logger.Err(err), fields).Msg(msg)
}

func (l ZeroLog) Debug(ctx context.Context, msg string, fields Fields) {
	addFields(l.logger.Debug(), fields).Msg(msg)
}
