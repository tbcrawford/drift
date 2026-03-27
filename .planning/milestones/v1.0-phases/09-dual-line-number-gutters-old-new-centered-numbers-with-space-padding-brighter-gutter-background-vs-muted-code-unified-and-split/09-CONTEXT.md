# Phase 9: Dual line-number gutters — Context

**Gathered:** 2026-03-26
**Status:** Ready for planning
**Mode:** Auto-generated (autonomous — aligned with 09-RESEARCH.md and approved plans 09-01 / 09-02)

<domain>
## Phase Boundary

Shared gutter helpers (width scan, centered numerals, Lip Gloss gutter styles with NoColor width-only fallback); unified output `old | new | prefix + code`; split output with gutters inside each panel and recomputed code widths; `WithLineNumbers` / `WithoutLineNumbers` (default: show gutters).

</domain>

<decisions>
## Implementation Decisions

### Gutter layout
- Separator between old and new gutter columns: ` │ ` (Unicode) or ` | ` when `NoColor`.
- Default: line numbers on; CLI `--no-line-numbers` maps to `WithoutLineNumbers()`.

### Claude's Discretion
- Gutter background colors: fixed ANSI256-style lipgloss colors for dark/light; muted distinct old vs new columns.

</decisions>

<code_context>
## Existing Code Insights

Unified and split renderers in `internal/render/`; `RenderConfig` extended with `ShowLineNumbers` and `IsDark`; options plumbed from `render.go`.

</code_context>

<specifics>
## Specific Ideas

Match existing `separatorNoColor` / `separatorColor` split behavior for gutter separators.

</specifics>

<deferred>
## Deferred Ideas

Theme-tinted full-line add/delete (Phase 10) and word-level intra-line highlights (Phase 11).

</deferred>
