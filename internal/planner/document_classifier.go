package planner

import "strings"

type DocumentClassifier struct{}

func NewDocumentClassifier() *DocumentClassifier { return &DocumentClassifier{} }

func (c *DocumentClassifier) Classify(documentType, prompt string) (string, []string) {
	normalized := strings.ToLower(documentType + " " + prompt)
	switch {
	case strings.Contains(normalized, "sop"), strings.Contains(normalized, "procedure"):
		return "SOP Manual", classifyTerms(prompt)
	case strings.Contains(normalized, "runbook"):
		return "Operations Runbook", classifyTerms(prompt)
	case strings.Contains(normalized, "playbook"):
		return "Security Playbook", classifyTerms(prompt)
	case strings.Contains(normalized, "report"), strings.Contains(normalized, "assessment"):
		return "Assessment Report", classifyTerms(prompt)
	default:
		return documentType, classifyTerms(prompt)
	}
}

func classifyTerms(prompt string) []string {
	parts := strings.Fields(strings.ToLower(prompt))
	seen := map[string]struct{}{}
	out := make([]string, 0, 6)
	for _, part := range parts {
		part = strings.Trim(part, ".,:;!?()[]{}\"'")
		if len(part) < 5 {
			continue
		}
		if _, ok := seen[part]; ok {
			continue
		}
		seen[part] = struct{}{}
		out = append(out, part)
		if len(out) == 6 {
			break
		}
	}
	return out
}
