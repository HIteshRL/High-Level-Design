# Frontend Implementation TODO (Detailed)

## Phase 1 — Product Definition and Boundaries
- [x] Define MVP scope: auth, chat inference, conversations list, error/loading states.
- [x] Freeze backend contract used by UI: `/auth/register`, `/auth/token`, `/auth/me`, `/inference/complete`, `/conversations`.
- [x] Decide state strategy: local React state + typed API client (no heavy state manager for MVP).
- [x] Define non-goals for v1: no streaming renderer, no markdown parser, no route-level SSR.

## Phase 2 — Frontend Foundation
- [x] Scaffold Vite + React + TypeScript app.
- [x] Add Tailwind v4 via `@tailwindcss/vite` plugin.
- [x] Configure aliasing (`@/*`) across Vite + TypeScript.
- [x] Set global design tokens and semantic color scales.

## Phase 3 — UI System (shadcn-style)
- [x] Install shadcn-friendly dependencies (`cva`, `clsx`, `tailwind-merge`, Radix primitives).
- [x] Implement reusable UI primitives: Button, Input, Textarea, Card, Tabs, Badge, Switch, ScrollArea, Separator.
- [x] Add utility `cn()` helper and strict class composition patterns.

## Phase 4 — Data and API Layer
- [x] Add typed API client with central error handling.
- [x] Add request/response model types for auth, user, conversation, inference.
- [x] Add token persistence strategy (localStorage key namespace).
- [x] Add auth bootstrap flow (`me` + initial `conversations`).

## Phase 5 — Feature Surfaces
- [x] Build Login/Register panel with validation and clear UX feedback.
- [x] Build Chat workspace with conversation list and prompt composer.
- [x] Bind prompt submit to backend inference endpoint.
- [x] Sync conversations after each successful completion.
- [x] Add robust busy/error handling and toast feedback.

## Phase 6 — Quality Gates
- [ ] Add component tests (AuthPanel, ChatPanel) with mocked API.
- [ ] Add API contract tests for parse/error boundaries.
- [ ] Add E2E smoke test for auth + inference path.
- [ ] Add accessibility pass (focus order, labels, contrast, keyboard).

## Phase 7 — Delivery and Ops
- [x] Add frontend env template (`VITE_API_BASE_URL`).
- [ ] Add npm scripts for lint + test + format.
- [ ] Add CI job to build frontend and run tests.
- [ ] Document local run flow and troubleshooting in frontend README.
