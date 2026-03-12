# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

xPert is a Go modular monolith that generates documents using LLM-powered pipelines. It handles planning, generation, review, synthesis, queueing, and storage in a single process.

## Commands

```bash
# Run the server (default: 0.0.0.0:8080)
go run ./cmd/server

# Run tests
go test ./...

# Run a specific package's tests
go test ./internal/llm/...

# Build binary
go build -o xpert ./cmd/server

# Docker
docker compose up --build
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `XPERT_LLM_PROVIDER` | `mock` | LLM provider: `mock`, `openai`, `openrouter`, `deepseek`, `ollama`, `custom` |
| `XPERT_OPENAI_API_KEY` | - | API key for OpenAI-compatible providers |
| `XPERT_OPENAI_BASE_URL` | `https://api.openai.com/v1` | Base URL for LLM API |
| `XPERT_OPENAI_MODEL` | `gpt-4o-mini` | Model identifier (single model mode) |
| `XPERT_AI_MODEL_POOL` | - | Comma-separated model pool (e.g., `gpt-4o,gpt-4o-mini,claude-3`) |
| `XPERT_ENABLE_RANDOM_MODEL_SELECTION` | `false` | Enable random model selection from pool |
| `XPERT_STORAGE_BACKEND` | `sqlite` | Storage: `sqlite`, `postgres`, `file` |
| `XPERT_DATA_PATH` | `./data/xpert.sqlite` | Database file path |
| `XPERT_MAX_PARALLEL_SECTIONS` | `4` | Concurrent section generation |

**Model Pool**: When `XPERT_ENABLE_RANDOM_MODEL_SELECTION=true` and `XPERT_AI_MODEL_POOL` is set, the system randomly selects a model for each LLM request. If a model fails, it automatically retries with another model from the pool.

## Architecture

### Document Generation Pipeline

The core flow in `internal/orchestrator/pipeline.go`:

1. **Intent Detection** (`internal/planner/intent_detector.go`) - Analyzes user prompt
2. **Document Classification** (`internal/planner/document_classifier.go`) - Determines document type and focus terms
3. **Master Planning** (`internal/planner/master_planner.go`) - Creates high-level section plans
4. **Section Planning** (`internal/planner/section_planner.go`) - Expands each section with subsections
5. **Expert Agent** (`internal/agents/expert_agent.go`) - Generates content via LLM, runs in parallel per section
6. **Review** (`internal/review/reviewer.go`) - Reviews generated content, triggers revisions if needed
7. **Gap Detection** (`internal/review/gap_detector.go`) - Identifies content gaps
8. **Synthesis** (`internal/synthesis/`) - Combines sections into final document
9. **Formatting** (`internal/formatter/`) - Outputs as markdown, HTML, or PDF

### Job Processing

- `internal/queue/memory_queue.go` - In-memory job queue
- `internal/orchestrator/scheduler.go` - Polls queue and dispatches to job manager
- `internal/orchestrator/job_manager.go` - Manages job lifecycle and persistence
- Jobs have stages: `pending` → `planning` → `generating` → `reviewing` → `synthesizing` → `formatting` → `completed`

### LLM Routing

`internal/llm/router.go` selects the client based on `XPERT_LLM_PROVIDER`:
- `mock` - Returns prompts as-is (for local dev)
- `openai` / `openrouter` / `deepseek` / `custom` - OpenAI-compatible API
- `ollama` - Local Ollama instance

**Model Pool Support** (`internal/llm/model_selector.go`, `internal/llm/client_factory.go`):
- Random model selection from configurable pool
- Automatic fallback on model failure
- Per-request model selection with usage tracking

### Storage

Three backends in `internal/storage/`:
- `sqlite` (default) - SQLite via modernc.org/sqlite (pure Go, no CGO)
- `postgres` - PostgreSQL for production
- `file` - JSON file-based repository

## API Endpoints

```
POST   /documents       - Create document job
POST   /documents/batch - Create batch of document jobs
GET    /documents       - List documents
GET    /documents/{id}  - Get document by ID
GET    /jobs            - List jobs
GET    /jobs/{id}       - Get job details
DELETE /jobs/{id}       - Cancel job
```

## Key Types

- `storage.DocumentRequest` - Input for document generation (prompt, type, tone, format)
- `storage.JobRecord` - Job state and metadata
- `storage.DocumentRecord` - Generated document output
- `storage.PipelineTrace` - Debug info about generation steps
