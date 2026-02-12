# Skill: Resizable and Collapsible Sidebar (React Layout)

## Goal
Allow desktop users to resize sidebar width and collapse/expand it without breaking viewport-fit layout.

## Preferred Pattern
1. Keep container as CSS grid with dynamic `gridTemplateColumns`.
2. Persist width/collapsed state in localStorage.
3. Use pointer drag handle on desktop only.
4. Clamp width with strict min/max bounds.
5. Provide explicit collapse/expand button independent of drag.

## Guardrails
- Must not create page-level overflow.
- Drag must stop on pointer-up anywhere.
- Collapsed mode should preserve access to expand action.

## Validation Checklist
- Drag updates width smoothly.
- Width stays within min/max constraints.
- Collapse and expand persist between reloads.
- Mobile layout remains unaffected.
