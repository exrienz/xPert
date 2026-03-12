package storage

import "time"

type JobStatus string

const (
	JobQueued    JobStatus = "queued"
	JobRunning   JobStatus = "running"
	JobCompleted JobStatus = "completed"
	JobFailed    JobStatus = "failed"
	JobCanceled  JobStatus = "canceled"
)

type JobStage string

const (
	StagePending      JobStage = "pending"
	StagePlanning     JobStage = "planning"
	StageGenerating   JobStage = "generating"
	StageReviewing    JobStage = "reviewing"
	StageSynthesizing JobStage = "synthesizing"
	StageFormatting   JobStage = "formatting"
	StageCompleted    JobStage = "completed"
	StageFailed       JobStage = "failed"
	StageCanceled     JobStage = "canceled"
)

type DocumentRequest struct {
	Prompt          string `json:"prompt"`
	DocumentType    string `json:"document_type"`
	TargetWordCount int    `json:"target_word_count"`
	Tone            string `json:"tone"`
	OutputFormat    string `json:"output_format"`
}

type BatchDocumentRequest struct {
	Requests []DocumentRequest `json:"requests"`
}

type SectionPlan struct {
	Title            string   `json:"title"`
	Objective        string   `json:"objective"`
	SubsectionTitles []string `json:"subsection_titles"`
	Keywords         []string `json:"keywords"`
	Summary          string   `json:"summary,omitempty"`
	Terminology      []string `json:"terminology,omitempty"`
	RequiredElements []string `json:"required_elements,omitempty"`
}

type PipelineTrace struct {
	Intent             string            `json:"intent"`
	DocumentType       string            `json:"document_type"`
	DocumentFocus      string            `json:"document_focus"`
	GlobalContext      []string          `json:"global_context"`
	Terminology        []string          `json:"terminology,omitempty"`
	SectionSummaries   map[string]string `json:"section_summaries,omitempty"`
	StructureChecklist []string          `json:"structure_checklist,omitempty"`
	Sections           []SectionPlan     `json:"sections"`
	ReviewSummary      []string          `json:"review_summary"`
}

type JobRecord struct {
	ID           string          `json:"id"`
	DocumentID   string          `json:"document_id"`
	Status       JobStatus       `json:"status"`
	Stage        JobStage        `json:"stage"`
	Progress     int             `json:"progress"`
	AttemptCount int             `json:"attempt_count"`
	MaxAttempts  int             `json:"max_attempts"`
	BatchID      string          `json:"batch_id,omitempty"`
	Request      DocumentRequest `json:"request"`
	Result       string          `json:"result,omitempty"`
	Error        string          `json:"error,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
	StartedAt    *time.Time      `json:"started_at,omitempty"`
	CompletedAt  *time.Time      `json:"completed_at,omitempty"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

type BatchJobRecord struct {
	ID           string            `json:"id"`
	Status       JobStatus         `json:"status"`
	Stage        JobStage          `json:"stage"`
	Progress     int               `json:"progress"`
	AttemptCount int               `json:"attempt_count"`
	Requests     []DocumentRequest `json:"requests"`
	ChildJobIDs  []string          `json:"child_job_ids"`
	CreatedAt    time.Time         `json:"created_at"`
	StartedAt    *time.Time        `json:"started_at,omitempty"`
	CompletedAt  *time.Time        `json:"completed_at,omitempty"`
	UpdatedAt    time.Time         `json:"updated_at"`
}

type DocumentRecord struct {
	ID            string          `json:"id"`
	JobID         string          `json:"job_id"`
	Request       DocumentRequest `json:"request"`
	Title         string          `json:"title"`
	Markdown      string          `json:"markdown"`
	Content       string          `json:"content"`
	OutputFormat  string          `json:"output_format"`
	SectionTitles []string        `json:"section_titles"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

type JobDetailResponse struct {
	Job      JobRecord       `json:"job"`
	Document *DocumentRecord `json:"document,omitempty"`
	Trace    *PipelineTrace  `json:"trace,omitempty"`
}

type DocumentDetailResponse struct {
	Document DocumentRecord `json:"document"`
	Job      *JobRecord     `json:"job,omitempty"`
	Trace    *PipelineTrace `json:"trace,omitempty"`
}

type BatchJobDetailResponse struct {
	Batch BatchJobRecord `json:"batch"`
	Jobs  []JobRecord    `json:"jobs"`
}

type CreateJobResponse struct {
	ID         string    `json:"id"`
	Status     JobStatus `json:"status"`
	DocumentID string    `json:"document_id"`
}
