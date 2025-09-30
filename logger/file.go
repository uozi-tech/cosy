package logger

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap/zapcore"
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
