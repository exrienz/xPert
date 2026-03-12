package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"sync"
)

type repositoryState struct {
	Jobs      map[string]JobRecord      `json:"jobs"`
	BatchJobs map[string]BatchJobRecord `json:"batch_jobs"`
	Documents map[string]DocumentRecord `json:"documents"`
	Traces    map[string]PipelineTrace  `json:"traces"`
}

type Repository struct {
	path  string
	mu    sync.RWMutex
	state repositoryState
}

func NewRepository(path string) *Repository {
	return &Repository{
		path: path,
		state: repositoryState{
			Jobs:      map[string]JobRecord{},
			BatchJobs: map[string]BatchJobRecord{},
			Documents: map[string]DocumentRecord{},
			Traces:    map[string]PipelineTrace{},
		},
	}
}

func (r *Repository) Load() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := os.MkdirAll(filepath.Dir(r.path), 0o755); err != nil {
		return err
	}

	data, err := os.ReadFile(r.path)
	if err != nil {
		if os.IsNotExist(err) {
			return r.flushLocked()
		}
		return err
	}
	if len(data) == 0 {
		return nil
	}

	return json.Unmarshal(data, &r.state)
}

func (r *Repository) CreateJob(job JobRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.state.Jobs[job.ID] = job
	return r.flushLocked()
}

func (r *Repository) UpdateJob(job JobRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.state.Jobs[job.ID] = job
	return r.flushLocked()
}

func (r *Repository) GetJob(id string) (JobRecord, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	job, ok := r.state.Jobs[id]
	return job, ok
}

func (r *Repository) ListJobs(limit int) []JobRecord {
	r.mu.RLock()
	defer r.mu.RUnlock()
	items := make([]JobRecord, 0, len(r.state.Jobs))
	for _, job := range r.state.Jobs {
		items = append(items, job)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})
	if limit > 0 && len(items) > limit {
		items = items[:limit]
	}
	return items
}

func (r *Repository) CreateBatchJob(batch BatchJobRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.state.BatchJobs[batch.ID] = batch
	return r.flushLocked()
}

func (r *Repository) UpdateBatchJob(batch BatchJobRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.state.BatchJobs[batch.ID] = batch
	return r.flushLocked()
}

func (r *Repository) GetBatchJob(id string) (BatchJobRecord, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	batch, ok := r.state.BatchJobs[id]
	return batch, ok
}

func (r *Repository) SaveDocument(document DocumentRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.state.Documents[document.ID] = document
	return r.flushLocked()
}

func (r *Repository) GetDocument(id string) (DocumentRecord, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	document, ok := r.state.Documents[id]
	return document, ok
}

func (r *Repository) GetDocumentByJob(jobID string) (DocumentRecord, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, document := range r.state.Documents {
		if document.JobID == jobID {
			return document, true
		}
	}
	return DocumentRecord{}, false
}

func (r *Repository) ListDocuments(limit int) []DocumentRecord {
	r.mu.RLock()
	defer r.mu.RUnlock()
	items := make([]DocumentRecord, 0, len(r.state.Documents))
	for _, document := range r.state.Documents {
		items = append(items, document)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})
	if limit > 0 && len(items) > limit {
		items = items[:limit]
	}
	return items
}

func (r *Repository) SaveTrace(jobID string, trace PipelineTrace) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.state.Traces[jobID] = trace
	return r.flushLocked()
}

func (r *Repository) GetTrace(jobID string) (PipelineTrace, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	trace, ok := r.state.Traces[jobID]
	return trace, ok
}

func (r *Repository) flushLocked() error {
	payload, err := json.MarshalIndent(r.state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(r.path, payload, 0o644)
}
