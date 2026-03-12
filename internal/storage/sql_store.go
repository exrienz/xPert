package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/lib/pq"
	_ "modernc.org/sqlite"
)

type SQLStore struct {
	db      *sql.DB
	dialect string
	dsn     string
}

func NewSQLiteRepository(path string) *SQLStore {
	dsn := path
	if strings.TrimSpace(dsn) == "" {
		dsn = "./data/docgen.sqlite"
	}
	return &SQLStore{
		db:      mustOpen("sqlite", dsn),
		dialect: "sqlite",
		dsn:     dsn,
	}
}

func NewPostgresRepository(dsn string) *SQLStore {
	return &SQLStore{
		db:      mustOpen("postgres", dsn),
		dialect: "postgres",
		dsn:     dsn,
	}
}

func mustOpen(driverName, dsn string) *sql.DB {
	db, err := sql.Open(driverName, dsn)
	if err != nil {
		panic(err)
	}
	return db
}

func (s *SQLStore) Load() error {
	if s.dialect == "sqlite" {
		dir := filepath.Dir(s.dsn)
		if dir != "." && dir != "" {
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return err
			}
		}
	}
	if err := s.db.Ping(); err != nil {
		return err
	}
	for _, stmt := range s.schema() {
		if _, err := s.db.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}

func (s *SQLStore) CreateJob(job JobRecord) error {
	payload, err := json.Marshal(job.Request)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(s.rebind(`INSERT INTO jobs (
id, document_id, status, stage, progress, attempt_count, max_attempts, batch_id, request_json, result, error, created_at, started_at, completed_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`),
		job.ID, job.DocumentID, string(job.Status), string(job.Stage), job.Progress, job.AttemptCount, job.MaxAttempts, nullableString(job.BatchID),
		string(payload), nullableString(job.Result), nullableString(job.Error), job.CreatedAt, nullableTime(job.StartedAt), nullableTime(job.CompletedAt), job.UpdatedAt,
	)
	return err
}

func (s *SQLStore) UpdateJob(job JobRecord) error {
	payload, err := json.Marshal(job.Request)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(s.rebind(`UPDATE jobs SET
document_id = ?, status = ?, stage = ?, progress = ?, attempt_count = ?, max_attempts = ?, batch_id = ?, request_json = ?, result = ?, error = ?, created_at = ?, started_at = ?, completed_at = ?, updated_at = ?
WHERE id = ?`),
		job.DocumentID, string(job.Status), string(job.Stage), job.Progress, job.AttemptCount, job.MaxAttempts, nullableString(job.BatchID),
		string(payload), nullableString(job.Result), nullableString(job.Error), job.CreatedAt, nullableTime(job.StartedAt), nullableTime(job.CompletedAt), job.UpdatedAt, job.ID,
	)
	return err
}

func (s *SQLStore) GetJob(id string) (JobRecord, bool) {
	row := s.db.QueryRow(s.rebind(`SELECT id, document_id, status, stage, progress, attempt_count, max_attempts, batch_id, request_json, result, error, created_at, started_at, completed_at, updated_at FROM jobs WHERE id = ?`), id)
	record, err := scanJob(row)
	return record, err == nil
}

func (s *SQLStore) ListJobs(limit int) []JobRecord {
	query := `SELECT id, document_id, status, stage, progress, attempt_count, max_attempts, batch_id, request_json, result, error, created_at, started_at, completed_at, updated_at FROM jobs ORDER BY created_at DESC`
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}
	rows, err := s.db.Query(s.rebind(query))
	if err != nil {
		return nil
	}
	defer rows.Close()

	var items []JobRecord
	for rows.Next() {
		item, err := scanJob(rows)
		if err == nil {
			items = append(items, item)
		}
	}
	return items
}

func (s *SQLStore) CreateBatchJob(batch BatchJobRecord) error {
	requestsJSON, err := json.Marshal(batch.Requests)
	if err != nil {
		return err
	}
	childIDsJSON, err := json.Marshal(batch.ChildJobIDs)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(s.rebind(`INSERT INTO batch_jobs (
id, status, stage, progress, attempt_count, requests_json, child_job_ids_json, created_at, started_at, completed_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`),
		batch.ID, string(batch.Status), string(batch.Stage), batch.Progress, batch.AttemptCount, string(requestsJSON), string(childIDsJSON),
		batch.CreatedAt, nullableTime(batch.StartedAt), nullableTime(batch.CompletedAt), batch.UpdatedAt,
	)
	return err
}

