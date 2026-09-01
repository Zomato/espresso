package utils

import (
	"context"
	"io"
	"os"
	"strings"
	"time"

	customLogger "github.com/Zomato/espresso/lib/logger"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

type LogConfig struct {
	Disabled bool
	Level    string
	Format   string
	Output   io.Writer
}

type ZeroLog struct {
	logger zerolog.Logger
}

var (
	Logger ZeroLog
)

// ParseLogLevel parses a string log level into zerolog.Level
func ParseLogLevel(level string) zerolog.Level {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "trace":
		return zerolog.TraceLevel
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn", "warning":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	case "panic":
		return zerolog.PanicLevel
	case "disabled", "off", "none":
		return zerolog.Disabled
	default:
		return zerolog.DebugLevel
	}
}

// NewZeroLoggerWithConfig initializes ZeroLog with given configuration parameters
func NewZeroLoggerWithConfig(cfg LogConfig) ZeroLog {
	if cfg.Disabled {
		zerolog.SetGlobalLevel(zerolog.Disabled)
		zLog := zerolog.Nop()
		log.Logger = zLog
		zeroLog := ZeroLog{logger: zLog}
		Logger = zeroLog
		return zeroLog
	}

	out := cfg.Output
	if out == nil {
		out = os.Stderr
	}

	var zLog zerolog.Logger
	if strings.ToLower(cfg.Format) == "json" {
		zLog = zerolog.New(out).With().Timestamp().Logger()
	} else {
		zLog = zerolog.New(zerolog.ConsoleWriter{
			Out:        out,
			TimeFormat: time.RFC3339,
		}).With().Timestamp().Logger()
	}

	lvl := ParseLogLevel(cfg.Level)
	zLog = zLog.Level(lvl)
	zerolog.SetGlobalLevel(lvl)
	log.Logger = zLog

	zeroLog := ZeroLog{
		logger: zLog,
	}

	Logger = zeroLog
	return zeroLog
}

// NewZeroLogger initializes ZeroLog based on configuration from viper or defaults
func NewZeroLogger() ZeroLog {
	disabled := viper.GetBool("logger.disabled") || viper.GetBool("disableLogs") || viper.GetBool("disable_logs")
	level := viper.GetString("logger.level")
	if level == "" {
		level = viper.GetString("logLevel")
	}
	if level == "" {
		level = viper.GetString("log_level")
	}
	if level == "" {
		level = "debug"
	}

	format := viper.GetString("logger.format")
	if format == "" {
		format = viper.GetString("log_format")
	}

	return NewZeroLoggerWithConfig(LogConfig{
		Disabled: disabled,
		Level:    level,
		Format:   format,
		Output:   os.Stderr,
	})
}

func addFields(event *zerolog.Event, fields map[string]any) *zerolog.Event {
	for k, v := range fields {
		event = event.Interface(k, v)
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

