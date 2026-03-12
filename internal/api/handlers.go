package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"docgen/internal/storage"
)

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(`<!DOCTYPE html><html><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><title>DocGen</title><style>body{font-family:system-ui,sans-serif;background:#f5f1e8;color:#1f2933;max-width:960px;margin:0 auto;padding:40px 20px;line-height:1.5}code,pre{background:#fff;border:1px solid #d5d9e0;border-radius:12px}code{padding:2px 6px}pre{padding:16px;overflow:auto}</style></head><body><h1>DocGen</h1><p>Go is the only active runtime in this repository.</p><p>Storage backend: <code>` + s.config.StorageBackend + `</code></p><p>Available endpoints:</p><pre>POST   /documents
POST   /documents/batch
GET    /documents
GET    /documents/{id}
GET    /jobs
GET    /jobs/{id}
DELETE /jobs/{id}</pre><p>LLM provider: <code>` + s.config.LLMProvider + `</code></p></body></html>`))
}

func (s *Server) handleDocuments(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var payload storage.DocumentRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
			return
		}
		job, err := s.jobManager.CreateJob(payload)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusAccepted, storage.CreateJobResponse{ID: job.ID, Status: job.Status, DocumentID: job.DocumentID})
	case http.MethodGet:
		writeJSON(w, http.StatusOK, map[string]any{"documents": s.repository.ListDocuments(50)})
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleBatchDocuments(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var payload storage.BatchDocumentRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	batch, err := s.jobManager.CreateBatchJob(payload)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]any{
		"id":            batch.ID,
		"status":        batch.Status,
		"child_job_ids": batch.ChildJobIDs,
	})
}

func (s *Server) handleDocumentByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/documents/")
	if id == "" || id == "/documents" {
		http.NotFound(w, r)
		return
	}
	if detail, ok := s.jobManager.GetDocumentDetail(id); ok {
		writeJSON(w, http.StatusOK, detail)
		return
	}
	if detail, ok := s.jobManager.GetJobDetail(id); ok {
		writeJSON(w, http.StatusOK, detail)
		return
	}
	writeJSON(w, http.StatusNotFound, map[string]string{"error": "document not found"})
}

func (s *Server) handleJobs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"jobs": s.repository.ListJobs(50)})
}

func (s *Server) handleJobByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/jobs/")
	if id == "" {
		http.NotFound(w, r)
		return
	}

	switch r.Method {
	case http.MethodGet:
		if detail, ok := s.jobManager.GetJobDetail(id); ok {
			writeJSON(w, http.StatusOK, detail)
			return
		}
		if detail, ok := s.jobManager.GetBatchDetail(id); ok {
			writeJSON(w, http.StatusOK, detail)
			return
		}
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "job not found"})
	case http.MethodDelete:
		if _, ok := s.jobManager.CancelJob(id); !ok {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "job not found"})
			return
		}
		w.WriteHeader(http.StatusAccepted)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
