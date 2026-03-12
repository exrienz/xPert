Below is a **complete production blueprint adapted to your requirements**:

* **Multi-LLM routing (OpenAI compatible scheme)**
* **Dynamic document size**
* **Interactive + batch generation**
* **Persistent storage**
* **Parallel generation control**
* **Monorepo**
* **Single container deployment**
* **Edge-device optimized**

The design prioritizes:

1. correctness
2. security
3. performance
4. operational simplicity
5. scalability

---

# 1. Assumptions

Based on your requirements:

**Runtime**

```
single binary
single container
edge hardware capable
```

**Hardware targets**

```
2–8 CPU
4–16GB RAM
```

**LLM providers**

OpenAI-compatible APIs:

```
OpenAI
OpenRouter
DeepSeek
Local Ollama
Custom endpoints
```

**Workload**

```
30k–100k+ word documents
parallel subsection generation
```

**Architecture**

```
modular monolith
worker engine inside same process
queue abstraction
```

---

# 2. System Architecture Overview

The application runs as **one Go process** containing:

```
HTTP API
Worker Scheduler
Agent Engine
Planning Engine
Review Engine
Synthesis Engine
Formatting Engine
LLM Router
Storage
Queue
```

The system internally uses **in-memory scheduling + optional Redis fallback**.

Execution pipeline:

```
User Request
↓
Document Job Created
↓
Master Planner
↓
Section Planners
↓
Agent Generator
↓
Expert Agents
↓
Section Reviewers
↓
Section Synthesizer
↓
Gap Detector
↓
Global Synthesizer
↓
Formatter
↓
Stored Document
```

---

# 3. Architecture Diagram

```
                 User
                   │
                   ▼
              HTTP API
                   │
                   ▼
             Job Manager
                   │
           (internal queue)
                   │
 ┌─────────────────┼─────────────────┐
 ▼                 ▼                 ▼
Planner Engine   Agent Engine    Review Engine
 │                 │                 │
 ▼                 ▼                 ▼
Section Plans   Expert Agents   Section Reviews
 │                 │                 │
 └───────────────┬─┴─┬───────────────┘
                 ▼   ▼
           Synthesis Engine
                 │
                 ▼
           Formatter Engine
                 │
                 ▼
             PostgreSQL
                 │
                 ▼
           Final Documents
```

Everything runs inside **one binary**.

---

# 4. Repository Structure

Monorepo designed for modular extraction later.

```
xpert/

cmd/
 └── server/
      main.go

internal/

  api/
    router.go
    handlers.go
    middleware.go

  orchestrator/
    pipeline.go
    job_manager.go
    scheduler.go

  planner/
    master_planner.go
    section_planner.go
    models.go

  agents/
    agent_factory.go
    expert_agent.go
    agent_types.go

  review/
    reviewer.go
    gap_detector.go

  synthesis/
    section_synth.go
    global_synth.go

  formatter/
    markdown.go
    html.go
    pdf.go

  llm/
    router.go
    client.go
    openai_client.go
    ollama_client.go

  context/
    compressor.go
    summarizer.go

  queue/
    memory_queue.go
    job_types.go

  storage/
    postgres.go
    repository.go
    models.go

  concurrency/
    worker_pool.go
    rate_limiter.go

pkg/
  logger/
  config/

migrations/

docker/
  Dockerfile

docker-compose.yml
```

---

# 5. Core Components

## 5.1 API Layer

API supports **interactive and batch workflows**.

Routes:

```
POST /documents
POST /documents/batch
GET  /documents/{id}
GET  /jobs/{id}
DELETE /jobs/{id}
```

Example request:

```
POST /documents

{
  "topic": "Web Penetration Testing SOP",
  "document_type": "SOP",
  "complexity": "auto",
  "format": "markdown"
}
```

---

# 5.2 Job Manager

Handles execution lifecycle.

States:

```
pending
planning
generating
reviewing
synthesizing
formatting
completed
failed
```

Job struct:

```go
type Job struct {
	ID          string
	DocumentID  string
	Status      string
	Stage       string
	Progress    int
	CreatedAt   time.Time
}
```

---

# 5.3 Planner Engine

### Master Planner

Creates top-level sections.

Example output:

```
Governance
Preparation
Reconnaissance
Vulnerability Discovery
Exploitation
Post Exploitation
Reporting
Quality Assurance
Deliverables
```

### Section Planner

Expands sections.

Example:

```
Reconnaissance
  Passive Recon
  Subdomain Enumeration
  Technology Fingerprinting
  Endpoint Discovery
```

Planner interface:

```go
type Planner interface {
	Plan(ctx context.Context, doc *Document) ([]Section, error)
}
```

---

# 5.4 Agent Engine

Agents generate subsection content.

Agent struct:

```go
type Agent struct {
	ID          string
	Name        string
	Expertise   string
	SectionID   string
	Prompt      string
	MaxTokens   int
}
```

Agents operate **in parallel worker pools**.

Typical generation:

```
500–2000 words per subsection
```

---

# 5.5 Review Engine

Reviewers validate section quality.

