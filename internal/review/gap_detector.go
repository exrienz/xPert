package review

import (
	"fmt"
	"sort"
	"strings"
)

type GapDetector struct{}

func NewGapDetector() *GapDetector { return &GapDetector{} }

func (g *GapDetector) Summarize(sectionNotes map[string][]string, sections map[string]string) []string {
	out := make([]string, 0, len(sectionNotes))
	titles := make([]string, 0, len(sectionNotes))
	for title := range sectionNotes {
		titles = append(titles, title)
	}
	sort.Strings(titles)
	for _, title := range titles {
		notes := sectionNotes[title]
		if len(notes) == 1 && notes[0] == "Section meets the minimum structure requirements." {
			if content, ok := sections[title]; ok && repeatedHeading(content) {
				out = append(out, fmt.Sprintf("%s: duplicate phrasing detected across subsection headings.", title))
			}
			continue
		}
		out = append(out, fmt.Sprintf("%s: %s", title, joinNotes(notes)))
	}
	if missing := missingCoverage(sections); missing != "" {
		out = append(out, missing)
	}
	if len(out) == 0 {
		return []string{"No major structural gaps detected."}
	}
	return out
}

func joinNotes(notes []string) string {
	result := ""
	for i, note := range notes {
		if i > 0 {
			result += " "
		}
		result += note
	}
	return result
}

func repeatedHeading(content string) bool {
	lines := strings.Split(content, "\n")
	seen := map[string]struct{}{}
	for _, line := range lines {
		line = strings.TrimSpace(strings.ToLower(line))
		if !strings.HasPrefix(line, "### ") {
			continue
		}
		if _, ok := seen[line]; ok {
			return true
		}
		seen[line] = struct{}{}
	}
	return false
}

func missingCoverage(sections map[string]string) string {
	for title, content := range sections {
		lower := strings.ToLower(content)
		if !strings.Contains(lower, "implementation") || !strings.Contains(lower, "risk") {
			return fmt.Sprintf("%s: global review found incomplete coverage across implementation and risk guidance.", title)
		}
	}
	return ""
}
