package logger

import (
	"context"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.Logger

const RequestIDKey = "request_id"

// Init initializes the logger
func Init(env string) error {
	var config zap.Config

	if env == "dev" || env == "development" {
		// Development mode: pretty console output
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	} else {
		// Production mode: JSON output
		config = zap.NewProductionConfig()
		config.EncoderConfig.TimeKey = "timestamp"
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}

	logger, err := config.Build()
	if err != nil {
		return err
	}

	Log = logger
	return nil
}

// InitDefault initializes logger with default development config
func InitDefault() {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	logger, _ := config.Build()
	Log = logger
}

// Sync flushes any buffered log entries
func Sync() {
	if Log != nil {
		_ = Log.Sync()
	}
}

// Info logs an info message
func Info(msg string, fields ...zap.Field) {
	if Log != nil {
		Log.Info(msg, fields...)
	}
}

// Error logs an error message
func Error(msg string, fields ...zap.Field) {
	if Log != nil {
		Log.Error(msg, fields...)
	}
}

// Debug logs a debug message
func Debug(msg string, fields ...zap.Field) {
	if Log != nil {
		Log.Debug(msg, fields...)
	}
}

// Warn logs a warning message
func Warn(msg string, fields ...zap.Field) {
	if Log != nil {
		Log.Warn(msg, fields...)
	}
}

// Fatal logs a fatal message and exits
func Fatal(msg string, fields ...zap.Field) {
	if Log != nil {
		Log.Fatal(msg, fields...)
	} else {
		os.Exit(1)
	}
}

// WithContext creates a logger with request_id from context
// Usage: logger.WithContext(ctx).Info("message", zap.String("key", "value"))
func WithContext(ctx context.Context) *zap.Logger {
	if Log == nil {
		return zap.NewNop()
	}

	// Extract request_id from context
	if requestID := ctx.Value(RequestIDKey); requestID != nil {
		if id, ok := requestID.(string); ok {
			return Log.With(zap.String("request_id", id))
		}
	}

	return Log
}

// GetRequestID extracts request_id from context
func GetRequestID(ctx context.Context) string {
	if requestID := ctx.Value(RequestIDKey); requestID != nil {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return "unknown"
}
