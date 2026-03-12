package synthesis

import (
	"fmt"
	"strings"
)

type GlobalSynthesizer struct{}

func NewGlobalSynthesizer() *GlobalSynthesizer { return &GlobalSynthesizer{} }

func (s *GlobalSynthesizer) Assemble(title string, sections []string, reviewSummary []string) string {
	content := fmt.Sprintf("## Document Overview\n\nGenerated title: %s\n\n## Quality Summary\n", title)
	for _, note := range reviewSummary {
		content += fmt.Sprintf("- %s\n", note)
	}
	content += "\n## Section Index\n"
	for _, section := range sections {
		if heading := firstHeading(section); heading != "" {
			content += "- " + heading + "\n"
		}
	}
	content += "\n"
	for _, section := range sections {
		content += section + "\n\n"
	}
	return content
}

func firstHeading(section string) string {
	for _, line := range strings.Split(section, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "## ") {
			return strings.TrimPrefix(line, "## ")
		}
	}
	return ""
}
