package audit

import (
	"fmt"
	"strings"
)

func BuildFieldQuery(value, field string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}

	escaped := escapeTerm(trimmed)

	return formatTerm(field, escaped)
}

func escapeTerm(term string) string {
	// Escape all backslashes and all double quotes.
	// This ensures literal backslashes are preserved (\\) and quotes are safe (\").
	escaped := strings.ReplaceAll(term, "\\", "\\\\")
	escaped = strings.ReplaceAll(escaped, "\"", "\\\"")
	return escaped
}

func formatTerm(field, value string) string {
	// Use SLS phrase search for fielded queries to avoid tokenizer surprises on spaces/non-ASCII
	return fmt.Sprintf("%s:#\"%s\"", field, value)
}

// BuildFullTextQuery builds a plain full-text phrase query across all fields
// by returning an escaped quoted phrase, which can be concatenated with other
// field conditions using 'and'. Example output: "\"some phrase: with colon\""
func BuildFullTextQuery(phrase string) string {
	trimmed := strings.TrimSpace(phrase)
	if trimmed == "" {
		return ""
	}
	escaped := escapeTerm(trimmed)
	// Use global phrase search
	return fmt.Sprintf("#\"%s\"", escaped)
}
