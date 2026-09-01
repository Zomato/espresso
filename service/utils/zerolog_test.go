package utils

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	customLogger "github.com/Zomato/espresso/lib/logger"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected zerolog.Level
	}{
		{"trace", zerolog.TraceLevel},
		{"debug", zerolog.DebugLevel},
		{"info", zerolog.InfoLevel},
		{"warn", zerolog.WarnLevel},
		{"warning", zerolog.WarnLevel},
		{"error", zerolog.ErrorLevel},
		{"fatal", zerolog.FatalLevel},
		{"panic", zerolog.PanicLevel},
		{"disabled", zerolog.Disabled},
		{"off", zerolog.Disabled},
		{"none", zerolog.Disabled},
		{"unknown", zerolog.DebugLevel},
		{"", zerolog.DebugLevel},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, ParseLogLevel(tt.input))
		})
	}
}

func TestNewZeroLoggerWithConfig_Disabled(t *testing.T) {
	var buf bytes.Buffer
	logger := NewZeroLoggerWithConfig(LogConfig{
		Disabled: true,
		Output:   &buf,
	})

	ctx := context.Background()
	logger.Info(ctx, "test info message", customLogger.Fields{"key": "value"})
	logger.Error(ctx, "test error message", errors.New("err"), customLogger.Fields{"key": "value"})
	logger.Debug(ctx, "test debug message", nil)
	logger.Warn(ctx, "test warn message", nil)

	assert.Empty(t, buf.String(), "Output should be empty when logger is disabled")
}

func TestNewZeroLoggerWithConfig_LevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	logger := NewZeroLoggerWithConfig(LogConfig{
		Disabled: false,
		Level:    "warn",
		Format:   "json",
		Output:   &buf,
	})

	ctx := context.Background()
	logger.Debug(ctx, "debug msg should be filtered out", customLogger.Fields{"foo": "bar"})
	logger.Info(ctx, "info msg should be filtered out", customLogger.Fields{"foo": "bar"})

	assert.Empty(t, buf.String(), "Debug and Info messages should be filtered out at Warn level")

	logger.Warn(ctx, "warning message logged", customLogger.Fields{"level_test": "warn"})
	assert.Contains(t, buf.String(), "warning message logged")
	assert.Contains(t, buf.String(), `"level_test":"warn"`)

	buf.Reset()
	logger.Error(ctx, "error message logged", errors.New("sample error"), customLogger.Fields{"err_key": "err_val"})
	assert.Contains(t, buf.String(), "error message logged")
	assert.Contains(t, buf.String(), "sample error")
	assert.Contains(t, buf.String(), `"err_key":"err_val"`)
}

func TestNewZeroLoggerWithConfig_JSONFormat(t *testing.T) {
	var buf bytes.Buffer
	logger := NewZeroLoggerWithConfig(LogConfig{
		Disabled: false,
		Level:    "info",
		Format:   "json",
		Output:   &buf,
	})

	ctx := context.Background()
	logger.Info(ctx, "json formatted log", customLogger.Fields{"service": "espresso", "req_id": "12345"})

	output := buf.String()
	assert.True(t, strings.HasPrefix(strings.TrimSpace(output), "{"), "Expected JSON object output")
	assert.Contains(t, output, `"message":"json formatted log"`)
	assert.Contains(t, output, `"service":"espresso"`)
	assert.Contains(t, output, `"req_id":"12345"`)
}

func TestNewZeroLogger_ViperConfig(t *testing.T) {
	viper.Reset()
	viper.Set("logger.disabled", true)
	viper.Set("logger.level", "error")

	logger := NewZeroLogger()
	ctx := context.Background()
	logger.Info(ctx, "should not log", nil)

	viper.Reset()
	viper.Set("disableLogs", true)
	logger = NewZeroLogger()
	logger.Info(ctx, "should not log either", nil)
}
