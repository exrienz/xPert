package storage

type Store interface {
	Load() error
	CreateJob(job JobRecord) error
	UpdateJob(job JobRecord) error
	GetJob(id string) (JobRecord, bool)
	ListJobs(limit int) []JobRecord
	CreateBatchJob(batch BatchJobRecord) error
	UpdateBatchJob(batch BatchJobRecord) error
	GetBatchJob(id string) (BatchJobRecord, bool)
	SaveDocument(document DocumentRecord) error
	GetDocument(id string) (DocumentRecord, bool)
	GetDocumentByJob(jobID string) (DocumentRecord, bool)
	ListDocuments(limit int) []DocumentRecord
	SaveTrace(jobID string, trace PipelineTrace) error
	GetTrace(jobID string) (PipelineTrace, bool)
}
