package debug

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/uozi-tech/cosy/logger"
	"go.uber.org/zap/zapcore"
)

func TestRequestTraceWithPreferredLogsUsesCorrelationLogs(t *testing.T) {
	previousHasSLS := hasSLSSupport
	previousQuery := queryCorrelationLogs
	hasSLSSupport = func() bool { return true }
	queryCorrelationLogs = func(_ context.Context, correlationID string, from, to int64) ([]logger.LogItem, error) {
		if correlationID != "correlation-1" {
			t.Fatalf("unexpected correlation id: %q", correlationID)
		}
		if from != 938 || to != 1160 {
			t.Fatalf("unexpected query window: %d..%d", from, to)
		}
		return []logger.LogItem{{
			Time:    1001,
			Level:   zapcore.InfoLevel,
			Caller:  "service/example.go:42",
			Message: "from default log",
		}}, nil
	}
	t.Cleanup(func() {
		hasSLSSupport = previousHasSLS
		queryCorrelationLogs = previousQuery
	})

	trace := &RequestTrace{
		CorrelationID: "correlation-1",
		SessionLogs:   `[{"message":"fallback"}]`,
		StartTime:     1000,
		EndTime:       1100,
		Latency:       "2s",
	}
	result := requestTraceWithPreferredLogs(context.Background(), trace)

	if result == trace {
		t.Fatal("expected a copy so monitor history remains unchanged")
	}
	if trace.SessionLogs != `[{"message":"fallback"}]` {
		t.Fatalf("source trace was mutated: %q", trace.SessionLogs)
	}
	var logs []logger.LogItem
	if err := json.Unmarshal([]byte(result.SessionLogs), &logs); err != nil {
		t.Fatalf("decode preferred logs: %v", err)
	}
	if len(logs) != 1 || logs[0].Message != "from default log" {
		t.Fatalf("expected correlation logs, got %#v", logs)
	}
}

func TestRequestTraceWithPreferredLogsFallsBackToSessionLogs(t *testing.T) {
	previousHasSLS := hasSLSSupport
	previousQuery := queryCorrelationLogs
	hasSLSSupport = func() bool { return true }
	t.Cleanup(func() {
		hasSLSSupport = previousHasSLS
		queryCorrelationLogs = previousQuery
	})

	for _, test := range []struct {
		name string
		logs []logger.LogItem
		err  error
	}{
		{name: "empty result"},
		{name: "query error", err: errors.New("SLS unavailable")},
	} {
		t.Run(test.name, func(t *testing.T) {
			queryCorrelationLogs = func(context.Context, string, int64, int64) ([]logger.LogItem, error) {
				return test.logs, test.err
			}
			trace := &RequestTrace{
				CorrelationID: "correlation-1",
				SessionLogs:   `[{"message":"bounded fallback"}]`,
				StartTime:     1000,
				EndTime:       1100,
			}
			result := requestTraceWithPreferredLogs(context.Background(), trace)
			if result.SessionLogs != trace.SessionLogs {
				t.Fatalf("expected session log fallback, got %q", result.SessionLogs)
			}
		})
	}
}

func TestRequestTraceWithPreferredLogsSkipsSLSWhenUnavailable(t *testing.T) {
	previousHasSLS := hasSLSSupport
	previousQuery := queryCorrelationLogs
	hasSLSSupport = func() bool { return false }
	queryCorrelationLogs = func(context.Context, string, int64, int64) ([]logger.LogItem, error) {
		t.Fatal("SLS query must not run without an active producer")
		return nil, nil
	}
	t.Cleanup(func() {
		hasSLSSupport = previousHasSLS
		queryCorrelationLogs = previousQuery
	})

	trace := &RequestTrace{
		CorrelationID: "correlation-1",
		SessionLogs:   `[{"message":"bounded fallback"}]`,
	}
	result := requestTraceWithPreferredLogs(context.Background(), trace)
	if result.SessionLogs != trace.SessionLogs {
		t.Fatalf("expected session log fallback, got %q", result.SessionLogs)
	}
}
