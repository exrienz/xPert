package planner

import "strings"

type IntentDetector struct{}

func NewIntentDetector() *IntentDetector { return &IntentDetector{} }

func (d *IntentDetector) Detect(prompt string) string {
	lower := strings.ToLower(prompt)
	switch {
	case strings.Contains(lower, "analyze"), strings.Contains(lower, "assessment"), strings.Contains(lower, "evaluate"):
		return "analyze_and_document"
	case strings.Contains(lower, "migrate"), strings.Contains(lower, "rollout"), strings.Contains(lower, "deploy"):
		return "implementation_plan"
	case strings.Contains(lower, "runbook"), strings.Contains(lower, "playbook"), strings.Contains(lower, "incident"), strings.Contains(lower, "sop"):
		return "operational_guidance"
	default:
		return "generate_document"
	}
}
