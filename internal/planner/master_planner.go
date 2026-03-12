package planner

import (
	"fmt"
	"regexp"
	"strings"

	"docgen/internal/storage"
)

type MasterPlanner struct{}

func NewMasterPlanner() *MasterPlanner { return &MasterPlanner{} }

func (p *MasterPlanner) Plan(request storage.DocumentRequest) []storage.SectionPlan {
	defaults := p.defaultsForType(request.DocumentType)
	focusTerms := p.extractFocusTerms(request.Prompt)
	sectionCount := request.TargetWordCount / 1200
	if sectionCount < 4 {
		sectionCount = 4
	}
	if sectionCount > 10 {
		sectionCount = 10
	}

	sections := make([]storage.SectionPlan, 0, sectionCount)
	for i := 0; i < sectionCount; i++ {
		title := defaults[i%len(defaults)]
		keywords := focusTerms
		if len(keywords) > 3 {
			keywords = keywords[:3]
		}
		sections = append(sections, storage.SectionPlan{
			Title:     title,
			Objective: fmt.Sprintf("Explain how %s supports the requested document about: %s", strings.ToLower(title), request.Prompt),
			Keywords:  keywords,
		})
	}
	return sections
}

func (p *MasterPlanner) defaultsForType(documentType string) []string {
	normalized := strings.ToLower(documentType)
	switch {
	case strings.Contains(normalized, "sop"), strings.Contains(normalized, "procedure"):
		return []string{"Purpose and Scope", "Roles and Responsibilities", "Preconditions and Inputs", "Step-by-Step Procedure", "Decision Points and Escalations", "Controls and Risk Handling", "Validation and Quality Checks", "Reporting and Evidence", "Maintenance and Updates", "Appendix"}
	case strings.Contains(normalized, "runbook"), strings.Contains(normalized, "playbook"):
		return []string{"Mission Overview", "Activation Criteria", "Environment and Dependencies", "Execution Workflow", "Tooling and Automation", "Failure Modes and Recovery", "Security and Guardrails", "Observability and Reporting", "Operational Readiness", "Appendix"}
	default:
		return []string{"Executive Summary", "Problem Framing", "Target Architecture", "Workflow Design", "Components and Interfaces", "Data and State", "Operations and Support", "Security and Governance", "Delivery Plan", "Appendix"}
	}
}

func (p *MasterPlanner) extractFocusTerms(prompt string) []string {
	re := regexp.MustCompile(`[A-Za-z][A-Za-z0-9_-]{4,}`)
	words := re.FindAllString(strings.ToLower(prompt), -1)
	stop := map[string]struct{}{"generate": {}, "detailed": {}, "technical": {}, "document": {}, "implementation": {}, "guide": {}, "manual": {}, "internal": {}, "their": {}, "about": {}, "with": {}}
	seen := map[string]struct{}{}
	terms := make([]string, 0, 8)
	for _, word := range words {
		if _, ok := stop[word]; ok {
			continue
		}
		if _, ok := seen[word]; ok {
			continue
		}
		seen[word] = struct{}{}
		terms = append(terms, word)
	}
	return terms
}