func (s *SQLStore) UpdateBatchJob(batch BatchJobRecord) error {
	requestsJSON, err := json.Marshal(batch.Requests)
	if err != nil {
		return err
	}
	childIDsJSON, err := json.Marshal(batch.ChildJobIDs)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(s.rebind(`UPDATE batch_jobs SET
status = ?, stage = ?, progress = ?, attempt_count = ?, requests_json = ?, child_job_ids_json = ?, created_at = ?, started_at = ?, completed_at = ?, updated_at = ?
WHERE id = ?`),
		string(batch.Status), string(batch.Stage), batch.Progress, batch.AttemptCount, string(requestsJSON), string(childIDsJSON),
		batch.CreatedAt, nullableTime(batch.StartedAt), nullableTime(batch.CompletedAt), batch.UpdatedAt, batch.ID,
	)
	return err
}

func (s *SQLStore) GetBatchJob(id string) (BatchJobRecord, bool) {
	row := s.db.QueryRow(s.rebind(`SELECT id, status, stage, progress, attempt_count, requests_json, child_job_ids_json, created_at, started_at, completed_at, updated_at FROM batch_jobs WHERE id = ?`), id)
	record, err := scanBatch(row)
	return record, err == nil
}

func (s *SQLStore) SaveDocument(document DocumentRecord) error {
	requestJSON, err := json.Marshal(document.Request)
	if err != nil {
		return err
	}
	sectionTitlesJSON, err := json.Marshal(document.SectionTitles)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(s.rebind(`INSERT INTO documents (
id, job_id, request_json, title, markdown, content, output_format, section_titles_json, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT (id) DO UPDATE SET
job_id = EXCLUDED.job_id, request_json = EXCLUDED.request_json, title = EXCLUDED.title, markdown = EXCLUDED.markdown, content = EXCLUDED.content,
output_format = EXCLUDED.output_format, section_titles_json = EXCLUDED.section_titles_json, created_at = EXCLUDED.created_at, updated_at = EXCLUDED.updated_at`),
		document.ID, document.JobID, string(requestJSON), document.Title, document.Markdown, document.Content, document.OutputFormat, string(sectionTitlesJSON), document.CreatedAt, document.UpdatedAt,
	)
	if s.dialect == "sqlite" && err != nil {
		_, err = s.db.Exec(s.rebind(`INSERT OR REPLACE INTO documents (
id, job_id, request_json, title, markdown, content, output_format, section_titles_json, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`),
			document.ID, document.JobID, string(requestJSON), document.Title, document.Markdown, document.Content, document.OutputFormat, string(sectionTitlesJSON), document.CreatedAt, document.UpdatedAt,
		)
	}
	return err
}

func (s *SQLStore) GetDocument(id string) (DocumentRecord, bool) {
	row := s.db.QueryRow(s.rebind(`SELECT id, job_id, request_json, title, markdown, content, output_format, section_titles_json, created_at, updated_at FROM documents WHERE id = ?`), id)
	record, err := scanDocument(row)
	return record, err == nil
}

func (s *SQLStore) GetDocumentByJob(jobID string) (DocumentRecord, bool) {
	row := s.db.QueryRow(s.rebind(`SELECT id, job_id, request_json, title, markdown, content, output_format, section_titles_json, created_at, updated_at FROM documents WHERE job_id = ?`), jobID)
	record, err := scanDocument(row)
	return record, err == nil
}

func (s *SQLStore) ListDocuments(limit int) []DocumentRecord {
	query := `SELECT id, job_id, request_json, title, markdown, content, output_format, section_titles_json, created_at, updated_at FROM documents ORDER BY created_at DESC`
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}
	rows, err := s.db.Query(s.rebind(query))
	if err != nil {
		return nil
	}
	defer rows.Close()

	var items []DocumentRecord
	for rows.Next() {
		item, err := scanDocument(rows)
		if err == nil {
			items = append(items, item)
		}
	}
	return items
}

