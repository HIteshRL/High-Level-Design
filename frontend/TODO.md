# Frontend + Backend Iteration TODO (Convergence Pass)

## A) Process Discipline (Do First)
- [ ] Add/refresh focused `skills/*.skill.md` docs for this iteration:
	- code-block syntax highlighting
	- resizable/collapsible layouts
	- persisted chat history contract
	- context-aware cache-key strategy
- [ ] Update `SELF_PROMPT.md` with this iteration’s risk checks.
- [ ] Update `DEVILS_ADVOCATE.md` with explicit tradeoff analysis and final convergence calls.

## B) Backend Contract Upgrades
- [ ] Add protected endpoint: `GET /api/v1/conversations/{id}/messages`.
- [ ] Add orchestrator method that verifies conversation ownership before returning messages.
- [ ] Ensure response order is chronological and fields are UI-ready.
- [ ] Return correct status mapping: `404` (not found), `403` (wrong owner).

## C) Context-Aware Caching
- [ ] Replace prompt-only cache keying with context-aware keying.
- [ ] Build cache fingerprint from final LLM input context:
	- system/context messages
	- historical messages
	- current user prompt
	- model + temperature + max tokens + user identity
- [ ] Keep deterministic JSON serialization before hashing.
- [ ] Add/adjust unit tests to prove key changes when context changes.

## D) Frontend Persistence + UX
- [ ] Extend API client/types for conversation-message history endpoint.
- [ ] On conversation selection, fetch persisted messages from backend and render them.
- [ ] On initial load, auto-select most recent conversation and load its history.
- [ ] Preserve current optimistic send UX while reconciling with persisted state.

## E) Frontend Rendering + Layout
- [ ] Add syntax highlighting for fenced markdown code blocks.
- [ ] Keep inline-code rendering separate from fenced blocks.
- [ ] Implement desktop side panel:
	- [ ] draggable width (within min/max constraints)
	- [ ] collapse/expand control
	- [ ] persisted width/collapse preference

## F) Validation and Delivery
- [ ] Run backend tests (`go test ./...`) and ensure compile success.
- [ ] Run frontend lint/build and verify no regressions.
- [ ] Manually verify end-to-end flow:
	- login/register
	- select conversation and see historical messages
	- new prompt round-trip
	- syntax-highlighted code rendering
	- sidebar resize/collapse behavior
- [ ] Commit and push a single integrated convergence commit.
