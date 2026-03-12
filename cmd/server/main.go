package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"

	"xpert/internal/agents"
	"xpert/internal/api"
	"xpert/internal/config"
	"xpert/internal/formatter"
	"xpert/internal/llm"
	"xpert/internal/orchestrator"
	"xpert/internal/planner"
	"xpert/internal/queue"
	"xpert/internal/review"
	"xpert/internal/storage"
	"xpert/internal/structure"
	"xpert/internal/synthesis"
)

func main() {
	cfg := config.Load()

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		log.Fatalf("configuration error: %v", err)
	}

	// Log model configuration
	if cfg.EnableRandomModelSelection {
		log.Printf("random model selection enabled with pool: %v", cfg.AIModelPool)
	} else {
		log.Printf("using single model: %s", cfg.OpenAIModel)
	}

	repo := newStore(cfg)
	if err := repo.Load(); err != nil {
		log.Fatalf("load repository: %v", err)
	}

	workQueue := queue.NewMemoryQueue(cfg.MaxParallelSections * 4)
	router := llm.NewRouter(cfg)
	pipeline := orchestrator.NewPipeline(
		planner.NewIntentDetector(),
		planner.NewDocumentClassifier(),
		planner.NewMasterPlanner(),
		planner.NewSectionPlanner(),
		agents.NewAgentFactory(router),
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

	log.Printf("xPert listening on %s", cfg.Address())
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
