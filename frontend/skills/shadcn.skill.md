# shadcn.skill.md

## Purpose
Use a token-driven design system with reusable UI primitives that can scale without visual drift.

## Applied Guidance
- Use composable component primitives (`Button`, `Input`, `Card`, `Tabs`, etc.).
- Use `cva` + `cn()` patterns for variant-safe styling.
- Keep components framework-agnostic and focused on ergonomics.
- Build from primitives first, feature components second.

## Implementation Notes
- Added shadcn-style primitives under `src/components/ui/*`.
- Added alias import support (`@/*`) for clean imports.
- Centralized class merging with `clsx + tailwind-merge`.
