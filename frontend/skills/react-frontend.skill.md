# react-frontend.skill.md

## Purpose
Build predictable, typed UI flows with clear state ownership.

## Applied Guidance
- Lift shared state to app-level (`token`, `user`, `conversations`, `messages`).
- Keep API side effects in dedicated client module.
- Keep feature components focused (`AuthPanel`, `ChatPanel`).
- Handle loading and failures explicitly.

## Implementation Notes
- `src/lib/api.ts` centralizes HTTP calls and typed errors.
- `src/App.tsx` orchestrates auth bootstrap and chat interactions.
- Components are stateless where possible and receive explicit props.
