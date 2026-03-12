package orchestrator

import (
	"testing"

	"docgen/internal/queue"
	"docgen/internal/storage"
)

func TestNormalizeRequestDefaults(t *testing.T) {
	request := normalizeRequest(storage.DocumentRequest{})
	if request.Prompt == "" {
		t.Fatal("expected default prompt")
	}
	if request.DocumentType == "" {
		t.Fatal("expected default document type")
	}
	if request.TargetWordCount == 0 {
		t.Fatal("expected default target word count")
	}
	if request.OutputFormat == "" {
		t.Fatal("expected default output format")
	}
}

func TestFailJobRequeuesUntilMaxAttempts(t *testing.T) {
	repo := storage.NewRepository(t.TempDir() + "/jobs.json")
	if err := repo.Load(); err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	workQueue := queue.NewMemoryQueue(2)
	manager := NewJobManager(repo, workQueue, nil, 2)

	job, err := manager.CreateJob(storage.DocumentRequest{Prompt: "test"})
	if err != nil {
		t.Fatalf("CreateJob returned error: %v", err)
	}
	job.AttemptCount = 1
	manager.failJob(job, errTest)

	updated, ok := repo.GetJob(job.ID)
	if !ok {
		t.Fatal("expected stored job")
	}
	if updated.Status != storage.JobQueued {
		t.Fatalf("expected queued job, got %s", updated.Status)
	}
}

type staticError string

func (e staticError) Error() string { return string(e) }

const errTest = staticError("boom")
