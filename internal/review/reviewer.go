package review

import (
	"fmt"
	"strings"
)

type Reviewer struct{}

func NewReviewer() *Reviewer { return &Reviewer{} }

func (r *Reviewer) Review(content string) []string {
	notes := make([]string, 0, 5)
	lower := strings.ToLower(content)
	if !strings.Contains(lower, "risk") {
		notes = append(notes, "Add explicit risk handling language.")
	}
	if !strings.Contains(lower, "implementation") {
		notes = append(notes, "Strengthen implementation detail.")
	}
	if !strings.Contains(lower, "validation") && !strings.Contains(lower, "verify") {
		notes = append(notes, "Add validation or verification guidance.")
	}
	if strings.Count(lower, "### ") < 2 {
		notes = append(notes, "Expand subsection coverage to avoid thin sections.")
	}
	if strings.Count(content, "\n") < 12 {
		notes = append(notes, fmt.Sprintf("Increase section depth; current section is only %d lines.", strings.Count(content, "\n")+1))
	}
	if len(notes) == 0 {
		return []string{"Section meets the minimum structure requirements."}
	}
	return notes
}
