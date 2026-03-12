package structure

import (
	"fmt"
	"strings"

	"docgen/internal/storage"
)

type DocumentStructurer struct{}

func NewDocumentStructurer() *DocumentStructurer { return &DocumentStructurer{} }

func (s *DocumentStructurer) Checklist(documentType string) []string {
	lower := strings.ToLower(documentType)
	switch {
	case strings.Contains(lower, "sop"), strings.Contains(lower, "procedure"):
		return []string{"Objective", "Scope", "Inputs", "Procedure", "Tools", "Expected Output", "Validation"}
	case strings.Contains(lower, "tutorial"):
		return []string{"Concept", "Explanation", "Example", "Summary"}
	case strings.Contains(lower, "report"), strings.Contains(lower, "assessment"):
		return []string{"Abstract", "Methodology", "Analysis", "Conclusion"}
	default:
		return []string{"Context", "Implementation", "Operations", "Risks", "Validation"}
	}
}

func (s *DocumentStructurer) EnrichSectionPlan(documentType string, plan storage.SectionPlan) storage.SectionPlan {
	plan.RequiredElements = s.Checklist(documentType)
	return plan
}

func (s *DocumentStructurer) Apply(documentType, title, content string) string {
	elements := s.Checklist(documentType)
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("# %s\n\n", title))
	builder.WriteString("## Structured Outline\n")
	for _, element := range elements {
		builder.WriteString("- " + element + "\n")
	}
	builder.WriteString("\n")
	builder.WriteString(content)
	return builder.String()
}
