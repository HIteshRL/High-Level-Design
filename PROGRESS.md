# Roognis — Walking Skeleton Progress

> **Last updated:** 13 February 2026
> **Status:** Walking skeleton complete — local Ollama inference enabled via Docker Compose
> **Build/Test:** `go build ✅` | `go vet ✅` | `go test ✅`

---

## What is Roognis?

An educational AI platform with a modular backend architecture supporting text/image/video inference, RAG pipelines, psychographic intelligence, multi-persona access, and more. The full target architecture is documented in `design.swift` (Eraser.io D2 diagram, 280+ lines).

This walking skeleton implements the **text inference pipeline** end-to-end as the first vertical slice.

---

## Architecture Implemented

```
External Clients
    → Rate Limiter (Redis sliding window / in-memory fallback)
    → CORS Middleware
    → API Gateway (Go net/http, pattern-based routing)
    → JWT Auth Middleware
    → Request Router
    → Prompt Orchestrator
        → Cache Check (Redis)
        → Context Injector (RAG stub)
        → LLM Inference Node (OpenAI-compatible HTTP client)
        → Token Streamer (SSE)
    → Response → Client
```

**Cross-cutting:** Structured logging (slog/JSON), health checks (DB + Redis), interaction logger (Postgres), graceful shutdown.

---

## Tech Stack

| Component | Technology |
|-----------|-----------|
| Language | Go 1.24 (standard library `net/http`, no framework) |
| Database | PostgreSQL 16 + pgvector (`pgx/v5` driver, connection pool) |
| Cache | Redis 7 (`go-redis/v9`, sliding-window rate limiter + response cache) |
| Auth | JWT HS256 (`golang-jwt/jwt/v5`), bcrypt (`x/crypto/bcrypt`) |
| LLM | Direct HTTP client to OpenAI-compatible API (no SDK), configured for local Ollama |
| Streaming | Server-Sent Events (SSE) |
| Container | Multi-stage Docker build (alpine), Docker Compose (Postgres + Redis + Ollama) |
| IDs | UUID v4 (`google/uuid`) |

---

## Project Structure

```
roognis/
├── cmd/server/main.go              # Entry point — wires all components
├── internal/
│   ├── config/config.go             # Env-based config with validation
│   ├── db/
│   │   ├── db.go                    # pgx pool + embedded migration runner
│   │   ├── queries.go               # Hand-written SQL (users, conversations, messages)
│   │   └── migrations/
│   │       ├── 000001_init.up.sql   # Schema: users, conversations, messages + pgvector
│   │       └── 000001_init.down.sql # Teardown
│   ├── handler/
│   │   ├── auth.go                  # POST register, POST login, GET me
│   │   ├── health.go                # GET /health (DB + Redis probe)
│   │   └── inference.go             # POST complete + GET conversations
│   ├── middleware/
│   │   ├── auth.go                  # JWT validation, user context injection, role-gating helper
│   │   ├── cors.go                  # Origin allowlist + preflight
│   │   ├── logger.go                # Request logging (method, path, status, duration)
│   │   └── ratelimiter.go           # Redis sorted-set sliding window + memory fallback
│   ├── models/models.go             # Domain types + API contracts + LLM types
│   └── service/
│       ├── auth.go                  # Register (forced student role), authenticate, bcrypt
│       ├── cache.go                 # Redis get/set/JSON, semantic hash (user-scoped)
│       ├── context.go               # RAG stub — prepends system prompt
│       ├── llm.go                   # OpenAI HTTP client (streaming + non-streaming)
│       ├── orchestrator.go          # Pipeline conductor (cache→RAG→LLM→persist)
│       └── sse.go                   # SSE writer with error propagation
├── docker-compose.yml               # Postgres (pgvector:pg16) + Redis (7-alpine) + Ollama
├── Dockerfile                       # Multi-stage build (golang:1.24 → alpine:3.19)
├── Makefile                         # build, run, dev, test, lint, docker-up/down, migrate
├── .env.example                     # All config vars with defaults
├── .gitignore
├── design.swift                     # Full target architecture (D2 diagram)
├── go.mod
└── go.sum
```

**21 Go files** | 2 SQL migrations | 5 direct dependencies | 0 external frameworks

---

## API Endpoints

### Public

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/health` | Health check — reports DB and Redis status |
| `POST` | `/api/v1/auth/register` | Create account (always assigns `student` role) |
| `POST` | `/api/v1/auth/token` | Login — returns JWT |

### Protected (Bearer token required)

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/v1/auth/me` | Current user profile |
| `POST` | `/api/v1/inference/complete` | Text inference (streaming SSE or JSON) |
| `GET` | `/api/v1/conversations` | List user conversations (latest first, max 50) |

