package storage

import (
	"path/filepath"
	"testing"
	"time"
)

func TestSQLiteStoreRoundTrip(t *testing.T) {
	store := NewSQLiteRepository(filepath.Join(t.TempDir(), "docgen.sqlite"))
	if err := store.Load(); err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	now := time.Now().UTC().Truncate(time.Second)
	job := JobRecord{
		ID:          "job-1",
		DocumentID:  "doc-1",
		Status:      JobQueued,
		Stage:       StagePending,
		Progress:    0,
		MaxAttempts: 2,
		Request: DocumentRequest{
			Prompt:          "Generate a runbook for service rollout.",
			DocumentType:    "Runbook",
			TargetWordCount: 3000,
			Tone:            "Operational",
			OutputFormat:    "markdown",
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := store.CreateJob(job); err != nil {
		t.Fatalf("CreateJob returned error: %v", err)
	}
	if got, ok := store.GetJob(job.ID); !ok || got.Request.Prompt != job.Request.Prompt {
		t.Fatalf("GetJob mismatch: %#v, found=%v", got, ok)
	}

	trace := PipelineTrace{
		Intent:             "operational_guidance",
		DocumentType:       "Runbook",
		DocumentFocus:      "service rollout",
		GlobalContext:      []string{"service", "rollout"},
		Terminology:        []string{"deployment", "rollback"},
		SectionSummaries:   map[string]string{"Execution Workflow": "Explains deployment and rollback."},
		StructureChecklist: []string{"Context", "Implementation", "Operations"},
	}
	if err := store.SaveTrace(job.ID, trace); err != nil {
		t.Fatalf("SaveTrace returned error: %v", err)
	}
	if got, ok := store.GetTrace(job.ID); !ok || got.DocumentType != trace.DocumentType {
		t.Fatalf("GetTrace mismatch: %#v, found=%v", got, ok)
	}

	document := DocumentRecord{
		ID:            "doc-1",
		JobID:         "job-1",
		Request:       job.Request,
		Title:         "Runbook: rollout",
		Markdown:      "# Rollout",
		Content:       "# Rollout",
		OutputFormat:  "markdown",
		SectionTitles: []string{"Context", "Execution Workflow"},
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := store.SaveDocument(document); err != nil {
		t.Fatalf("SaveDocument returned error: %v", err)
	}
	if got, ok := store.GetDocument(document.ID); !ok || got.Title != document.Title {
		t.Fatalf("GetDocument mismatch: %#v, found=%v", got, ok)
	}
}
