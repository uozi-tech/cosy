package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"time"

	"log"

	"github.com/aliyun/aliyun-log-go-sdk/producer"
	"github.com/gin-gonic/gin"
	"github.com/uozi-tech/cosy/settings"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/protobuf/proto"

	"sync"

	sls "github.com/aliyun/aliyun-log-go-sdk"
)

const Topic = "audit"

// Audit producer instance for API audit logging
var auditProducer *producer.Producer

// SLSWriter is a writer that sends logs to SLS
type SLSWriter struct {
	logStore string
	producer *producer.Producer
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

	// Parse the log entry
	var logEntry map[string]interface{}
	if err := json.Unmarshal(p, &logEntry); err != nil {
		return len(p), err
	}

	// Create SLS log
	log := &sls.Log{
		Time: proto.Uint32(uint32(time.Now().Unix())),
	}

	// Convert log entry to SLS contents
	for key, value := range logEntry {
		log.Contents = append(log.Contents, &sls.LogContent{
			Key:   proto.String(key),
			Value: proto.String(fmt.Sprintf("%v", value)),
		})
	}

	// Send to SLS
	err = w.producer.SendLog(
		settings.SLSSettings.ProjectName,
		w.logStore,
		"",
		settings.SLSSettings.Source,
		log,
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

// InitProducer initializes the SLS producer for this SLSWriter instance
func (w *SLSWriter) InitProducer() error {
	slsSettings := settings.SLSSettings
	if !slsSettings.Enable() {
		return fmt.Errorf("SLS settings not enabled")
	}

	// Initialize LogStores and indexes first
	if err := InitializeSLS(); err != nil {
		return fmt.Errorf("failed to initialize SLS LogStores and indexes: %w", err)
	}

	producerConfig := producer.GetDefaultProducerConfig()
	producerConfig.Endpoint = slsSettings.EndPoint
	provider := slsSettings.GetCredentialsProvider()
	producerConfig.CredentialsProvider = provider
	producerConfig.GeneratePackId = true
	producerConfig.LogTags = []*sls.LogTag{
		{
			Key:   proto.String("type"),
			Value: proto.String("System"),
		},
	}

	w.producer = producer.InitProducer(producerConfig)
	w.producer.Start()

	return nil
}

// InitAuditSLSProducer initializes the audit SLS producer for API audit logging
func InitAuditSLSProducer(ctx context.Context) error {
	slsSettings := settings.SLSSettings
	if !slsSettings.Enable() {
		return fmt.Errorf("SLS settings not enabled")
	}

	// Initialize LogStores and indexes first
	if err := InitializeSLS(); err != nil {
		// Use standard log to avoid circular dependency during initialization
		log.Printf("Failed to initialize SLS LogStores and indexes: %v\n", err)
		return err
	}

	producerConfig := producer.GetDefaultProducerConfig()
	// Note: Don't set custom logger here to avoid circular dependency
	// The producer will use its default logger
	producerConfig.Endpoint = slsSettings.EndPoint
	provider := slsSettings.GetCredentialsProvider()
	producerConfig.CredentialsProvider = provider
	// if you want to use log context, set the GeneratePackId to true
	producerConfig.GeneratePackId = true
	producerConfig.LogTags = []*sls.LogTag{
		{
			Key:   proto.String("type"),
			Value: proto.String(Topic),
		},
	}

	auditProducer = producer.InitProducer(producerConfig)
	auditProducer.Start()

	// Wait for context cancellation
	go func() {
		<-ctx.Done()
		if auditProducer != nil {
			auditProducer.SafeClose()
		}
	}()

	return nil
}

// GetAuditProducer returns the audit producer instance
func GetAuditProducer() *producer.Producer {
	return auditProducer
}

type SLSLogItem struct {
	Time    int64         `json:"time"`
	Level   zapcore.Level `json:"level"`
	Caller  string        `json:"caller"`
	Message string        `json:"message"`
}

type SLSLogStack struct {
	Items []SLSLogItem `json:"items"`
	mutex sync.Mutex
}

func NewSLSLogStack() *SLSLogStack {
	return &SLSLogStack{
		Items: make([]SLSLogItem, 0),
		mutex: sync.Mutex{},
	}
}

func (l *SLSLogStack) Append(item SLSLogItem) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.Items = append(l.Items, item)
}

func (l *SLSLogStack) AppendLog(level zapcore.Level, message string) {
	_, file, line, _ := runtime.Caller(2)
	l.Append(SLSLogItem{
		Time:    time.Now().Unix(),
		Level:   level,
		Caller:  fmt.Sprintf("%s:%d", file, line),
		Message: message,
	})
}

// ZapLogger is a hack logger for SLS
type ZapLogger struct {
	logger *zap.SugaredLogger
}

// Log logs the message to console with zap logger from sls
func (zl ZapLogger) Log(keyvals ...interface{}) error {
	if len(keyvals)%2 != 0 {
		return fmt.Errorf("odd number of arguments")
	}
	var loggerFunc func(args ...interface{})
	logger := zl.logger.WithOptions(zap.AddCallerSkip(2))
	for i := 0; i < len(keyvals); i += 2 {
		key, ok := keyvals[i].(string)
		if !ok {
			return fmt.Errorf("non-string key: %v", keyvals[i])
		}
		if key == "level" {
			switch keyvals[i+1] {
			case "debug":
				loggerFunc = logger.Debug
			case "warn":
				loggerFunc = logger.Warn
			case "error":
				loggerFunc = logger.Error
			case "info":
				loggerFunc = logger.Info
			default:
				loggerFunc = logger.Info
			}
		}
		if key == "msg" {
			loggerFunc(keyvals[i+1])
			return nil
		}
	}
	return nil
}
