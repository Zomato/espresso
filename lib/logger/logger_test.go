package log

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockLogger struct {
	infoCalls  []string
	warnCalls  []string
	errorCalls []string
	debugCalls []string
}

func (m *mockLogger) Info(ctx context.Context, msg string, fields Fields) {
	m.infoCalls = append(m.infoCalls, msg)
}

func (m *mockLogger) Warn(ctx context.Context, msg string, fields Fields) {
	m.warnCalls = append(m.warnCalls, msg)
}

func (m *mockLogger) Error(ctx context.Context, msg string, err error, fields Fields) {
	m.errorCalls = append(m.errorCalls, msg)
}

func (m *mockLogger) Debug(ctx context.Context, msg string, fields Fields) {
	m.debugCalls = append(m.debugCalls, msg)
}

func TestNoOpLogger(t *testing.T) {
	noOp := newNoOpLogger()
	ctx := context.Background()

	// Ensure calling NoOpLogger does not panic
	assert.NotPanics(t, func() {
		noOp.Info(ctx, "info", Fields{"k": "v"})
		noOp.Warn(ctx, "warn", Fields{"k": "v"})
		noOp.Error(ctx, "error", errors.New("err"), Fields{"k": "v"})
		noOp.Debug(ctx, "debug", Fields{"k": "v"})
	})
}

func TestInitialize(t *testing.T) {
	mock := &mockLogger{}
	Initialize(mock)
	t.Cleanup(func() {
		Initialize(newNoOpLogger())
	})

	ctx := context.Background()
	Logger.Info(ctx, "hello info", Fields{"foo": "bar"})
	Logger.Warn(ctx, "hello warn", Fields{"foo": "bar"})
	Logger.Error(ctx, "hello error", errors.New("sample error"), Fields{"foo": "bar"})
	Logger.Debug(ctx, "hello debug", Fields{"foo": "bar"})

	assert.Equal(t, []string{"hello info"}, mock.infoCalls)
	assert.Equal(t, []string{"hello warn"}, mock.warnCalls)
	assert.Equal(t, []string{"hello error"}, mock.errorCalls)
	assert.Equal(t, []string{"hello debug"}, mock.debugCalls)
}
