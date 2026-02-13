# Non-Essential Architecture — Low-Level Design (LLD)

This document defines the API attachment points for subsystems present in `design.swift` but not required for the current text-inference walking skeleton.

## Scope

These attachment points are intentionally lightweight and safe to evolve:

- Image async inference pipeline
- Video async inference pipeline
- RAG data ingest/retrieval APIs
- Psychographic ingest/state APIs
- Concept-based quiz APIs
- Analytics KPI APIs

## Attachment Point Strategy

1. **Stable entry paths now**: Route shapes and method contracts are available immediately.
2. **Safe placeholders**: Endpoints return either:
   - `202 Accepted` for async job enqueue simulation; or
   - `501 Not Implemented` with actionable next steps.
3. **Auth compatibility**: All routes are mounted under `/api/v1/*` protected middleware.
4. **Incremental implementation**: Each endpoint can later be wired to concrete service interfaces without changing public path contracts.

## Implemented API Attachment Points

### Discovery

- `GET /api/v1/architecture/attachment-points`
  - Returns a machine-readable catalog of non-essential API contracts.

### Image Inference (async)

- `POST /api/v1/image/jobs`
  - Returns `202 Accepted` with `job_id` and queue name (`image-jobs`).
- `GET /api/v1/image/jobs/{id}`
  - Returns `501 Not Implemented` placeholder.

### Video Inference (async)

- `POST /api/v1/video/jobs`
  - Returns `202 Accepted` with `job_id` and queue name (`video-jobs`).
- `GET /api/v1/video/jobs/{id}`
  - Returns `501 Not Implemented` placeholder.

### RAG

- `POST /api/v1/rag/documents` (placeholder)
- `POST /api/v1/rag/search` (placeholder)

### Psychographic Intelligence

- `POST /api/v1/psychographic/events` (placeholder)
- `GET /api/v1/psychographic/persona/{user_id}` (placeholder)

### Concept-based Questioning

- `POST /api/v1/quiz/generate` (placeholder)
- `POST /api/v1/quiz/attempts` (placeholder)

### Analytics

- `GET /api/v1/analytics/kpi` (placeholder)

## Next Implementation Layers

For each placeholder endpoint, complete in this order:

1. domain request/response schema
2. service interface + implementation
3. persistence/queue adapter
4. observability (logs, tracing, metrics)
5. integration and contract tests
