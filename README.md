# xPert

`xpert` is the target Go application described by `blueprint.md`. It is a modular monolith that keeps planning, generation, review, synthesis, queueing, and storage inside one Go process and one container.

## Current runtime

- Go monorepo runtime under `cmd/` and `internal/`
- Single-binary HTTP API
- Interactive and batch document jobs
- SQLite persistence for UAT and PostgreSQL support for production
- In-process worker scheduler and queue
- OpenAI-compatible LLM routing with a mock provider for local development

## Layout

```text
xpert/
  cmd/server/main.go
  internal/api/
  internal/orchestrator/
  internal/planner/
  internal/agents/
  internal/review/
  internal/synthesis/
  internal/formatter/
  internal/llm/
  internal/context/
  internal/queue/
  internal/storage/
```

## Run

```bash
go run ./cmd/server
```

Default address:

```bash
0.0.0.0:8080
```

## Environment

```bash
# Server
XPERT_HOST=0.0.0.0
XPERT_PORT=8080

# Storage
XPERT_DATA_PATH=./data/xpert.sqlite
XPERT_STORAGE_BACKEND=sqlite
XPERT_STORAGE_DSN=

# Processing
XPERT_MAX_PARALLEL_SECTIONS=4
XPERT_MAX_JOB_ATTEMPTS=2
XPERT_DEFAULT_WORD_COUNT=6000

# LLM Provider
XPERT_LLM_PROVIDER=mock
XPERT_OPENAI_BASE_URL=https://api.openai.com/v1
XPERT_OPENAI_API_KEY=
XPERT_OPENAI_MODEL=gpt-4o-mini

# Model Pool (optional)
XPERT_AI_MODEL_POOL=
XPERT_ENABLE_RANDOM_MODEL_SELECTION=false
```

### Model Pool

When `XPERT_ENABLE_RANDOM_MODEL_SELECTION=true`, the system randomly selects a model from `XPERT_AI_MODEL_POOL` for each LLM request. If a selected model fails, it automatically retries with another model from the pool.

```bash
# Example: Enable model pool with multiple models
XPERT_ENABLE_RANDOM_MODEL_SELECTION=true
XPERT_AI_MODEL_POOL=gpt-4o,gpt-4o-mini,gpt-4-turbo
```

This provides:
- Load distribution across multiple models
- Automatic failover if a model is unavailable
- Model usage tracking in pipeline traces

## API

```text
POST   /documents
POST   /documents/batch
GET    /documents
GET    /documents/{id}
GET    /jobs
GET    /jobs/{id}
DELETE /jobs/{id}
```

## Status

This repository now exposes only the Go runtime. The current implementation covers the single-binary HTTP API, background scheduling, SQLite persistence for UAT, PostgreSQL wiring for production, OpenAI-compatible provider routing, and multiple export formats.

Remaining work for a fuller production build is primarily:

- optional Redis-backed queue fallback
- native PDF rendering and richer downstream export integrations
- deeper review and gap-detection logic
- additional production hardening around provider-specific retries, rate limiting, and observability
