package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"

	"docgen/internal/agents"
	"docgen/internal/api"
	"docgen/internal/config"
	"docgen/internal/formatter"
	"docgen/internal/llm"
	"docgen/internal/orchestrator"
	"docgen/internal/planner"
	"docgen/internal/queue"
	"docgen/internal/review"
	"docgen/internal/storage"
	"docgen/internal/structure"
	"docgen/internal/synthesis"
)

func main() {
	cfg := config.Load()

	repo := newStore(cfg)
	if err := repo.Load(); err != nil {
		log.Fatalf("load repository: %v", err)
	}

	workQueue := queue.NewMemoryQueue(cfg.MaxParallelSections * 4)
	pipeline := orchestrator.NewPipeline(
		planner.NewIntentDetector(),
		planner.NewDocumentClassifier(),
		planner.NewMasterPlanner(),
		planner.NewSectionPlanner(),
		agents.NewExpertAgent(llm.NewRouter(cfg)),
		review.NewReviewer(),
		review.NewGapDetector(),
		synthesis.NewSectionSynthesizer(),
		synthesis.NewGlobalSynthesizer(),
		structure.NewDocumentStructurer(),
		formatter.NewFormatterSet(),
		cfg.MaxParallelSections,
	)
	jobManager := orchestrator.NewJobManager(repo, workQueue, pipeline, cfg.MaxJobAttempts)
	scheduler := orchestrator.NewScheduler(workQueue, jobManager)
	scheduler.Start()
	defer scheduler.Stop()

	handler := api.NewServer(cfg, repo, jobManager)
	httpServer := &http.Server{
		Addr:    cfg.Address(),
		Handler: handler.Routes(),
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		<-ctx.Done()
		_ = httpServer.Shutdown(context.Background())
	}()

	log.Printf("docgen listening on %s", cfg.Address())
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("listen: %v", err)
	}
}

func newStore(cfg config.Config) storage.Store {
	switch cfg.StorageBackend {
	case "postgres":
		return storage.NewPostgresRepository(firstNonEmpty(cfg.StorageDSN, cfg.DataPath))
	case "file":
		return storage.NewRepository(cfg.DataPath)
	default:
		return storage.NewSQLiteRepository(firstNonEmpty(cfg.StorageDSN, cfg.DataPath))
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
