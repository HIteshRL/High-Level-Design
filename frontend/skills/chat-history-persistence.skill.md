# Skill: Conversation History Persistence (API + UI State)

## Goal
Persist and restore chat history per conversation by fetching canonical messages from backend storage.

## Preferred Pattern
1. Backend exposes `GET /api/v1/conversations/{id}/messages`.
2. Frontend selects conversation ID as source-of-truth.
3. On selection, fetch and render history.
4. On initial load, optionally hydrate newest conversation.
5. Keep optimistic sends for responsiveness; rely on fetch for canonical history.

## Guardrails
- Enforce ownership checks server-side.
- Handle 403/404/500 distinctly in client UX.
- Clear history view only on explicit new conversation or logout.

## Validation Checklist
- Refreshing page preserves history for saved conversations.
- Switching conversations swaps message history correctly.
- No cross-user/cross-conversation leakage.
