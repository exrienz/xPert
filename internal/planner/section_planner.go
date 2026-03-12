package planner

import (
	"fmt"
	"strings"

	"xpert/internal/storage"
)

type SectionPlanner struct{}

func NewSectionPlanner() *SectionPlanner { return &SectionPlanner{} }

func (p *SectionPlanner) Expand(request storage.DocumentRequest, section storage.SectionPlan) storage.SectionPlan {
	focus := strings.ToLower(request.DocumentType)
	if len(section.Keywords) > 0 {
		focus = strings.Join(section.Keywords, ", ")
	}
	switch {
	case strings.Contains(strings.ToLower(section.Title), "purpose"), strings.Contains(strings.ToLower(section.Title), "summary"):
		section.SubsectionTitles = []string{
			fmt.Sprintf("%s: scope and outcomes", section.Title),
			fmt.Sprintf("%s: business and technical context for %s", section.Title, focus),
			fmt.Sprintf("%s: success criteria and exclusions", section.Title),
		}
	case strings.Contains(strings.ToLower(section.Title), "procedure"), strings.Contains(strings.ToLower(section.Title), "workflow"):
		section.SubsectionTitles = []string{
			fmt.Sprintf("%s: preparation for %s", section.Title, focus),
			fmt.Sprintf("%s: execution steps and checkpoints", section.Title),
			fmt.Sprintf("%s: rollback, exceptions, and escalation", section.Title),
		}
	default:
		section.SubsectionTitles = []string{
			fmt.Sprintf("%s: intent and scope for %s", section.Title, focus),
			fmt.Sprintf("%s: implementation approach and dependencies", section.Title),
			fmt.Sprintf("%s: risks, controls, and operational checks", section.Title),
		}
	}
	return section
}
