package log

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

const LoggerKey = "_logger"

// Zap init zap logger
func Zap(opts *Config) *zap.Logger {
	var cores []zapcore.Core
	if opts.FilePerLevel {
		// Each level of log output to the corresponding log file. eg debug.log info.log warn.log error.log
		debugPriority := zap.LevelEnablerFunc(func(lev zapcore.Level) bool {
			return lev == zap.DebugLevel
		})
		infoPriority := zap.LevelEnablerFunc(func(lev zapcore.Level) bool {
			return lev == zap.InfoLevel
		})
		warnPriority := zap.LevelEnablerFunc(func(lev zapcore.Level) bool {
			return lev == zap.WarnLevel
		})
		errorPriority := zap.LevelEnablerFunc(func(lev zapcore.Level) bool {
			return lev >= zap.ErrorLevel
		})
		cores = append(cores, NewZapCore(opts, "debug.log", debugPriority),
			NewZapCore(opts, "info.log", infoPriority),
			NewZapCore(opts, "warn.log", warnPriority),
			NewZapCore(opts, "error.log", errorPriority))
	} else {
		// only one log file for all log level
		defaultLevel := zap.NewAtomicLevelAt(logLevel(opts.LogLevel))
		cores = append(cores, NewZapCore(opts, opts.LogFile, defaultLevel))
	}
	if opts.StacktraceLevel == "" {
		opts.StacktraceLevel = "error"
	}
	return zap.New(zapcore.NewTee(cores...), zap.AddCaller(), zap.AddCallerSkip(2),
		zap.AddStacktrace(logLevel(opts.StacktraceLevel)))
}

// newWriteSyncer new file writer , fileName can empty, will use Config.LogFile
func newWriteSyncer(opts *Config, fileName string) zapcore.WriteSyncer {
	if fileName == "" {
		fileName = opts.LogFile
	}
	if opts.LogDir == "" {
		opts.LogDir = os.Getenv("LOG_DIR")
	}
	hook := lumberjack.Logger{
		Filename:   filepath.Join(opts.LogDir, fileName),
		MaxSize:    opts.MaxSize,
		MaxBackups: opts.MaxBackups,
		MaxAge:     opts.MaxAge,
		Compress:   opts.Compress,
		LocalTime:  true,
	}
	return zapcore.AddSync(&hook)
}

// NewZapCore new zap core and hook for zap logger
func NewZapCore(opts *Config, fileName string, level zapcore.LevelEnabler) zapcore.Core {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	// Do you need to use json format ?
	encoder := zapcore.NewConsoleEncoder(encoderConfig)
	if opts.JsonEncode {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}
	writeSyncer := newWriteSyncer(opts, fileName)

	if opts.Stdout {
		writeSyncer = zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), writeSyncer)
	}
	return zapcore.NewCore(encoder, writeSyncer, level)
}

func (b *baseLogger) AddCallerSkip(callerSkip int) Logger {
	return newLoggerWithExtraSkip(b.Logger, callerSkip)
}

func (b *baseLogger) WithName(name string) Logger {
	l := b.Logger.Named(name)
	return newLoggerWithExtraSkip(l, -1)
}

func (b *baseLogger) WithValues(keysAndValues ...interface{}) Logger {
	l := b.Logger.With(handleFields(keysAndValues)...)
	return newLoggerWithExtraSkip(l, -1)
}

// WithContext get logger from context, you can set some key value paris into logger with .WithValues method. such as tracId spanId ...
// example: for gin
// c.Request = c.Request.WithContext(context.WithValue(ctx, log.LoggerKey, log.WithContext(ctx).WithValues("requestId", requestId, "traceId", traceId)))
// when you need to print log , you can use log.WithContext(ctx).Warn("some message")
func (b *baseLogger) WithContext(ctx context.Context) Logger {
	if ctx == nil {
		return logger
	}
	l := ctx.Value(LoggerKey)
	ctxLogger, ok := l.(Logger)
	if ok {
		return ctxLogger
	}
	return logger
}

func (b *baseLogger) Debugf(format string, a ...interface{}) {
	b.Logger.Debug(fmt.Sprintf(format, a...))
}

func (b *baseLogger) Infof(format string, a ...interface{}) {
	b.Logger.Info(fmt.Sprintf(format, a...))
}

func (b *baseLogger) Warnf(format string, a ...interface{}) {
	b.Logger.Warn(fmt.Sprintf(format, a...))
}

func (b *baseLogger) Errorf(format string, a ...interface{}) {
	b.Logger.Error(fmt.Sprintf(format, a...))
}

func (b *baseLogger) Fatalf(format string, a ...interface{}) {
	b.Logger.Fatal(fmt.Sprintf(format, a...))
}

func (b *baseLogger) Debug(msg string, fields ...zap.Field) {
	b.Logger.Debug(msg, fields...)
}

func (b *baseLogger) Info(msg string, fields ...zap.Field) {
	b.Logger.Info(msg, fields...)
}

func (b *baseLogger) Warn(msg string, fields ...zap.Field) {
	b.Logger.Warn(msg, fields...)
}

func (b *baseLogger) Error(msg string, fields ...zap.Field) {
	b.Logger.Error(msg, fields...)
}

func (b *baseLogger) Fatal(msg string, fields ...zap.Field) {
	b.Logger.Fatal(msg, fields...)
}

// handleFields converts key value pairs to Zap fields
func handleFields(args []interface{}, additional ...zap.Field) []zap.Field {
	if len(args) == 0 {
		// fast-return if we have no suggared fields.
		return additional
	}

	fields := make([]zap.Field, 0, len(args)/2+len(additional))
	for i := 0; i < len(args); {
		// check just in case for strongly-typed Zap fields, which is illegal (since
		// it breaks implementation agnosticism), so we can give a better error message.
		if _, ok := args[i].(zap.Field); ok {
			break
		}
		// make sure this isn't a mismatched key
		if i == len(args)-1 {
			break
		}
		// process a key value pair, ensuring that the key is a string
		key, val := args[i], args[i+1]
		keyStr, isString := key.(string)
		if !isString {
			// if the key isn't a string
			break
		}
		fields = append(fields, zap.Any(keyStr, val))
		i += 2
	}
	return append(fields, additional...)
}

// newLoggerWithExtraSkip allows zap logger with callstack skipping
func newLoggerWithExtraSkip(l *zap.Logger, callerSkip int) Logger {
	return &baseLogger{l.WithOptions(zap.AddCallerSkip(callerSkip))}
}

// logLevel string logger level to zap logger level, default is debug level
func logLevel(level string) zapcore.Level {
	switch strings.ToLower(level) {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn", "warning":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "fatal":
		return zapcore.FatalLevel
	default:
		return zapcore.DebugLevel
	}
}
