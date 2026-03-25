---
phase: 03-unified-rendering
plan: "03-04"
subsystem: render
tags: [chroma, unified-diff, syntax-highlighting, hunk-headers, line-prefixes]

# Dependency graph
requires:
  - phase: 03-unified-rendering/03-01
    provides: internal/theme — dark background detection
  - phase: 03-unified-rendering/03-02
    provides: internal/highlight — HighlightLine, DetectLexer, FormatterForProfile, SelectTheme
  - phase: 03-unified-rendering/03-03
    provides: highlight.DetectLexer with filename/content fallback; WithLang/WithTheme options
provides:
  - internal/render package with Unified() function and RenderConfig struct
  - Git-compatible unified diff output with @@ hunk headers and +/-/space prefixes
  - Per-line Chroma syntax highlighting pipeline integrated into renderer
  - Fail-open rendering: plain text used on highlight error
affects:
  - 03-05 (NoColor + color depth wiring consumes RenderConfig)
  - phase 5 CLI (will call render.Unified via drift.Render public API)

# Tech tracking
tech-stack:
  added: []
  patterns:
    - RenderConfig struct passes pre-resolved Chroma Lexer/Style/Formatter into renderer
    - Nil-safe defaults in Unified() — callers can pass partial configs
    - Fail-open on HighlightLine error — plain content preferred over error propagation

key-files:
  created:
    - internal/render/unified.go
    - internal/render/unified_test.go
  modified: []

key-decisions:
  - "RenderConfig holds pre-resolved Chroma objects (not strings) — avoids redundant lexer/style lookups per-render call"
  - "Unified() fail-open on HighlightLine error: uses plain content rather than returning error — diff output always appears"
  - "No file headers written when Hunks is empty — matches git diff behavior for identical files"

patterns-established:
  - "render.RenderConfig: pre-resolve all Chroma objects before passing to renderer for efficiency"
  - "linePrefix() helper: clean switch on Op, default returns space for Equal"

requirements-completed: [REND-01, REND-03, REND-04]

# Metrics
duration: 8min
completed: 2026-03-25
---

# Plan 03-04: UnifiedRenderer with hunk headers and +/- prefixes

**`internal/render` package with `Unified()` writing Git-compatible unified diff — @@ hunk headers, +/-/space prefixes, per-line Chroma highlighting**

## Performance

- **Duration:** ~8 min
- **Started:** 2026-03-25T22:15:00Z
- **Completed:** 2026-03-25T22:23:00Z
- **Tasks:** 2
- **Files modified:** 2

## Accomplishments
- `internal/render/unified.go` exports `Unified(result, w, cfg)` — full Git-compatible unified diff output
- `RenderConfig` struct carries pre-resolved Chroma Lexer/Style/Formatter for efficient multi-call rendering
- Per-line syntax highlighting via `highlight.HighlightLine` with fail-open fallback on error
- 6 unit tests covering: empty result, hunk headers, line prefixes, custom filenames, TrueColor ANSI output, nil Lexer fallback
- All tests pass with `-race` flag

## Task Commits

Each task was committed atomically:

1. **Task 03-04-01: Create internal/render/unified.go** - `939c1bf` (feat)
2. **Task 03-04-02: Write unit tests for UnifiedRenderer** - `5cab679` (test)

## Files Created/Modified
- `internal/render/unified.go` — UnifiedRenderer: RenderConfig struct, Unified() function, linePrefix() helper
- `internal/render/unified_test.go` — 6 unit tests covering all acceptance criteria

## Decisions Made
- Pre-resolved Chroma objects in `RenderConfig` rather than strings, to avoid redundant lookups on every render call
- Fail-open on `HighlightLine` error: plain content used, diff output never blocked
- Empty hunks → no output (no file headers), matching git diff behavior for identical files

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- `internal/render` package ready; `Unified()` is fully functional
- Plan 03-05 can now wire `WithNoColor()` and terminal color depth detection using the existing `RenderConfig.NoColor` and `RenderConfig.Profile` fields

---
*Phase: 03-unified-rendering*
*Completed: 2026-03-25*
