package agents

import (
	"fmt"
	"strings"

	contextutil "docgen/internal/context"
	"docgen/internal/llm"
	"docgen/internal/storage"
)

type ExpertAgent struct {
	router *llm.Router
}

func NewExpertAgent(router *llm.Router) *ExpertAgent {
	return &ExpertAgent{router: router}
}

func (a *ExpertAgent) WriteSubsection(request storage.DocumentRequest, section storage.SectionPlan, subsectionTitle string) (string, error) {
	systemPrompt := "You are a focused domain specialist drafting one subsection of a larger implementation-ready document."
	userPrompt := fmt.Sprintf("Document type: %s\nTone: %s\nSection: %s\nSubsection: %s\nObjective: %s\nOriginal request: %s\n",
		request.DocumentType, request.Tone, section.Title, subsectionTitle, section.Objective, contextutil.CompressPrompt(request.Prompt, 1200))
	return a.router.Generate(systemPrompt, userPrompt)
}

func (a *ExpertAgent) ReviseSection(request storage.DocumentRequest, section storage.SectionPlan, content string, reviewNotes []string) (string, error) {
	systemPrompt := "You revise generated markdown so it addresses reviewer feedback while preserving structure."
	userPrompt := fmt.Sprintf("Document type: %s\nSection: %s\nReviewer notes:\n- %s\n\nCurrent draft:\n%s\n",
		request.DocumentType, section.Title, strings.Join(reviewNotes, "\n- "), content)
	return a.router.Generate(systemPrompt, userPrompt)
}
