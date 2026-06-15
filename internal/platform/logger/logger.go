package logger

import (
	"github.com/flyluman/scratch/internal/ports"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type zapAdapter struct {
	inner *zap.Logger
}

func (a *zapAdapter) Info(msg string, fields ...ports.Field) {
	a.inner.Info(msg, toZap(fields...)...)
}

func (a *zapAdapter) Warn(msg string, fields ...ports.Field) {
	a.inner.Warn(msg, toZap(fields...)...)
}

func (a *zapAdapter) Error(msg string, fields ...ports.Field) {
	a.inner.Error(msg, toZap(fields...)...)
}

func (a *zapAdapter) Fatal(msg string, fields ...ports.Field) {
	a.inner.Fatal(msg, toZap(fields...)...)
}

func (a *zapAdapter) Sync() error {
	return a.inner.Sync()
}

func toZap(fields ...ports.Field) []zap.Field {
	out := make([]zap.Field, len(fields))
	for i, f := range fields {
		out[i] = zap.Any(f.Key, f.Value)
	}
	return out
}

func New(serviceName, appEnv, logLevel string) ports.Logger {
	var level zapcore.Level
	_ = level.Set(logLevel)

	cfg := zap.NewProductionConfig()
	cfg.Level = zap.NewAtomicLevelAt(level)
	cfg.EncoderConfig.TimeKey = "timestamp"
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.InitialFields = map[string]any{
		"service": serviceName,
		"env":     appEnv,
	}

	inner, _ := cfg.Build()
	return &zapAdapter{inner: inner}
}
