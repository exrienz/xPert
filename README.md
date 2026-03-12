# DocGen

`docgen` is the target Go application described by `blueprint.md`. It is a modular monolith that keeps planning, generation, review, synthesis, queueing, and storage inside one Go process and one container.

## Current runtime

- Go monorepo runtime under `cmd/` and `internal/`
- Single-binary HTTP API
- Interactive and batch document jobs
- SQLite persistence for UAT and PostgreSQL support for production
- In-process worker scheduler and queue
- OpenAI-compatible LLM routing with a mock provider for local development

## Layout

```text
docgen/
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
DOCGEN_HOST=0.0.0.0
DOCGEN_PORT=8080
DOCGEN_DATA_PATH=./data/docgen.sqlite
DOCGEN_STORAGE_BACKEND=sqlite
DOCGEN_STORAGE_DSN=
DOCGEN_MAX_PARALLEL_SECTIONS=4
DOCGEN_MAX_JOB_ATTEMPTS=2
DOCGEN_DEFAULT_WORD_COUNT=6000
DOCGEN_LLM_PROVIDER=mock
DOCGEN_OPENAI_BASE_URL=https://api.openai.com/v1
DOCGEN_OPENAI_API_KEY=
DOCGEN_OPENAI_MODEL=gpt-4o-mini
```

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
# xPert
# xPert