Checks:

```
accuracy
missing concepts
step ordering
tool examples
terminology
```

Reviewer interface:

```go
type Reviewer interface {
	Review(ctx context.Context, section *Section) (*ReviewResult, error)
}
```

---

# 5.6 Gap Detector

Global semantic check.

Finds:

```
duplicate sections
missing references
concept ordering errors
terminology mismatch
```

Example:

```
Exploitation references SQL injection before vulnerability classification exists.
```

---

# 5.7 Synthesis Engine

Two phases.

### Section synthesis

Merge subsections into chapters.

### Global synthesis

Merge chapters.

Responsibilities:

```
remove duplication
normalize terminology
improve transitions
```

---

# 5.8 Formatter Engine

Converts internal markdown to formats:

```
markdown
html
pdf
notion export
confluence export
```

Markdown becomes canonical storage format.

---

# 6. Data Model Strategy

Database: **PostgreSQL**

Reasons:

```
structured relationships
document persistence
transaction safety
high reliability
```

---

### documents

```
id
topic
document_type
status
created_at
updated_at
```

---

### sections

```
id
document_id
title
summary
order_index
content
```

---

### subsections

```
id
section_id
title
content
agent_id
```

---

### agents

```
id
name
expertise
prompt_template
```

---

### jobs

```
id
document_id
status
stage
progress
```

---

# 7. LLM Routing System

Multi-provider routing.

Supported endpoints:

```
OpenAI
OpenRouter
DeepSeek
Ollama
Custom
```

Router interface:

```go
type LLMClient interface {
	Generate(ctx context.Context, req Request) (*Response, error)
}
```

Router:

```go
type Router struct {
	providers map[string]LLMClient
}
```

Example config:

```
provider=openai
model=gpt-4o

provider=deepseek
model=deepseek-chat

provider=ollama
model=llama3
```

---

# 8. Concurrency Strategy

Parallelism is critical.

Worker pools:

```
planner workers
agent workers
review workers
synthesis workers
```

Example worker pool:

```go
type WorkerPool struct {
	jobs chan Job
	wg   sync.WaitGroup
}
```

---

### Parallel execution

Example:

```
8 sections
→ 8 planners

35 subsections
→ 35 agents
```

All run concurrently.

---

### Rate limiting

Each provider has limits.

Limiter:

```
requests/sec
tokens/sec
max concurrency
```

Example:

```
OpenAI: 10 requests/sec
Ollama: 2 concurrent
```

---

# 9. Context Management

Critical for large documents.

Agents receive **compressed context**:

```
global summary
section summary
terminology dictionary
previous section outputs
```

Context compressor reduces token usage.

Example:

```
full doc → 80k tokens
compressed → 2k tokens
```

---

# 10. Deployment Layout

Single container deployment.

```
services:

  xpert:
    build: .
    ports:
      - "8080:8080"
    environment:
      - DATABASE_URL=postgres://...
      - OPENAI_API_KEY=...
```

Optional external services:

```
PostgreSQL
Redis
```

But **system works without Redis**.

---

# 11. Dockerfile

Minimal image.

```
FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY . .
RUN go build -o xpert ./cmd/server

FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/xpert .

EXPOSE 8080

CMD ["./xpert"]
```

Image size ≈ **30–50MB**.

---

# 12. Performance Strategy

### Parallel agents

Up to:

```
CPU * 4 workers
```

Example:

```
8 CPU → 32 concurrent agents
```

---

### Streaming synthesis

Sections merged incrementally.

---

### Token budgeting

Agents limited:

```
1500–2500 tokens
```

---

### Memory control

Large documents stored as **section fragments**.

Not entire text in RAM.

---

# 13. Security Strategy

### API authentication

```
API keys
JWT tokens
```

---

### Prompt safety

Prevent injection:

```
strip tool instructions
validate system prompts
limit context injection
```

---

### Secret management

Use:

```
.env
docker secrets
```

---

### Network exposure

Expose only:

```
API port
```

Block internal services.

---

# 14. Scaling Path

Although monolithic, system evolves easily.

Future split:

```
planner-service
agent-service
review-service
synth-service
```

Queue upgraded to:

```
NATS
Kafka
```

GPU cluster for local LLMs.

---

# 15. Example Pipeline Execution

```
User Request
↓
Generate Security Playbook
↓
Master Planner → 9 sections
↓
Section Planners → 34 subsections
↓
34 Expert Agents run parallel
↓
Section Reviewers
↓
Section Synthesizers
↓
Gap Detector
↓
Global Synthesizer
↓
Formatter
↓
80k word document produced
```

---

# 16. Expected System Size

Approximate implementation size.

```
Core MVP
8k–12k LOC
```

Production system:

```
20k–30k LOC
```

---

# 17. Recommended Next Step

The **most powerful improvement** for this architecture is replacing the linear pipeline with a **DAG execution engine**.

Instead of:

```
Planner → Agents → Review → Synth
```

You build:

```
dynamic execution graph
```

This allows:

```
conditional agent creation
dynamic retries
adaptive planning
self-healing workflows
```

