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

func (l ZeroLog) Info(ctx context.Context, msg string) {
	l.logger.Info().Msg(msg)
}

func (l ZeroLog) Warn(ctx context.Context, msg string) {
	l.logger.Warn().Msg(msg)
}

func (l ZeroLog) Error(ctx context.Context, msg string, err error) {
	l.logger.Err(err).Msg(msg)
}

func (l ZeroLog) Debug(ctx context.Context, msg string) {
	l.logger.Debug().Msg(msg)
}
