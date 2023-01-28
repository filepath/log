package log

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"log"
	"os"
	"path"
)

const LoggerKey = iota

// InitWithOpts init logger with options
func InitWithOpts(opts *Config) {
	logger = &baseLogger{l: newZapLogger(opts)}
}

// newDefaultLogger init logger use default config
func newDefaultLogger() Logger {
	return New(&Config{
		LogDir:          "",
		LogFile:         "app.log",
		MaxAge:          0,
		MaxBackups:      0,
		MaxSize:         0,
		Compress:        false,
		LogLevel:        "debug",
		JsonEncode:      true,
		StacktraceLevel: "error",
		FilePerLevel:    false,
	})
}

// New create logger with options
func New(opts *Config) Logger {
	return &baseLogger{l: newZapLogger(opts)}
}

// newZapLogger new a zap logger
func newZapLogger(opts *Config) *zap.Logger {
	var cores []zapcore.Core
	if opts.FilePerLevel {
		// a log file per log level
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
		cores = append(cores, newZapCore(opts, "debug.log", debugPriority),
			newZapCore(opts, "info.log", infoPriority),
			newZapCore(opts, "warn.log", warnPriority),
			newZapCore(opts, "error.log", errorPriority))
	} else {
		// only one log file for all log level
		defaultLevel := zap.NewAtomicLevel()
		defaultLevel.SetLevel(logLevel(opts.LogLevel))
		cores = append(cores, newZapCore(opts, opts.LogFile, defaultLevel))
	}
	return zap.New(zapcore.NewTee(cores...), zap.AddCaller(), zap.AddCallerSkip(2),
		zap.Development(), zap.AddStacktrace(logLevel(opts.StacktraceLevel)))
}

func newWriteSyncer(opts *Config, file string) zapcore.WriteSyncer {
	hook := lumberjack.Logger{
		Filename:   path.Join(opts.LogDir, file),
		MaxSize:    opts.MaxSize,
		MaxBackups: opts.MaxBackups,
		MaxAge:     opts.MaxAge,
		Compress:   opts.Compress,
	}
	return zapcore.AddSync(&hook)
}

// newZapCore new zap core and hook for zap logger
func newZapCore(opts *Config, file string, level zapcore.LevelEnabler) zapcore.Core {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		FunctionKey:    "func",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
		EncodeName:     zapcore.FullNameEncoder,
	}

	encoder := zapcore.NewConsoleEncoder(encoderConfig)
	if opts.JsonEncode {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	writeSyncer := newWriteSyncer(opts, file)

	if opts.Stdout {
		writeSyncer = zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), writeSyncer)
	}
	return zapcore.NewCore(encoder, writeSyncer, level)
}

func (c *baseLogger) AddCallerSkip(callerSkip int) Logger {
	return newLoggerWithExtraSkip(c.l, callerSkip)
}

func (c *baseLogger) WithName(name string) Logger {
	l := c.l.Named(name)
	return newLoggerWithExtraSkip(l, -1)
}

func (c *baseLogger) WithValues(keysAndValues ...interface{}) Logger {
	l := c.l.With(handleFields(c.l, keysAndValues)...)
	return newLoggerWithExtraSkip(l, -1)
}

func (c *baseLogger) WithContext(ctx context.Context) Logger {
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

func (c *baseLogger) Debugf(format string, a ...interface{}) {
	c.l.Debug(fmt.Sprintf(format, a...))
}

func (c *baseLogger) Infof(format string, a ...interface{}) {
	c.l.Info(fmt.Sprintf(format, a...))
}

func (c *baseLogger) Warnf(format string, a ...interface{}) {
	c.l.Warn(fmt.Sprintf(format, a...))
}

func (c *baseLogger) Errorf(format string, a ...interface{}) {
	c.l.Error(fmt.Sprintf(format, a...))
}

func (c *baseLogger) Fatalf(format string, a ...interface{}) {
	c.l.Fatal(fmt.Sprintf(format, a...))
}

func (c *baseLogger) Debug(msg string, fields ...zap.Field) {
	c.l.Debug(msg, fields...)
}

func (c *baseLogger) Info(msg string, fields ...zap.Field) {
	c.l.Info(msg, fields...)
}

func (c *baseLogger) Warn(msg string, fields ...zap.Field) {
	c.l.Warn(msg, fields...)
}

func (c *baseLogger) Error(msg string, fields ...zap.Field) {
	c.l.Error(msg, fields...)
}

func (c *baseLogger) Fatal(msg string, fields ...zap.Field) {
	c.l.Fatal(msg, fields...)
}

// copy form http://github.com/go-logr/zapr/zapr.go
// handleFields converts a bunch of arbitrary key-value pairs into Zap fields.  It takes
// additional pre-converted Zap fields, for use with automatically attached fields, like
// `error`.
func handleFields(l *zap.Logger, args []interface{}, additional ...zap.Field) []zap.Field {
	if len(args) == 0 {
		return additional
	}

	fields := make([]zap.Field, 0, len(args)/2+len(additional))
	for i := 0; i < len(args); {
		if _, ok := args[i].(zap.Field); ok {
			break
		}
		if i == len(args)-1 {
			break
		}

		key, val := args[i], args[i+1]
		keyStr, isString := key.(string)
		if !isString {
			break
		}
		fields = append(fields, zap.Any(keyStr, val))
		i += 2
	}
	return append(fields, additional...)
}

// newLoggerWithExtraSkip allows creation of loggers with variable levels of callstack skipping
func newLoggerWithExtraSkip(l *zap.Logger, callerSkip int) Logger {
	_l := l.WithOptions(zap.AddCallerSkip(callerSkip))
	return &baseLogger{l: _l}
}

// logLevel log level string to zap logger level
func logLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "fatal":
		return zapcore.FatalLevel
	default:
		log.Fatalf("unknown log level %s", level)
	}
	return 0
}
