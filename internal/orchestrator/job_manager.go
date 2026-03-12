package orchestrator

import (
	"log"
	"strings"
	"time"

	"xpert/internal/queue"
	"xpert/internal/storage"
)

type JobManager struct {
	repository     storage.Store
	queue          queue.Queue
	pipeline       *Pipeline
	maxJobAttempts int
}

func NewJobManager(repository storage.Store, workQueue queue.Queue, pipeline *Pipeline, maxJobAttempts int) *JobManager {
	return &JobManager{
		repository:     repository,
		queue:          workQueue,
		pipeline:       pipeline,
		maxJobAttempts: maxJobAttempts,
	}
}

func (m *JobManager) CreateJob(request storage.DocumentRequest) (storage.JobRecord, error) {
	now := time.Now().UTC()
	job := storage.JobRecord{
		ID:          newID(),
		DocumentID:  newID(),
		Status:      storage.JobQueued,
		Stage:       storage.StagePending,
		Progress:    0,
		MaxAttempts: m.maxJobAttempts,
		Request:     normalizeRequest(request),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := m.repository.CreateJob(job); err != nil {
		return storage.JobRecord{}, err
	}
	m.queue.Put(job.ID)
	return job, nil
}

func (m *JobManager) CreateBatchJob(request storage.BatchDocumentRequest) (storage.BatchJobRecord, error) {
	now := time.Now().UTC()
	batch := storage.BatchJobRecord{
		ID:        newID(),
		Status:    storage.JobQueued,
		Stage:     storage.StagePending,
		Progress:  0,
		Requests:  request.Requests,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := m.repository.CreateBatchJob(batch); err != nil {
		return storage.BatchJobRecord{}, err
	}

	childIDs := make([]string, 0, len(request.Requests))
	for _, item := range request.Requests {
		job, err := m.CreateJob(item)
		if err != nil {
			return storage.BatchJobRecord{}, err
		}
		job.BatchID = batch.ID
		_ = m.repository.UpdateJob(job)
		childIDs = append(childIDs, job.ID)
	}
	batch.ChildJobIDs = childIDs
	batch.UpdatedAt = time.Now().UTC()
	_ = m.repository.UpdateBatchJob(batch)
	return batch, nil
}

func (m *JobManager) ProcessJob(jobID string) {
	job, ok := m.repository.GetJob(jobID)
	if !ok || job.Status == storage.JobCanceled {
		return
	}

	now := time.Now().UTC()
	job.AttemptCount++
	job.Status = storage.JobRunning
	job.Stage = storage.StagePlanning
	job.Progress = 10
	job.StartedAt = &now
	job.UpdatedAt = now
	_ = m.repository.UpdateJob(job)
	log.Printf("job %s started stage=%s progress=%d attempt=%d/%d", job.ID, job.Stage, job.Progress, job.AttemptCount, job.MaxAttempts)

	report := func(stage storage.JobStage, progress int) {
		now := time.Now().UTC()
		job.Stage = stage
		job.Progress = progress
		job.UpdatedAt = now
		_ = m.repository.UpdateJob(job)
		log.Printf("job %s stage=%s progress=%d", job.ID, job.Stage, job.Progress)
	}

	markdown, formatted, trace, err := m.pipeline.Run(job.Request, report)
	if err != nil {
		log.Printf("job %s failed: %v", job.ID, err)
		m.failJob(job, err)
		return
	}
	completed := time.Now().UTC()
	job.Status = storage.JobCompleted
	job.Stage = storage.StageCompleted
	job.Progress = 100
	job.Result = formatted
	job.CompletedAt = &completed
	job.UpdatedAt = completed
	_ = m.repository.UpdateJob(job)
	log.Printf("job %s completed", job.ID)
	_ = m.repository.SaveTrace(job.ID, trace)
	_ = m.repository.SaveDocument(storage.DocumentRecord{
		ID:            job.DocumentID,
		JobID:         job.ID,
		Request:       job.Request,
		Title:         trace.DocumentType + ": " + truncate(job.Request.Prompt, 80),
		Markdown:      markdown,
		Content:       formatted,
		OutputFormat:  job.Request.OutputFormat,
		SectionTitles: sectionTitles(trace.Sections),
		CreatedAt:     job.CreatedAt,
		UpdatedAt:     completed,
	})

	if job.BatchID != "" {
		m.refreshBatch(job.BatchID)
	}
}

func (m *JobManager) failJob(job storage.JobRecord, err error) {
	now := time.Now().UTC()
	job.Error = err.Error()
	job.UpdatedAt = now
	log.Printf("job %s attempt %d/%d error: %s", job.ID, job.AttemptCount, job.MaxAttempts, job.Error)
	if job.AttemptCount < job.MaxAttempts {
		job.Status = storage.JobQueued
		job.Stage = storage.StagePending
		job.Progress = 0
		_ = m.repository.UpdateJob(job)
		m.queue.Put(job.ID)
		return
	}
	job.Status = storage.JobFailed
	job.Stage = storage.StageFailed
	job.Progress = 100
	job.CompletedAt = &now
	_ = m.repository.UpdateJob(job)
	if job.BatchID != "" {
		m.refreshBatch(job.BatchID)
	}
}

func (m *JobManager) GetJobDetail(jobID string) (storage.JobDetailResponse, bool) {
	job, ok := m.repository.GetJob(jobID)
	if !ok {
		return storage.JobDetailResponse{}, false
	}
	document, hasDocument := m.repository.GetDocument(job.DocumentID)
	trace, hasTrace := m.repository.GetTrace(job.ID)
	response := storage.JobDetailResponse{Job: job}
	if hasDocument {
		response.Document = &document
	}
	if hasTrace {
		response.Trace = &trace
	}
	return response, true
}

func (m *JobManager) GetDocumentDetail(documentID string) (storage.DocumentDetailResponse, bool) {
	document, ok := m.repository.GetDocument(documentID)
	if !ok {
		return storage.DocumentDetailResponse{}, false
	}
	response := storage.DocumentDetailResponse{Document: document}
	if job, ok := m.repository.GetJob(document.JobID); ok {
		response.Job = &job
		if trace, ok := m.repository.GetTrace(job.ID); ok {
			response.Trace = &trace
		}
	}
	return response, true
}

func (m *JobManager) GetBatchDetail(batchID string) (storage.BatchJobDetailResponse, bool) {
	batch, ok := m.repository.GetBatchJob(batchID)
	if !ok {
		return storage.BatchJobDetailResponse{}, false
	}
	jobs := make([]storage.JobRecord, 0, len(batch.ChildJobIDs))
	for _, jobID := range batch.ChildJobIDs {
		if job, ok := m.repository.GetJob(jobID); ok {
			jobs = append(jobs, job)
		}
	}
	return storage.BatchJobDetailResponse{Batch: batch, Jobs: jobs}, true
}

func (m *JobManager) CancelJob(jobID string) (storage.JobRecord, bool) {
	job, ok := m.repository.GetJob(jobID)
	if !ok {
		return storage.JobRecord{}, false
	}
	now := time.Now().UTC()
	job.Status = storage.JobCanceled
	job.Stage = storage.StageCanceled
	job.Progress = 100
	job.Error = "Job canceled."
	job.CompletedAt = &now
	job.UpdatedAt = now
	_ = m.repository.UpdateJob(job)
	return job, true
}

func (m *JobManager) refreshBatch(batchID string) {
	batch, ok := m.repository.GetBatchJob(batchID)
	if !ok {
		return
	}
	completedCount := 0
	failedCount := 0
	for _, childID := range batch.ChildJobIDs {
		job, ok := m.repository.GetJob(childID)
		if ok && job.Status == storage.JobCompleted {
			completedCount++
		}
		if ok && job.Status == storage.JobFailed {
			failedCount++
		}
	}
	if len(batch.ChildJobIDs) > 0 {
		batch.Progress = completedCount * 100 / len(batch.ChildJobIDs)
	}
	if failedCount > 0 {
		batch.Status = storage.JobFailed
		batch.Stage = storage.StageFailed
		batch.UpdatedAt = time.Now().UTC()
		_ = m.repository.UpdateBatchJob(batch)
		return
	}
	if completedCount == len(batch.ChildJobIDs) && len(batch.ChildJobIDs) > 0 {
		now := time.Now().UTC()
		batch.Status = storage.JobCompleted
		batch.Stage = storage.StageCompleted
		batch.CompletedAt = &now
	}
	batch.UpdatedAt = time.Now().UTC()
	_ = m.repository.UpdateBatchJob(batch)
}

func normalizeRequest(request storage.DocumentRequest) storage.DocumentRequest {
	request.Prompt = trimSpaceOrDefault(request.Prompt, "Generate an implementation-ready technical document.")
	if request.DocumentType == "" {
		request.DocumentType = "Technical Blueprint"
	}
	if request.TargetWordCount == 0 {
		request.TargetWordCount = 6000
	}
	if request.Tone == "" {
		request.Tone = "Executive and implementation-ready"
	}
	if request.OutputFormat == "" {
		request.OutputFormat = "markdown"
	}
	return request
}

func trimSpaceOrDefault(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return strings.TrimSpace(value)
}

func sectionTitles(sections []storage.SectionPlan) []string {
	out := make([]string, 0, len(sections))
	for _, section := range sections {
		out = append(out, section.Title)
	}
	return out
}