### Middleware Chain

```
Incoming Request → Logger → CORS → Rate Limiter → Router
                                                      ├── Public routes (no auth)
                                                      └── /api/v1/* → JWT Auth → Protected routes
```

---

## Database Schema

Three tables with pgvector extension:

- **users** — id, username (unique), email (unique), hashed_password, full_name, role (enum), is_active, timestamps
- **conversations** — id, user_id (FK → users), title, timestamps
- **messages** — id, conversation_id (FK → conversations), role (enum), content, token_count, model_used, latency_ms, embedding (vector(1536)), timestamps

Auto-updated `updated_at` triggers on users and conversations.

---

## Security Hardening Applied

After a devil's advocate review, the following issues were identified and fixed:

### Critical (Fixed)
- **JWT secret validation** — Server panics on startup in non-dev mode if JWT_SECRET is default or < 32 chars
- **Role escalation blocked** — Registration always assigns `student` role; admin/teacher roles require separate admin flow
- **IDOR protection** — Explicit ownership check on conversation access; returns auth error instead of silent fall-through

### High (Fixed)
- **Credential logging removed** — Database DSN and Redis URL no longer logged in plaintext
- **Rate limiter hardened** — Uses `RemoteAddr` only; no blind trust of `X-Forwarded-For`

### Medium (Fixed)
- **Cache scoped per user** — Cache key includes userID + model + prompt + temperature + maxTokens (no cross-user or cross-parameter collisions)
- **SSE write errors propagated** — Client disconnects terminate the streaming pipeline
- **LLM error body bounded** — Error responses capped at 64 KB via `io.LimitReader`
- **Scanner buffer increased** — 1 MB max line size for long SSE chunks
- **Docker hardened** — Ports bound to `127.0.0.1`, Redis password required
- **Conversation handling** — Explicit not-found/forbidden behavior for `conversation_id`; no silent fall-through

### Low (Fixed)
- **Dotenv quote stripping** — Handles `KEY="value"` and `KEY='value'`
- **Prompt length limit** — 32K Unicode character cap enforced server-side

---

## How to Run

```bash
# 1. Start infrastructure
make docker-up

# 2. Copy environment config
cp .env.example .env
# Defaults already target local Ollama:
#   LLM_API_BASE=http://localhost:11434/v1
#   LLM_MODEL=qwen2.5:0.5b
#   LLM_API_KEY=ollama-local

# 3. Run the server
make dev
# or
make build && make run

# 4. Test health
curl http://localhost:8080/health

# 4.1 Verify Ollama model availability
curl http://localhost:11434/api/tags

# 5. Register a user
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"username":"alice","email":"alice@example.com","password":"securepass123"}'

# 6. Get a completion
TOKEN="<access_token from step 5>"
curl -X POST http://localhost:8080/api/v1/inference/complete \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"prompt":"Explain photosynthesis","stream":false}'
```

---

## What's NOT Implemented Yet

These are the remaining modules from `design.swift` that are **not** part of the walking skeleton:

| Module | Status |
|--------|--------|
| RAG Pipeline (ETL → Embedding → VectorDB → Retriever → Re-Ranker) | Stub only (system prompt) |
| Psychographic Intelligence Engine | Not started |
| Concept-Based Questioning Pipeline | Not started |
| Image Inference (OCR/diagram analysis) | Not started |
| Video Inference (transcription/analysis) | Not started |
| Data Analysis Pipeline | Not started |
| Multi-Persona Access (parent/teacher dashboards) | Schema ready (role enum), UI not started |
| Kafka / CDC Event Streaming | Not started |
| Observability (OpenTelemetry, Prometheus, Grafana) | Structured logging only |
| gRPC inter-service communication | Not started |
| OAuth/OIDC (external identity providers) | JWT only |
| Unit tests | Added for config validation, cache hash behavior, prompt validation |
| Integration tests | Not started |
| CI/CD pipeline | Not started |

---

## Next Steps (Recommended Order)

1. **Tests** — Expand unit tests for auth/orchestrator handlers and add integration tests with testcontainers
2. **RAG Pipeline** — Embedding generation → pgvector storage → retrieval → context injection
3. **Kafka + CDC** — Event-driven architecture for interaction logging
4. **Observability** — OpenTelemetry traces + Prometheus metrics
5. **Psychographic Intelligence** — Learning style detection from interaction patterns
6. **Multi-Persona Dashboards** — Parent/teacher views with role-gated endpoints
7. **Concept-Based Questioning** — Bloom's taxonomy question generation
8. **Image/Video Inference** — Multimodal pipeline integration
