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
	// Only escape backslashes and double quotes for phrase safety.
	// Do NOT alter ':' — colons inside quoted phrases are valid in SLS queries.
	escaped := strings.ReplaceAll(term, `\\`, `\\\\`)
	escaped = strings.ReplaceAll(escaped, `"`, `\\"`)
	return escaped
}

func formatTerm(field, value string) string {
	return fmt.Sprintf("%s:\"%s\"", field, value)
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
	return fmt.Sprintf("\"%s\"", escaped)
}
