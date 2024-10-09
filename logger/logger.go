package logger

import (
    "github.com/gin-gonic/gin"
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
    "os"
    "time"
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

    // Join the outputs, encoders, and level-handling functions into
    // zapcore.Cores, then tee the two cores together.
    core := zapcore.NewTee(
        zapcore.NewCore(consoleEncoder, consoleErrors, highPriority),
        zapcore.NewCore(consoleEncoder, consoleDebugging, lowPriority),
    )

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
func Debug(args ...interface{}) {
    logger.Debugln(args...)
}

// Info logs a message at InfoLevel.
func Info(args ...interface{}) {
    logger.Infoln(args...)
}

// Warn logs a message at WarnLevel.
func Warn(args ...interface{}) {
    logger.Warnln(args...)
}

// Error logs a message at ErrorLevel.
func Error(args ...interface{}) {
    logger.Errorln(args...)
}

// DPanic logs a message at DPanicLevel.
func DPanic(args ...interface{}) {
    logger.DPanic(args...)
}

// Panic logs a message at PanicLevel.
func Panic(args ...interface{}) {
    logger.Panic(args...)
}

// Fatal logs a message at FatalLevel.
func Fatal(args ...interface{}) {
    logger.Fatalln(args...)
}

// Debugf logs a message at DebugLevel.
func Debugf(format string, args ...interface{}) {
    logger.Debugf(format, args...)
}

// Infof logs a message at InfoLevel.
func Infof(format string, args ...interface{}) {
    logger.Infof(format, args...)
}

// Warnf logs a message at WarnLevel.
func Warnf(format string, args ...interface{}) {
    logger.Warnf(format, args...)
}

// Errorf logs a message at ErrorLevel.
func Errorf(format string, args ...interface{}) {
    logger.Errorf(format, args...)
}

// DPanicf logs a message at DPanicLevel.
func DPanicf(format string, args ...interface{}) {
    logger.DPanicf(format, args...)
}

// Panicf logs a message at PanicLevel.
func Panicf(format string, args ...interface{}) {
    logger.Panicf(format, args...)
}

// Fatalf logs a message at FatalLevel.
func Fatalf(format string, args ...interface{}) {
    logger.Fatalf(format, args...)
}
