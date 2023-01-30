package log

import (
	"context"
	"os"

	"go.uber.org/zap"
)

type Logger interface {
	// AddCallerSkip new logger with callstack skipping.
	AddCallerSkip(callerSkip int) Logger

	// WithName adds some key-value pairs of context to a logger.
	WithName(name string) Logger

	WithContext(ctx context.Context) Logger

	// WithValues adds a new element to the logger's name.
	WithValues(keysAndValues ...interface{}) Logger

	Debugf(format string, a ...interface{})

	Infof(format string, a ...interface{})

	Warnf(format string, a ...interface{})

	Errorf(format string, a ...interface{})

	Fatalf(format string, a ...interface{})

	Debug(msg string, fields ...zap.Field)

	Info(msg string, fields ...zap.Field)

	Warn(msg string, fields ...zap.Field)

	Error(msg string, fields ...zap.Field)

	Fatal(msg string, fields ...zap.Field)
}

type baseLogger struct {
	*zap.Logger
}

var logger Logger

// New create logger with options and init global logger
func New(opts *Config) Logger {
	l := Zap(opts)
	logger = &baseLogger{l}
	// replaces the zap global Logger and SugaredLogger
	zap.ReplaceGlobals(l)
	return logger
}

// defaultLogger new default Logger if logger is nil
func defaultLogger() Logger {
	if logger == nil {
		logger = newDefaultLogger()
	}
	return logger
}

// newDefaultLogger if not new a logger, will generate a default logger
func newDefaultLogger() Logger {
	return New(&Config{
		LogDir:       os.Getenv("LOG_DIR"),
		JsonEncode:   true,
		FilePerLevel: true,
	})
}

func Debugf(format string, a ...interface{}) {
	defaultLogger().Debugf(format, a...)
}

func Infof(format string, a ...interface{}) {
	defaultLogger().Infof(format, a...)
}

func Warnf(format string, a ...interface{}) {
	defaultLogger().Warnf(format, a...)
}

func Errorf(format string, a ...interface{}) {
	defaultLogger().Errorf(format, a...)
}

func Fatalf(format string, a ...interface{}) {
	defaultLogger().Fatalf(format, a...)
}

func Debug(msg string, fields ...zap.Field) {
	defaultLogger().Debug(msg, fields...)
}

func Info(msg string, fields ...zap.Field) {
	defaultLogger().Info(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	defaultLogger().Warn(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	defaultLogger().Error(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	defaultLogger().Fatal(msg, fields...)
}

func WithName(name string) Logger {
	return defaultLogger().WithName(name).AddCallerSkip(-1)
}

func WithValues(keysAndValues ...interface{}) Logger {
	return defaultLogger().WithValues(keysAndValues).AddCallerSkip(-1)
}

func WithContext(ctx context.Context) Logger {
	return defaultLogger().WithContext(ctx)
}
