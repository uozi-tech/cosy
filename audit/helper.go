package audit

import (
	"fmt"
	"strings"
)

// BuildFieldQuery builds query expression for a field, handling spaces by splitting into multiple AND terms
func BuildFieldQuery(value, field string) string {
	if strings.Contains(value, " ") {
		words := strings.Fields(value)
		validTerms := make([]string, 0, len(words))
		for _, word := range words {
			// Filter out empty words, words containing colons, and words with only special characters
			if word != "" && !strings.Contains(word, ":") && strings.TrimSpace(strings.Trim(word, "':\"")) != "" {
				// Escape special characters
				cleanWord := strings.ReplaceAll(word, "'", "\\'")
				cleanWord = strings.ReplaceAll(cleanWord, "\"", "\\\"")
				validTerms = append(validTerms, fmt.Sprintf("%s:%s*", field, cleanWord))
			}
		}
		if len(validTerms) == 0 {
			return ""
		}
		if len(validTerms) == 1 {
			return validTerms[0]
		}
		return "(" + strings.Join(validTerms, " and ") + ")"
	}

	// Return empty string if single word contains colon
	if strings.Contains(value, ":") {
		return ""
	}

	// Single word also needs to escape special characters
	cleanValue := strings.ReplaceAll(value, "'", "\\'")
	cleanValue = strings.ReplaceAll(cleanValue, "\"", "\\\"")
	return fmt.Sprintf("%s = %s", field, cleanValue)
}
