package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/uozi-tech/cosy/settings"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func GetFileEncoder(mode string) zapcore.Encoder {
	encoderCaller := zapcore.FullCallerEncoder
	if mode == gin.ReleaseMode {
		encoderCaller = zapcore.ShortCallerEncoder
	}

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   encoderCaller,
	}
	encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.DateTime + ".000")

	return zapcore.NewJSONEncoder(encoderConfig)
}

// NewFileCores builds zap cores with size-based log rotation.
// Error and above go to error.log; lower levels go to info.log.
func NewFileCores(mode string, highPriority, lowPriority zapcore.LevelEnabler) ([]zapcore.Core, error) {
	if err := os.MkdirAll(settings.LogSettings.Dir, 0755); err != nil {
		return nil, fmt.Errorf("create log directory: %w", err)
	}

	encoder := GetFileEncoder(mode)

	var (
		errorSyncer zapcore.WriteSyncer
		infoSyncer  zapcore.WriteSyncer
		err         error
	)

	if settings.LogSettings.EnableRotate {
		errorSyncer = newRotateWriteSyncer(filepath.Join(settings.LogSettings.Dir, "error.log"))
		infoSyncer = newRotateWriteSyncer(filepath.Join(settings.LogSettings.Dir, "info.log"))
	} else {
		errorSyncer, err = newPlainWriteSyncer(filepath.Join(settings.LogSettings.Dir, "error.log"))
		if err != nil {
			return nil, fmt.Errorf("open error log file: %w", err)
		}
		infoSyncer, err = newPlainWriteSyncer(filepath.Join(settings.LogSettings.Dir, "info.log"))
		if err != nil {
			return nil, fmt.Errorf("open info log file: %w", err)
		}
	}

	errorCore := zapcore.NewCore(
		encoder,
		errorSyncer,
		highPriority,
	)

	infoCore := zapcore.NewCore(
		encoder,
		infoSyncer,
		lowPriority,
	)

	return []zapcore.Core{errorCore, infoCore}, nil
}

func newRotateWriteSyncer(filename string) zapcore.WriteSyncer {
	return zapcore.AddSync(&lumberjack.Logger{
		Filename:   filename,
		MaxSize:    settings.LogSettings.MaxSize,
		MaxBackups: settings.LogSettings.MaxBackups,
		MaxAge:     settings.LogSettings.MaxAge,
		LocalTime:  true,
		Compress:   settings.LogSettings.Compress,
	})
}

func newPlainWriteSyncer(filename string) (zapcore.WriteSyncer, error) {
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, err
	}
	return zapcore.AddSync(f), nil
}
