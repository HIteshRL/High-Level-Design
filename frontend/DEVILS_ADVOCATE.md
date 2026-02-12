# Devil's Advocate Review (Frontend)

## 1) Could this architecture be over-engineered?
- **Risk:** Adding too many abstractions early slows feature delivery.
- **Decision:** Keep MVP with local state and typed API module; defer heavy global state libraries.

## 2) Could token storage in localStorage be abused?
- **Risk:** XSS could expose token.
- **Decision:** Accept for MVP/local tool. For production, move to httpOnly cookies + CSRF strategy.

## 3) Could UI claim features backend doesn't support yet?
- **Risk:** Mismatch creates user confusion.
- **Decision:** Avoid streaming UI and advanced controls until backend guarantees behavior.

## 4) Could conversation switching be misleading without history fetch API?
- **Risk:** Users expect historic messages when selecting a conversation.
- **Decision:** Clear local behavior for now; next iteration should add `GET /conversations/:id/messages`.

## 5) Could error handling hide useful debugging info?
- **Risk:** Generic errors harm debugging.
- **Decision:** Show friendly messages to users; keep structured details in browser dev tools/network.

## 6) Could design system drift happen quickly?
- **Risk:** Ad-hoc styles break consistency.
- **Decision:** Force new UI through shared primitives and tokenized styles.

## 7) Could dark theme create unreadable combinations?
- **Risk:** Token choices may look good in one screen but fail in others.
- **Decision:** Use semantic variables only (`background`, `foreground`, `muted`, etc.) and avoid hard-coded theme colors in feature components.

## 8) Could advanced controls confuse users?
- **Risk:** Too many knobs can reduce usability.
- **Decision:** Keep controls compact, prefilled with sane defaults, and colocate them above message stream for clear mental model.

## 9) Could plain-text rendering degrade answer quality perception?
- **Risk:** Useful structured outputs (lists/code/tables) become unreadable.
- **Decision:** Render markdown with GFM support and strong defaults for code/table readability.

## 10) Could keyboard submit hurt multiline composition?
- **Risk:** Enter-to-send can conflict with multiline editing and IME input.
- **Decision:** Use Enter-to-send + Shift+Enter newline and guard against IME composition states.

## Convergence Outcome
- Current implementation balances speed, maintainability, and accuracy for MVP.
- Highest-value next steps: conversation message-history endpoint integration, component tests, accessibility pass, and optional streaming renderer.

---

# Devil's Advocate Review (Convergence Iteration)

## 1) Is syntax highlighting worth added dependency weight?
- **Risk:** Extra parser/highlighter cost can bloat bundle.
- **Decision:** Use `rehype-highlight` with existing markdown pipeline; avoid heavy editor-grade dependencies.

## 2) Can resizable sidebars break viewport fit and accessibility?
- **Risk:** Drag interactions can create overflow or inaccessible control paths.
- **Decision:** Constrain width with explicit min/max bounds, provide one-click collapse/expand, and persist preferences safely.

## 3) Could history persistence cause stale or cross-conversation data leaks?
- **Risk:** Missing ownership checks or stale state transitions may show wrong messages.
- **Decision:** Add backend ownership verification per conversation and always hydrate on explicit conversation switch.

## 4) Could context-aware caching still be semantically wrong?
- **Risk:** Prompt-only cache keys ignore prior turns and system context.
- **Decision:** Hash deterministic full message context plus model parameters and user identity.

## 5) Could optimistic UI conflict with persisted history?
- **Risk:** Duplicate or reordered messages after response arrives.
- **Decision:** Keep optimistic append for responsiveness; on conversation switch, re-fetch canonical persisted history.

## 6) Could one massive commit increase rollback risk?
- **Risk:** Hard to isolate regressions.
- **Decision:** Implement in convergent passes (backend contract → frontend wiring → UX polish), with lint/build/test after each.

## Final Convergence Calls
- Keep architecture minimal, explicit, and typed.
- Favor deterministic behavior over magical client assumptions.
- Prefer bounded interactions (resize limits, prompt limits, ownership checks).
- Ship with validated builds/tests and clear follow-up targets (streaming UI, E2E coverage).