func (s *SQLStore) SaveTrace(jobID string, trace PipelineTrace) error {
	payload, err := json.Marshal(trace)
	if err != nil {
		return err
	}
	query := s.rebind(`INSERT INTO traces (job_id, trace_json, updated_at) VALUES (?, ?, ?)
ON CONFLICT (job_id) DO UPDATE SET trace_json = EXCLUDED.trace_json, updated_at = EXCLUDED.updated_at`)
	_, err = s.db.Exec(query, jobID, string(payload), time.Now().UTC())
	if s.dialect == "sqlite" && err != nil {
		_, err = s.db.Exec(s.rebind(`INSERT OR REPLACE INTO traces (job_id, trace_json, updated_at) VALUES (?, ?, ?)`), jobID, string(payload), time.Now().UTC())
	}
	return err
}

func (s *SQLStore) GetTrace(jobID string) (PipelineTrace, bool) {
	var raw string
	err := s.db.QueryRow(s.rebind(`SELECT trace_json FROM traces WHERE job_id = ?`), jobID).Scan(&raw)
	if err != nil {
		return PipelineTrace{}, false
	}
	var trace PipelineTrace
	if err := json.Unmarshal([]byte(raw), &trace); err != nil {
		return PipelineTrace{}, false
	}
	return trace, true
}

func (s *SQLStore) schema() []string {
	switch s.dialect {
	case "postgres":
		return []string{
			`CREATE TABLE IF NOT EXISTS jobs (
id TEXT PRIMARY KEY,
document_id TEXT NOT NULL,
status TEXT NOT NULL,
stage TEXT NOT NULL,
progress INTEGER NOT NULL,
attempt_count INTEGER NOT NULL,
max_attempts INTEGER NOT NULL,
batch_id TEXT,
request_json TEXT NOT NULL,
result TEXT,
error TEXT,
created_at TIMESTAMPTZ NOT NULL,
started_at TIMESTAMPTZ NULL,
completed_at TIMESTAMPTZ NULL,
updated_at TIMESTAMPTZ NOT NULL
)`,
			`CREATE TABLE IF NOT EXISTS batch_jobs (
id TEXT PRIMARY KEY,
status TEXT NOT NULL,
stage TEXT NOT NULL,
progress INTEGER NOT NULL,
attempt_count INTEGER NOT NULL,
requests_json TEXT NOT NULL,
child_job_ids_json TEXT NOT NULL,
created_at TIMESTAMPTZ NOT NULL,
started_at TIMESTAMPTZ NULL,
completed_at TIMESTAMPTZ NULL,
updated_at TIMESTAMPTZ NOT NULL
)`,
			`CREATE TABLE IF NOT EXISTS documents (
id TEXT PRIMARY KEY,
job_id TEXT NOT NULL,
request_json TEXT NOT NULL,
title TEXT NOT NULL,
markdown TEXT NOT NULL,
content TEXT NOT NULL,
output_format TEXT NOT NULL,
section_titles_json TEXT NOT NULL,
created_at TIMESTAMPTZ NOT NULL,
updated_at TIMESTAMPTZ NOT NULL
)`,
			`CREATE TABLE IF NOT EXISTS traces (
job_id TEXT PRIMARY KEY,
trace_json TEXT NOT NULL,
updated_at TIMESTAMPTZ NOT NULL
)`,
		}
	default:
		return []string{
			`CREATE TABLE IF NOT EXISTS jobs (
id TEXT PRIMARY KEY,
document_id TEXT NOT NULL,
status TEXT NOT NULL,
stage TEXT NOT NULL,
progress INTEGER NOT NULL,
attempt_count INTEGER NOT NULL,
max_attempts INTEGER NOT NULL,
batch_id TEXT,
request_json TEXT NOT NULL,
result TEXT,
error TEXT,
created_at TIMESTAMP NOT NULL,
started_at TIMESTAMP NULL,
completed_at TIMESTAMP NULL,
updated_at TIMESTAMP NOT NULL
)`,
			`CREATE TABLE IF NOT EXISTS batch_jobs (
id TEXT PRIMARY KEY,
status TEXT NOT NULL,
stage TEXT NOT NULL,
progress INTEGER NOT NULL,
attempt_count INTEGER NOT NULL,
requests_json TEXT NOT NULL,
child_job_ids_json TEXT NOT NULL,
created_at TIMESTAMP NOT NULL,
started_at TIMESTAMP NULL,
completed_at TIMESTAMP NULL,
updated_at TIMESTAMP NOT NULL
)`,
			`CREATE TABLE IF NOT EXISTS documents (
id TEXT PRIMARY KEY,
job_id TEXT NOT NULL,
request_json TEXT NOT NULL,
title TEXT NOT NULL,
markdown TEXT NOT NULL,
content TEXT NOT NULL,
output_format TEXT NOT NULL,
section_titles_json TEXT NOT NULL,
created_at TIMESTAMP NOT NULL,
updated_at TIMESTAMP NOT NULL
)`,
			`CREATE TABLE IF NOT EXISTS traces (
job_id TEXT PRIMARY KEY,
trace_json TEXT NOT NULL,
updated_at TIMESTAMP NOT NULL
)`,
		}
	}
}

