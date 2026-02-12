# Skill: Markdown Code Syntax Highlighting (React + react-markdown)

## Goal
Render fenced code blocks with language-aware syntax highlighting while keeping inline code simple and fast.

## Preferred Pattern
1. Keep `react-markdown` as renderer.
2. Add `rehype-highlight` plugin for fenced code blocks.
3. Keep inline code path custom-styled (no highlighter pass).
4. Ensure long blocks are horizontally scrollable and visually distinct.

## Guardrails
- Avoid executing any code from markdown.
- Keep code rendering pure presentation.
- Keep fallback behavior for unknown languages.

## Validation Checklist
- Fenced code block shows token colors.
- Inline code remains compact and unbroken.
- Works in light and dark themes.
- Lint/build pass without type issues.
