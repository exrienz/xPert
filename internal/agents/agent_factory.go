package agents

import (
	"fmt"
	"strings"

	"xpert/internal/llm"
	"xpert/internal/storage"
)

type AgentFactory struct {
	router  *llm.Router
	tracker *ModelUsageTracker
}

func NewAgentFactory(router *llm.Router) *AgentFactory {
	return &AgentFactory{
		router:  router,
		tracker: NewModelUsageTracker(),
	}
}

func (f *AgentFactory) ResetTracker() {
	f.tracker = NewModelUsageTracker()
}

func (f *AgentFactory) GetModelUsageStats() ([]string, int) {
	return f.tracker.GetStats()
}

func (f *AgentFactory) ForSubsection(section storage.SectionPlan, subsectionTitle string) *ExpertAgent {
	expertise := strings.TrimSpace(section.Objective)
	if expertise == "" {
		expertise = subsectionTitle
	}
	name := fmt.Sprintf("%s Specialist", subsectionTitle)
	profile := AgentProfile{
		Name:      name,
		Expertise: expertise,
	}
	return NewExpertAgentWithProfile(f.router, f.tracker, profile)
}

func (f *AgentFactory) ForSectionRevision(section storage.SectionPlan) *ExpertAgent {
	expertise := strings.TrimSpace(section.Objective)
	if expertise == "" {
		expertise = section.Title
	}
	name := fmt.Sprintf("%s Reviewer", section.Title)
	profile := AgentProfile{
		Name:      name,
		Expertise: expertise,
	}
	return NewExpertAgentWithProfile(f.router, f.tracker, profile)
}
