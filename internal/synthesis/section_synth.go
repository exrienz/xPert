package synthesis

import "fmt"

type SectionSynthesizer struct{}

func NewSectionSynthesizer() *SectionSynthesizer { return &SectionSynthesizer{} }

func (s *SectionSynthesizer) Synthesize(title string, content string, reviewNotes []string) string {
	block := "Review notes:\n"
	for _, note := range reviewNotes {
		block += fmt.Sprintf("- %s\n", note)
	}
	return fmt.Sprintf("## %s\n\n%s\n\n### Review Consolidation\n%s", title, content, block)
}
