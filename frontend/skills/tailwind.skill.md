# tailwind.skill.md

## Purpose
Use utility-first CSS with semantic design tokens for rapid iteration and consistency.

## Applied Guidance
- Configure Tailwind via Vite plugin (`@tailwindcss/vite`).
- Define semantic tokens for background, foreground, card, border, etc.
- Use responsive utility classes directly in components.
- Keep styles colocated with components to reduce context switching.

## Implementation Notes
- Tailwind imported in `src/index.css`.
- Theme tokens (light/dark capable) defined in CSS variables.
- Layout and spacing implemented through utility classes only.
