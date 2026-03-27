# Phase 10: Theme-aware full-line diff styling — Context

**Gathered:** 2026-03-26
**Status:** Ready for planning
**Mode:** Auto-generated (autonomous — recommended defaults from ROADMAP + Phase 9 deferred notes)

<domain>
## Phase Boundary

Theme-derived full-line backgrounds for added/removed code in unified and split views: derive tint from the active Chroma style’s `GenericDeleted` / `GenericInserted` (with blend/fallback when only foreground is set); preserve Chroma syntax colors by downgrading `\x1b[0m` to `\x1b[39m` between tokens; expose `WithLineDiffStyle` (default on) and builder wiring.

</domain>

<decisions>
## Implementation Decisions

### Visual / ANSI
- Full-line tint applies to unified `prefix + highlighted` (not gutter columns).
- Chroma TTY output resets full SGR between tokens; replace `\x1b[0m` with `\x1b[39m` before wrapping with Lip Gloss background so line-level background persists.
- No new runtime dependencies; Lip Gloss + Chroma only.

### API
- `WithLineDiffStyle(bool)` — default `true` in `defaultConfig`; `RenderConfig.LineDiffStyle` passed from `drift.Render` / `RenderWithNames`.

### Claude's Discretion
- When Chroma has no background for gd/gi, blend foreground toward a dark/light terminal base; if no gd/gi colours, use fixed subtle red/green-tint fallbacks.

</decisions>

<code_context>
## Existing Code Insights

`internal/highlight/HighlightLine`, `internal/render/unified.go`, `internal/render/split.go`, `render.go` option plumbing; Phase 9 gutters remain separate from code-column tint.

</code_context>

<specifics>
## Specific Ideas

Phase 9 CONTEXT deferred “theme-tinted full-line add/delete” to this phase.

</specifics>

<deferred>
## Deferred Ideas

Intra-line word highlights — Phase 11.

</deferred>
