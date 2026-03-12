package api

import (
	"net/http"

	"docgen/internal/config"
	"docgen/internal/orchestrator"
	"docgen/internal/storage"
)

type Server struct {
	config     config.Config
	repository storage.Store
	jobManager *orchestrator.JobManager
}

func NewServer(cfg config.Config, repository storage.Store, jobManager *orchestrator.JobManager) *Server {
	return &Server{
		config:     cfg,
		repository: repository,
		jobManager: jobManager,
	}
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/documents", s.handleDocuments)
	mux.HandleFunc("/documents/", s.handleDocumentByID)
	mux.HandleFunc("/documents/batch", s.handleBatchDocuments)
	mux.HandleFunc("/jobs", s.handleJobs)
	mux.HandleFunc("/jobs/", s.handleJobByID)
	return recoveryMiddleware(mux)
}
