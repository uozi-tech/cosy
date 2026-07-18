package sls

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientGetLogs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/logstores/default/logs" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Fatalf("unexpected content type: %q", r.Header.Get("Content-Type"))
		}
		var request GetLogsRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if request.From != 100 || request.To != 200 || request.Lines != 100 || request.Offset != 20 || !request.Reverse || request.Query != `correlation_id: "request-1"` {
			t.Fatalf("unexpected request body: %#v", request)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"time":"123.5","level":"info","msg":"hello"}]}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, Credentials{})
	client.httpClient = server.Client()
	logs, err := client.GetLogs(context.Background(), "", "default", GetLogsRequest{
		From:    100,
		To:      200,
		Lines:   100,
		Offset:  20,
		Reverse: true,
		Query:   `correlation_id: "request-1"`,
	})
	if err != nil {
		t.Fatalf("GetLogs: %v", err)
	}
	if len(logs) != 1 || logs[0]["msg"] != "hello" {
		t.Fatalf("unexpected logs: %#v", logs)
	}
}
