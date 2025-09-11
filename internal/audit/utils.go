package audit

import (
	"encoding/json"
	"strings"
)

// extractUserAgentFromHeaders extracts User-Agent from request headers JSON string
func extractUserAgentFromHeaders(headerJSON string) string {
	if headerJSON == "" {
		return ""
	}

	// Try to parse JSON format request headers
	var headers map[string][]string
	if err := json.Unmarshal([]byte(headerJSON), &headers); err != nil {
		return ""
	}

	// Find User-Agent, considering different case forms
	for key, values := range headers {
		if strings.EqualFold(key, "user-agent") && len(values) > 0 {
			return values[0]
		}
	}

	return ""
}
