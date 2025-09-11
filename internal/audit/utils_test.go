package audit

import (
	"testing"
)

func TestExtractUserAgentFromHeaders(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "invalid json",
			input:    "invalid json",
			expected: "",
		},
		{
			name:     "valid headers with user-agent",
			input:    `{"User-Agent":["Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"],"Accept":["text/html"]}`,
			expected: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
		},
		{
			name:     "lowercase user-agent",
			input:    `{"user-agent":["curl/7.68.0"],"host":["localhost:8080"]}`,
			expected: "curl/7.68.0",
		},
		{
			name:     "mixed case user-agent",
			input:    `{"User-agent":["PostmanRuntime/7.29.0"],"cache-control":["no-cache"]}`,
			expected: "PostmanRuntime/7.29.0",
		},
		{
			name:     "uppercase user-agent",
			input:    `{"USER-AGENT":["MyBot/1.0"],"connection":["close"]}`,
			expected: "MyBot/1.0",
		},
		{
			name:     "no user-agent header",
			input:    `{"Accept":["application/json"],"Content-Type":["application/json"]}`,
			expected: "",
		},
		{
			name:     "empty user-agent array",
			input:    `{"User-Agent":[],"Accept":["text/html"]}`,
			expected: "",
		},
		{
			name:     "multiple user-agent values",
			input:    `{"User-Agent":["Mozilla/5.0","Safari/537.36"],"Accept":["text/html"]}`,
			expected: "Mozilla/5.0",
		},
		{
			name:     "nested json structure",
			input:    `{"User-Agent":["Chrome/91.0.4472.124"],"X-Custom":["value"]}`,
			expected: "Chrome/91.0.4472.124",
		},
		{
			name:     "special characters in user-agent",
			input:    `{"User-Agent":["Mozilla/5.0 (X11; Linux x86_64; rv:89.0) Gecko/20100101 Firefox/89.0"]}`,
			expected: "Mozilla/5.0 (X11; Linux x86_64; rv:89.0) Gecko/20100101 Firefox/89.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractUserAgentFromHeaders(tt.input)
			if result != tt.expected {
				t.Errorf("extractUserAgentFromHeaders(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func BenchmarkExtractUserAgentFromHeaders(b *testing.B) {
	headerJSON := `{"User-Agent":["Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"],"Accept":["text/html","application/xhtml+xml","application/xml;q=0.9,*/*;q=0.8"],"Accept-Language":["en-US,en;q=0.5"],"Accept-Encoding":["gzip, deflate"],"Connection":["keep-alive"],"Upgrade-Insecure-Requests":["1"]}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		extractUserAgentFromHeaders(headerJSON)
	}
}
