package logger

import (
	"context"

	"go.uber.org/zap"
)

type ContextKey string

const Logger ContextKey = "logger"

// ContextWithLogger returns a copy of the provided context.Context that associates a key with the provided *zap.Logger.
func ContextWithLogger(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, Logger, logger)
}

// LoggerFromContext returns the *zap.Logger associated with a key in the provided context.Context.
func LoggerFromContext(ctx context.Context) *zap.Logger {
	return ctx.Value(Logger).(*zap.Logger)
}
