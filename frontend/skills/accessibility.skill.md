# accessibility.skill.md

## Purpose
Maintain usable and navigable UI from day one.

## Applied Guidance
- Use semantic controls (real buttons/inputs/textareas).
- Keep labels/placeholders and visible loading affordances.
- Preserve focus styles (`focus-visible` ring patterns).
- Ensure critical actions are keyboard reachable.

## Implementation Notes
- Inputs and action buttons use clear states and disabled handling.
- `focus-visible` styles are included in reusable primitives.
- Next improvement: add explicit label components and keyboard regression checks.
