# Skill: Context-Aware Inference Caching (Backend)

## Goal
Prevent stale/incorrect cache hits by keying on full semantic context instead of prompt-only text.

## Preferred Pattern
1. Build cache fingerprint from final LLM input messages (system + history + current prompt).
2. Include model and decoding params (`temperature`, `max_tokens`).
3. Include user identity partitioning.
4. Serialize deterministically (JSON) and hash (SHA-256).

## Guardrails
- Never share cache keys across users.
- Key must change when any prior message changes.
- Keep cache read/write failures non-fatal.

## Validation Checklist
- Same context => same key.
- Different history => different key.
- Different model/params => different key.
- Tests cover semantic key boundaries.
