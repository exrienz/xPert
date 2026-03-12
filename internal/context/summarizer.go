package contextutil

import "strings"

func SummarizeLines(lines []string) []string {
	if len(lines) <= 8 {
		return lines
	}
	return lines[:8]
}

func SummarizeText(text string, maxSentences int) string {
	if maxSentences < 1 {
		maxSentences = 2
	}
	parts := strings.Split(text, ".")
	filtered := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		filtered = append(filtered, part)
		if len(filtered) == maxSentences {
			break
		}
	}
	if len(filtered) == 0 {
		return ""
	}
	return strings.Join(filtered, ". ") + "."
}

func ExtractTerminology(values ...string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, 12)
	for _, value := range values {
		for _, part := range strings.Fields(strings.ToLower(value)) {
			part = strings.Trim(part, ".,:;!?()[]{}\"'")
			if len(part) < 6 {
				continue
			}
			if _, ok := seen[part]; ok {
				continue
			}
			seen[part] = struct{}{}
			out = append(out, part)
			if len(out) == 12 {
				return out
			}
		}
	}
	return out
}
