package agents

import (
	"fmt"
	"strings"
	"sync"

	contextutil "xpert/internal/context"
	"xpert/internal/llm"
	"xpert/internal/storage"
)

// ModelUsageTracker tracks models used during generation.
type ModelUsageTracker struct {
	mu         sync.Mutex
	modelsUsed map[string]int
	fallbacks  int
}

// NewModelUsageTracker creates a new tracker.
func NewModelUsageTracker() *ModelUsageTracker {
	return &ModelUsageTracker{
		modelsUsed: make(map[string]int),
	}
}

// Record records model usage from a generation result.
func (t *ModelUsageTracker) Record(result *llm.GenerateResult) {
	if result == nil {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()

	t.modelsUsed[result.ModelUsed]++
	if result.Attempts > 1 {
		t.fallbacks += result.Attempts - 1
	}
}

// GetStats returns models used and fallback count.
func (t *ModelUsageTracker) GetStats() ([]string, int) {
	t.mu.Lock()
	defer t.mu.Unlock()

	models := make([]string, 0, len(t.modelsUsed))
	for m := range t.modelsUsed {
		models = append(models, m)
	}
	return models, t.fallbacks
}

// ExpertAgent generates document content using LLM.
type ExpertAgent struct {
	router  *llm.Router
	tracker *ModelUsageTracker
	profile AgentProfile
}

// NewExpertAgent creates a new ExpertAgent with the given router.
func NewExpertAgent(router *llm.Router) *ExpertAgent {
	return &ExpertAgent{
		router:  router,
		tracker: NewModelUsageTracker(),
	}
}

// NewExpertAgentWithProfile creates a new ExpertAgent with a shared tracker and profile.
func NewExpertAgentWithProfile(router *llm.Router, tracker *ModelUsageTracker, profile AgentProfile) *ExpertAgent {
	if tracker == nil {
		tracker = NewModelUsageTracker()
	}
	return &ExpertAgent{
		router:  router,
		tracker: tracker,
		profile: profile,
	}
}

// WriteSubsection generates content for a single subsection.
func (a *ExpertAgent) WriteSubsection(request storage.DocumentRequest, section storage.SectionPlan, subsectionTitle string) (string, error) {
	systemPrompt := "You are a focused domain specialist drafting one subsection of a larger implementation-ready document."
	if a.profile.Expertise != "" {
		systemPrompt = fmt.Sprintf("%s Your expertise: %s.", systemPrompt, a.profile.Expertise)
	}
	if a.profile.Name != "" {
		systemPrompt = fmt.Sprintf("You are %s. %s", a.profile.Name, systemPrompt)
	}
	userPrompt := fmt.Sprintf("Document type: %s\nTone: %s\nSection: %s\nSubsection: %s\nObjective: %s\nOriginal request: %s\n",
		request.DocumentType, request.Tone, section.Title, subsectionTitle, section.Objective, contextutil.CompressPrompt(request.Prompt, 1200))

	result, err := a.router.GenerateWithMetadata(systemPrompt, userPrompt)
	if err != nil {
		return "", err
	}

	a.tracker.Record(result)
	return result.Content, nil
}

// ReviseSection revises a section based on reviewer feedback.
func (a *ExpertAgent) ReviseSection(request storage.DocumentRequest, section storage.SectionPlan, content string, reviewNotes []string) (string, error) {
	systemPrompt := "You revise generated markdown so it addresses reviewer feedback while preserving structure."
	if a.profile.Expertise != "" {
		systemPrompt = fmt.Sprintf("%s Your expertise: %s.", systemPrompt, a.profile.Expertise)
	}
	if a.profile.Name != "" {
		systemPrompt = fmt.Sprintf("You are %s. %s", a.profile.Name, systemPrompt)
	}
	userPrompt := fmt.Sprintf("Document type: %s\nSection: %s\nReviewer notes:\n- %s\n\nCurrent draft:\n%s\n",
		request.DocumentType, section.Title, strings.Join(reviewNotes, "\n- "), content)

	result, err := a.router.GenerateWithMetadata(systemPrompt, userPrompt)
	if err != nil {
		return "", err
	}

	a.tracker.Record(result)
	return result.Content, nil
}

// GetModelUsageStats returns models used and fallback count.
func (a *ExpertAgent) GetModelUsageStats() ([]string, int) {
	return a.tracker.GetStats()
}

// ResetTracker resets the usage tracker (call before each pipeline run).
func (a *ExpertAgent) ResetTracker() {
	a.tracker = NewModelUsageTracker()
}
