package sls

import "encoding/json"

// Credentials holds SLS access credentials.
type Credentials struct {
	AccessKeyID     string
	AccessKeySecret string
}

// Error represents an SLS API error response.
type Error struct {
	HTTPCode  int32  `json:"httpCode"`
	Code      string `json:"errorCode"`
	Message   string `json:"errorMessage"`
	RequestID string `json:"requestID"`
}

func (e *Error) Error() string {
	b, _ := json.Marshal(e)
	return string(b)
}

// LogContent is a key-value pair within a log entry.
type LogContent struct {
	Key   string
	Value string
}

// Log is a single log entry with a unix timestamp and contents.
type Log struct {
	Time     uint32
	TimeNs   uint32
	Contents []*LogContent
}

// LogTag is a key-value tag attached to a LogGroup.
type LogTag struct {
	Key   string
	Value string
}

// LogGroup is a batch of logs sharing the same topic, source and tags.
type LogGroup struct {
	Logs    []*Log
	Topic   string
	Source  string
	LogTags []*LogTag
}

// GenerateLog creates a Log from a unix timestamp and key-value contents.
func GenerateLog(timestamp uint32, contents map[string]string) *Log {
	l := &Log{Time: timestamp}
	for k, v := range contents {
		l.Contents = append(l.Contents, &LogContent{Key: k, Value: v})
	}
	return l
}

// Index types for LogStore index configuration (JSON-serializable).

type IndexKey struct {
	Token         []string `json:"token"`
	CaseSensitive bool     `json:"caseSensitive"`
	Type          string   `json:"type"`
	DocValue      bool     `json:"doc_value,omitempty"`
	Alias         string   `json:"alias,omitempty"`
	Chn           bool     `json:"chn"`
}

type IndexLine struct {
	Token         []string `json:"token"`
	CaseSensitive bool     `json:"caseSensitive"`
	IncludeKeys   []string `json:"include_keys,omitempty"`
	ExcludeKeys   []string `json:"exclude_keys,omitempty"`
	Chn           bool     `json:"chn"`
}

type Index struct {
	Keys map[string]IndexKey `json:"keys,omitempty"`
	Line *IndexLine          `json:"line,omitempty"`
}

// LogFunc is a function that logs a message with level ("warn" or "error") and text.
type LogFunc func(level, msg string, keyvals ...any)