func (s *SQLStore) rebind(query string) string {
	if s.dialect != "postgres" {
		return query
	}
	var b strings.Builder
	index := 1
	for _, r := range query {
		if r == '?' {
			b.WriteString(fmt.Sprintf("$%d", index))
			index++
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanJob(scanner rowScanner) (JobRecord, error) {
	var (
		record      JobRecord
		status      string
		stage       string
		batchID     sql.NullString
		requestJSON string
		result      sql.NullString
		errMsg      sql.NullString
		startedAt   sql.NullTime
		completedAt sql.NullTime
	)
	if err := scanner.Scan(
		&record.ID, &record.DocumentID, &status, &stage, &record.Progress, &record.AttemptCount, &record.MaxAttempts,
		&batchID, &requestJSON, &result, &errMsg, &record.CreatedAt, &startedAt, &completedAt, &record.UpdatedAt,
	); err != nil {
		return JobRecord{}, err
	}
	if err := json.Unmarshal([]byte(requestJSON), &record.Request); err != nil {
		return JobRecord{}, err
	}
	record.Status = JobStatus(status)
	record.Stage = JobStage(stage)
	record.BatchID = batchID.String
	record.Result = result.String
	record.Error = errMsg.String
	record.StartedAt = ptrTime(startedAt)
	record.CompletedAt = ptrTime(completedAt)
	return record, nil
}

func scanBatch(scanner rowScanner) (BatchJobRecord, error) {
	var (
		record          BatchJobRecord
		status          string
		stage           string
		requestsJSON    string
		childJobIDsJSON string
		startedAt       sql.NullTime
		completedAt     sql.NullTime
	)
	if err := scanner.Scan(
		&record.ID, &status, &stage, &record.Progress, &record.AttemptCount, &requestsJSON, &childJobIDsJSON,
		&record.CreatedAt, &startedAt, &completedAt, &record.UpdatedAt,
	); err != nil {
		return BatchJobRecord{}, err
	}
	if err := json.Unmarshal([]byte(requestsJSON), &record.Requests); err != nil {
		return BatchJobRecord{}, err
	}
	if err := json.Unmarshal([]byte(childJobIDsJSON), &record.ChildJobIDs); err != nil {
		return BatchJobRecord{}, err
	}
	record.Status = JobStatus(status)
	record.Stage = JobStage(stage)
	record.StartedAt = ptrTime(startedAt)
	record.CompletedAt = ptrTime(completedAt)
	return record, nil
}

func scanDocument(scanner rowScanner) (DocumentRecord, error) {
	var (
		record            DocumentRecord
		requestJSON       string
		sectionTitlesJSON string
	)
	if err := scanner.Scan(
		&record.ID, &record.JobID, &requestJSON, &record.Title, &record.Markdown, &record.Content, &record.OutputFormat, &sectionTitlesJSON,
		&record.CreatedAt, &record.UpdatedAt,
	); err != nil {
		return DocumentRecord{}, err
	}
	if err := json.Unmarshal([]byte(requestJSON), &record.Request); err != nil {
		return DocumentRecord{}, err
	}
	if err := json.Unmarshal([]byte(sectionTitlesJSON), &record.SectionTitles); err != nil {
		return DocumentRecord{}, err
	}
	return record, nil
}

func nullableString(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}

func nullableTime(value *time.Time) any {
	if value == nil || value.IsZero() {
		return nil
	}
	return value.UTC()
}

func ptrTime(value sql.NullTime) *time.Time {
	if !value.Valid {
		return nil
	}
	t := value.Time.UTC()
	return &t
}
