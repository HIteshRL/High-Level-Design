# Roognis Frontend (React + Tailwind + shadcn-style UI)

This frontend is a Vite + React + TypeScript client for the Roognis Go backend.

## Features

- Auth flow: register, login, profile bootstrap.
- Inference workspace: prompt entry + AI response rendering.
- Conversations panel: lists user conversations from backend.
- UI stack: Tailwind v4 + Radix primitives + shadcn-style components.

## Prerequisites

- Node.js 20+
- Backend running at `http://localhost:8080`

## Configuration

Create `.env` from the template:

```bash
cp .env.example .env
```

`VITE_API_BASE_URL` defaults to `http://localhost:8080`.

## Run

```bash
npm install
npm run dev
```

## Build

```bash
npm run build
```

## Key folders

- `src/components/ui`: reusable shadcn-style primitives
- `src/components`: feature components (`AuthPanel`, `ChatPanel`)
- `src/lib`: API client, shared types, utility helpers
- `skills/*.skill.md`: implementation skill notes used for this delivery
