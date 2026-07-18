package logger

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/uozi-tech/cosy/settings"
	"github.com/uozi-tech/cosy/sls"
	"go.uber.org/zap/zapcore"
)

const correlationLogPageSize = 100

type correlationLogPageFetcher func(context.Context, sls.GetLogsRequest) ([]map[string]string, error)

// QueryCorrelationLogs loads Default Log entries for one request or background
// task. It automatically walks all SLS result pages until exhausted or the
// caller cancels the request.
func QueryCorrelationLogs(ctx context.Context, correlationID string, from, to int64) ([]LogItem, error) {
	if !HasSLSSupport() {
		return nil, nil
	}
	if correlationID == "" {
		return nil, nil
	}
	client := sls.NewClient(settings.SLSSettings.EndPoint, settings.SLSSettings.GetCredentials())
	query := fmt.Sprintf(`%s: "%s"`, FieldCorrelationID, escapeSLSQueryValue(correlationID))
	return queryCorrelationLogPages(ctx, from, to, query, func(ctx context.Context, request sls.GetLogsRequest) ([]map[string]string, error) {
		return client.GetLogs(ctx, settings.SLSSettings.ProjectName, settings.SLSSettings.DefaultLogStoreName, request)
	})
}

func queryCorrelationLogPages(ctx context.Context, from, to int64, query string, fetchPage correlationLogPageFetcher) ([]LogItem, error) {
	items := make([]LogItem, 0, correlationLogPageSize)
	for offset := int64(0); ; offset += correlationLogPageSize {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		rows, err := fetchPage(ctx, sls.GetLogsRequest{
			From:    from,
			To:      to,
			Lines:   correlationLogPageSize,
			Offset:  offset,
			Reverse: false,
			Query:   query,
		})
		if err != nil {
			return nil, err
		}
		for _, row := range rows {
			items = append(items, logItemFromSLS(row))
		}
		if len(rows) < correlationLogPageSize {
			break
		}
	}
	return items, nil
}

func escapeSLSQueryValue(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	return strings.ReplaceAll(value, `"`, `\"`)
}

func logItemFromSLS(row map[string]string) LogItem {
	timestamp, _ := strconv.ParseFloat(row["time"], 64)
	level := zapcore.InfoLevel
	_ = level.Set(row["level"])
	caller := row[FieldDBCaller]
	if caller == "" {
		caller = row["caller"]
	}
	message := row["msg"]
	if message == "" {
		message = row["message"]
	}
	return LogItem{
		Time:    int64(timestamp),
		Level:   level,
		Caller:  caller,
		Message: message,
	}
}
