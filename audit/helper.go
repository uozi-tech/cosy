package audit

import (
	"fmt"
	"strings"
)

// BuildFieldQuery 构建字段精确匹配表达式，支持多词 AND 链接
func BuildFieldQuery(value, field string) string {
	return buildFieldQuery(value, field, false)
}

// BuildFuzzyFieldQuery 构建字段模糊匹配表达式，使用后缀通配符
func BuildFuzzyFieldQuery(value, field string) string {
	return buildFieldQuery(value, field, true)
}

func buildFieldQuery(value, field string, useSuffix bool) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}

	words := strings.Fields(trimmed)
	if len(words) == 0 {
		return ""
	}

	formatted := make([]string, 0, len(words))
	for _, word := range words {
		escaped := escapeTerm(word)
		formatted = append(formatted, formatTerm(field, escaped, useSuffix || len(words) > 1))
	}

	switch len(formatted) {
	case 0:
		return ""
	case 1:
		return formatted[0]
	default:
		return "(" + strings.Join(formatted, " and ") + ")"
	}
}

func escapeTerm(term string) string {
	escaped := strings.ReplaceAll(term, "'", "\\'")
	return strings.ReplaceAll(escaped, "\"", "\\\"")
}

func formatTerm(field, value string, useSuffix bool) string {
	if useSuffix {
		return fmt.Sprintf("%s:\"%s\"", field, value)
	}
	return fmt.Sprintf("%s = %s", field, value)
}
