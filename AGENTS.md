# AGENTS.md

## Commands

```bash
go run ./cmd/server            # Run server (0.0.0.0:8080)
go test ./...                  # All tests
go test ./internal/llm/...     # Single package
go build -o xpert ./cmd/server # Build binary
```

No linter, typechecker, or CI config exists in this repo. No `make` targets.

## Architecture

- Single entrypoint: `cmd/server/main.go` — wires all packages by hand, no DI framework
- Uses std `net/http` only, no HTTP framework
- Pure Go SQLite (`modernc.org/sqlite`) — **no CGO required**
- In-memory job queue (`internal/queue/`), no external broker needed
- Default LLM provider is `mock` (returns prompts as-is) — no API key needed for local dev/test

### Package roles

| Package | Purpose |
|---|---|
| `internal/config` | Env-var config loading |
| `internal/api` | HTTP handlers, router, middleware |
| `internal/orchestrator` | Pipeline, job manager, scheduler |
| `internal/planner` | Intent detection, doc classification, master & section planning |
| `internal/agents` | LLM-powered content generation (parallel per section) |
| `internal/review` | Content review and gap detection |
| `internal/synthesis` | Section + global synthesis |
| `internal/structure` | Document structuring |
| `internal/formatter` | Output: markdown, HTML, PDF |
| `internal/llm` | Provider routing, model pool, client factory |
| `internal/queue` | In-memory job queue |
| `internal/workerpool` | Generic worker pool |
| `internal/storage` | sqlite / postgres / file backends |
| `internal/context` | Context management |

### Pipeline flow

`orchestrator/pipeline.go` chains: intent → classify → master plan → section plan → agent generate (parallel) → review → gap detect → synthesize → structure → format

## Gotchas

- `XPERT_AI_MODEL_POOL` must be set when `XPERT_ENABLE_RANDOM_MODEL_SELECTION=true` or `config.Validate()` fails
- SQLite data dir (`./data/`) must exist before startup or the server fails
- Queue capacity is `MaxParallelSections * 4`, not configurable independently
- `storage.NewRepository()` (file backend) takes the same `DataPath` param as SQLite — it stores JSON, not a DB file