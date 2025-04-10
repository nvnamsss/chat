package logger

import (
	"context"
	"os"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var globalLogger *zap.Logger

// RequestIDKey is the context key for request ID
type ctxKey string

const (
	// RequestIDKey is the key for request ID in context
	RequestIDKey ctxKey = "request_id"
)

// Init initializes the logger
func Init(level string, env string) {
	config := zap.NewProductionConfig()

	// Set log level
	var logLevel zapcore.Level
	switch level {
	case "debug":
		logLevel = zap.DebugLevel
	case "info":
		logLevel = zap.InfoLevel
	case "warn":
		logLevel = zap.WarnLevel
	case "error":
		logLevel = zap.ErrorLevel
	default:
		logLevel = zap.InfoLevel
	}
	config.Level = zap.NewAtomicLevelAt(logLevel)

	// Configure output format
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// Add environment info
	config.InitialFields = map[string]interface{}{
		"service": "chat-service",
		"env":     env,
	}

	// Create logger
	var err error
	globalLogger, err = config.Build()
	if err != nil {
		// If we can't initialize the logger, use a simple fallback and exit
		zap.NewExample().Error("Failed to initialize logger", zap.Error(err))
		os.Exit(1)
	}

	zap.RedirectStdLog(globalLogger)
}

// WithRequestID adds a request ID to the logger
func WithRequestID(ctx context.Context) context.Context {
	if reqID, ok := ctx.Value(RequestIDKey).(string); ok && reqID != "" {
		return ctx
	}
	return context.WithValue(ctx, RequestIDKey, uuid.New().String())
}

// GetRequestID gets the request ID from context
func GetRequestID(ctx context.Context) string {
	if reqID, ok := ctx.Value(RequestIDKey).(string); ok {
		return reqID
	}
	return ""
}

// Field creates a zap field
func Field(key string, value interface{}) zap.Field {
	return zap.Any(key, value)
}

// Context returns a logger with context information
func Context(ctx context.Context) *zap.SugaredLogger {
	reqID := GetRequestID(ctx)
	if reqID == "" {
		return globalLogger.Sugar()
	}
	return globalLogger.With(zap.String("request_id", reqID)).Sugar()
}

// Debug logs a debug message
func Debug(msg string, fields ...zap.Field) {
	globalLogger.Debug(msg, fields...)
}

// Info logs an info message
func Info(msg string, fields ...zap.Field) {
	globalLogger.Info(msg, fields...)
}

// Warn logs a warning message
func Warn(msg string, fields ...zap.Field) {
	globalLogger.Warn(msg, fields...)
}

// Error logs an error message
func Error(msg string, fields ...zap.Field) {
	globalLogger.Error(msg, fields...)
}

// Fatal logs a fatal message and exits
func Fatal(msg string, fields ...zap.Field) {
	globalLogger.Fatal(msg, fields...)
}

// Sync flushes any buffered log entries
func Sync() error {
	if globalLogger != nil {
		return globalLogger.Sync()
	}
	return nil
}
