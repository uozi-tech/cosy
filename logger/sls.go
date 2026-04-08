package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/uozi-tech/cosy/settings"
	"github.com/uozi-tech/cosy/sls"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const Topic = "audit"

// Audit producer instance for API audit logging
var auditProducer *sls.Producer

// SLSWriter is a writer that sends logs to SLS
type SLSWriter struct {
	logStore string
	producer *sls.Producer
}

// NewSLSWriter creates a new SLS writer
func NewSLSWriter(logStore string) *SLSWriter {
	return &SLSWriter{
		logStore: logStore,
	}
}

// Write implements io.Writer interface
func (w *SLSWriter) Write(p []byte) (n int, err error) {
	if !settings.SLSSettings.Enable() {
		return len(p), nil
	}

	if w.producer == nil {
		return len(p), fmt.Errorf("SLS producer not initialized")
	}

	var logEntry map[string]any
	if err := json.Unmarshal(p, &logEntry); err != nil {
		return len(p), err
	}

	now := time.Now()
	l := &sls.Log{
		Time:   uint32(now.Unix()),
		TimeNs: uint32(now.Nanosecond()),
	}
	for key, value := range logEntry {
		l.Contents = append(l.Contents, &sls.LogContent{
			Key:   key,
			Value: fmt.Sprintf("%v", value),
		})
	}

	err = w.producer.SendLog(
		settings.SLSSettings.ProjectName,
		w.logStore,
		"",
		settings.SLSSettings.Source,
		l,
	)

	return len(p), err
}

// GetSLSEncoder returns a JSON encoder for SLS
func GetSLSEncoder(mode string) zapcore.Encoder {
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
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.EpochTimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   encoderCaller,
	}

	return zapcore.NewJSONEncoder(encoderConfig)
}

func newSLSProducer(tagValue string) (*sls.Producer, error) {
	s := settings.SLSSettings
	if !s.Enable() {
		return nil, fmt.Errorf("SLS settings not enabled")
	}
	if err := InitializeSLS(); err != nil {
		return nil, fmt.Errorf("failed to initialize SLS LogStores and indexes: %w", err)
	}
	p, err := sls.NewProducer(sls.ProducerConfig{
		Endpoint:    s.EndPoint,
		Credentials: s.GetCredentials(),
		LogTags:     []*sls.LogTag{{Key: "type", Value: tagValue}},
		LogFunc:     slsLogFunc(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create SLS producer: %w", err)
	}
	p.Start()
	return p, nil
}

// InitProducer initializes the SLS producer for this SLSWriter instance
func (w *SLSWriter) InitProducer() error {
	p, err := newSLSProducer("System")
	if err != nil {
		return err
	}
	w.producer = p
	return nil
}

// InitAuditSLSProducer initializes the audit SLS producer for API audit logging
func InitAuditSLSProducer(ctx context.Context) error {
	p, err := newSLSProducer(Topic)
	if err != nil {
		return err
	}
	auditProducer = p

	go func() {
		<-ctx.Done()
		if auditProducer != nil {
			auditProducer.SafeClose()
		}
	}()
	return nil
}

// GetAuditProducer returns the audit producer instance
func GetAuditProducer() *sls.Producer {
	return auditProducer
}

func slsLogFunc() sls.LogFunc {
	return func(level, msg string, keyvals ...any) {
		l := GetLogger()
		if l == nil {
			return
		}
		s := l.WithOptions(zap.AddCallerSkip(2))
		switch level {
		case "warn", "warning":
			s.Warnw(msg, keyvals...)
		case "error":
			s.Errorw(msg, keyvals...)
		}
	}
}
