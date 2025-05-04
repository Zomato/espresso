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
	logger zerolog.Logger
}

func NewZeroLogger() ZeroLog {
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
	})

	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	return ZeroLog{
		logger: log.Logger,
	}
}

func addFields(event *zerolog.Event, fields map[string]any) *zerolog.Event {
	for k, v := range fields {
		if k == "time" {
			event = event.Interface("timestamp", v)
		} else {
			event = event.Interface(k, v)
		}
	}

	return event
}

func (l ZeroLog) Info(ctx context.Context, msg string, fields customLogger.Fields) {
	addFields(l.logger.Info(), fields).Msg(msg)
}

func (l ZeroLog) Warn(ctx context.Context, msg string, fields customLogger.Fields) {
	addFields(l.logger.Warn(), fields).Msg(msg)
}

func (l ZeroLog) Error(ctx context.Context, msg string, err error, fields customLogger.Fields) {
	addFields(l.logger.Err(err), fields).Msg(msg)
}

func (l ZeroLog) Debug(ctx context.Context, msg string, fields customLogger.Fields) {
	addFields(l.logger.Debug(), fields).Msg(msg)
}
