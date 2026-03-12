package formatter

import "strings"

type MarkdownFormatter struct{}

func (f MarkdownFormatter) Format(content string) string {
	content = strings.ReplaceAll(content, "\r\n", "\n")
	content = strings.ReplaceAll(content, "\r", "\n")
	lines := strings.Split(content, "\n")
	formatted := make([]string, 0, len(lines))
	blankRun := 0

	for i, line := range lines {
		trimmedRight := strings.TrimRight(line, " \t")
		line = trimmedRight
		trimmed := strings.TrimSpace(line)

		if trimmed == "" {
			blankRun++
			if blankRun > 2 {
				continue
			}
			formatted = append(formatted, "")
			continue
		}
		blankRun = 0

		if strings.HasPrefix(trimmed, "#") {
			hashCount := 0
			for hashCount < len(trimmed) && trimmed[hashCount] == '#' {
				hashCount++
			}
			if hashCount > 0 && len(trimmed) > hashCount && trimmed[hashCount] != ' ' {
				trimmed = strings.Repeat("#", hashCount) + " " + strings.TrimSpace(trimmed[hashCount:])
			}
			if i > 0 && len(formatted) > 0 && formatted[len(formatted)-1] != "" {
				formatted = append(formatted, "")
			}
			formatted = append(formatted, trimmed)
			if i < len(lines)-1 {
				formatted = append(formatted, "")
			}
			continue
		}

		if strings.HasPrefix(trimmed, "-") || strings.HasPrefix(trimmed, "*") || strings.HasPrefix(trimmed, "+") {
			marker := trimmed[:1]
			rest := strings.TrimSpace(trimmed[1:])
			trimmed = marker + " " + rest
		}

		formatted = append(formatted, trimmed)
	}

	return strings.TrimSpace(strings.Join(formatted, "\n"))
}
