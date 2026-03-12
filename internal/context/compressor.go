package contextutil

import "strings"

func CompressPrompt(prompt string, limit int) string {
	if limit < 1 || len(prompt) <= limit {
		return prompt
	}
	return prompt[:limit] + "..."
}

func CompressTerms(values []string, limit int) []string {
	if limit < 1 || len(values) <= limit {
		return values
	}
	return values[:limit]
}

func NormalizeWhitespace(value string) string {
	return strings.Join(strings.Fields(value), " ")
}
