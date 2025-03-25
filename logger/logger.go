package logger

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/uozi-tech/cosy/settings"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var logger *zap.SugaredLogger

// Init initializes the logger with the given mode.
func Init(mode string) {
	// First, define our level-handling logic.
	highPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.ErrorLevel
	})
	lowPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		switch mode {
		case gin.ReleaseMode:
			return lvl >= zapcore.InfoLevel && lvl < zapcore.ErrorLevel
		default:
			fallthrough
		case gin.DebugMode:
			return lvl < zapcore.ErrorLevel
		}
	})

	// Directly output to stdout and stderr, and add caller information.
	consoleDebugging := zapcore.Lock(os.Stdout)
	consoleErrors := zapcore.Lock(os.Stderr)
	encodeCaller := zapcore.FullCallerEncoder
	if mode == gin.ReleaseMode {
		encodeCaller = zapcore.ShortCallerEncoder
	}
	encoderConfig := zapcore.EncoderConfig{
		// Keys can be anything except the empty string.
		TimeKey:        "T",
		LevelKey:       "L",
		NameKey:        "N",
		CallerKey:      "C",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "M",
		StacktraceKey:  "S",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   encodeCaller,
	}
	encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.DateTime)
	encoderConfig.ConsoleSeparator = "\t"
	encoderConfig.EncodeLevel = colorLevelEncoder
	consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)

	cores := []zapcore.Core{
		zapcore.NewCore(consoleEncoder, consoleErrors, highPriority),
		zapcore.NewCore(consoleEncoder, consoleDebugging, lowPriority),
	}

	if settings.LogSettings.EnableFileLog {
		if err := os.MkdirAll(settings.LogSettings.Dir, 0755); err != nil {
			log.Fatal(err)
		}

		errorLogWriter := &lumberjack.Logger{
			Filename:   filepath.Join(settings.LogSettings.Dir, "error.log"),
			MaxSize:    settings.LogSettings.MaxSize,
			MaxBackups: settings.LogSettings.MaxBackups,
			MaxAge:     settings.LogSettings.MaxAge,
			LocalTime:  true,
			Compress:   settings.LogSettings.Compress,
		}

		infoLogWriter := &lumberjack.Logger{
			Filename:   filepath.Join(settings.LogSettings.Dir, "info.log"),
			MaxSize:    settings.LogSettings.MaxSize,
			MaxBackups: settings.LogSettings.MaxBackups,
			MaxAge:     settings.LogSettings.MaxAge,
			LocalTime:  true,
			Compress:   settings.LogSettings.Compress,
		}

		fileEncoder := GetFileEncoder(mode)

		cores = append(cores,
			zapcore.NewCore(fileEncoder, zapcore.AddSync(errorLogWriter), highPriority),
			zapcore.NewCore(fileEncoder, zapcore.AddSync(infoLogWriter), lowPriority),
		)
	}

	// Join the outputs, encoders, and level-handling functions into
	// zapcore.Cores, then tee the two cores together.
	core := zapcore.NewTee(cores...)

	// From a zapcore.Core, it's easy to construct a Logger.
	logger = zap.New(core, zap.AddCaller()).WithOptions(zap.AddCallerSkip(1)).Sugar()
}

// Sync flushes any buffered log entries.
func Sync() {
	_ = logger.Sync()
}

// GetLogger returns the logger.
func GetLogger() *zap.SugaredLogger {
	return logger
}

// Debug logs a message at DebugLevel.
func Debug(args ...any) {
	logger.Debugln(args...)
}

// Info logs a message at InfoLevel.
func Info(args ...any) {
	logger.Infoln(args...)
}

// Warn logs a message at WarnLevel.
func Warn(args ...any) {
	logger.Warnln(args...)
}

// Error logs a message at ErrorLevel.
func Error(args ...any) {
	logger.Errorln(args...)
}

// DPanic logs a message at DPanicLevel.
func DPanic(args ...any) {
	logger.DPanic(args...)
}

// Panic logs a message at PanicLevel.
func Panic(args ...any) {
	logger.Panicln(args...)
}

// Fatal logs a message at FatalLevel.
func Fatal(args ...any) {
	logger.Fatalln(args...)
}

// Debugf logs a message at DebugLevel.
func Debugf(format string, args ...any) {
	logger.Debugf(format, args...)
}

// Infof logs a message at InfoLevel.
func Infof(format string, args ...any) {
	logger.Infof(format, args...)
}

// Warnf logs a message at WarnLevel.
func Warnf(format string, args ...any) {
	logger.Warnf(format, args...)
}

// Errorf logs a message at ErrorLevel.
func Errorf(format string, args ...any) {
	logger.Errorf(format, args...)
}

// DPanicf logs a message at DPanicLevel.
func DPanicf(format string, args ...any) {
	logger.DPanicf(format, args...)
}

// Panicf logs a message at PanicLevel.
func Panicf(format string, args ...any) {
	logger.Panicf(format, args...)
}

// Fatalf logs a message at FatalLevel.
func Fatalf(format string, args ...any) {
	logger.Fatalf(format, args...)
}
