package logger

import (
	"context"
	"fmt"
	"testing"

	"github.com/uozi-tech/cosy/sls"
	"go.uber.org/zap/zapcore"
)

func TestEscapeSLSQueryValue(t *testing.T) {
	if got := escapeSLSQueryValue(`request-"quoted"\tail`); got != `request-\"quoted\"\\tail` {
		t.Fatalf("unexpected escaped query value: %q", got)
	}
}

func TestLogItemFromSLS(t *testing.T) {
	item := logItemFromSLS(map[string]string{
		"time":        "123.75",
		"level":       "warn",
		"caller":      "logger/session.go:1",
		FieldDBCaller: "service/user.go:42",
		"msg":         "slow query",
	})
	if item.Time != 123 || item.Level != zapcore.WarnLevel || item.Caller != "service/user.go:42" || item.Message != "slow query" {
		t.Fatalf("unexpected log item: %#v", item)
	}
}

func TestQueryCorrelationLogPagesContinuesPastOneThousand(t *testing.T) {
	offsets := make([]int64, 0, 12)
	logs, err := queryCorrelationLogPages(context.Background(), 100, 200, `correlation_id: "request-1"`, func(_ context.Context, request sls.GetLogsRequest) ([]map[string]string, error) {
		offsets = append(offsets, request.Offset)
		count := correlationLogPageSize
		if request.Offset == 1100 {
			count = 1
		}
		rows := make([]map[string]string, count)
		for i := range rows {
			rows[i] = map[string]string{
				"time": "123",
				"msg":  fmt.Sprintf("log-%d", int(request.Offset)+i),
			}
		}
		return rows, nil
	})
	if err != nil {
		t.Fatalf("query pages: %v", err)
	}
	if len(logs) != 1101 {
		t.Fatalf("expected 1101 logs, got %d", len(logs))
	}
	if len(offsets) != 12 || offsets[0] != 0 || offsets[len(offsets)-1] != 1100 {
		t.Fatalf("unexpected offsets: %#v", offsets)
	}
}
