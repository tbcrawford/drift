---
phase: 03-unified-rendering
plan: "03-05"
subsystem: rendering
tags: [chroma, colorprofile, ansi, no-color, lipgloss, render]

# Dependency graph
requires:
  - phase: 03-unified-rendering/03-04
    provides: internal/render/unified.go with Unified() renderer and RenderConfig struct
  - phase: 03-unified-rendering/03-02
    provides: internal/highlight/ with DetectLexer, SelectTheme, FormatterForProfile
  - phase: 03-unified-rendering/03-01
    provides: internal/theme/ with DetectDarkBackground
provides:
  - "drift.Render(result DiffResult, w io.Writer, opts ...Option) error — public one-call API"
  - "drift.RenderWithNames() with custom file header labels"
  - "resolveProfile() wiring NO_COLOR env var, *os.File detection, and WithNoColor() option"
  - "Integration tests verifying no ANSI on NoTTY writers, NoColor option, NoColorEnvVar, hunk headers, RenderWithNames"
affects: [04-split-rendering, 05-cli, 06-api-hardening]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "resolveProfile: noColor flag → NO_COLOR env var → *os.File detect → NoTTY fallback"
    - "colorprofile.NewWriter wraps writer for automatic ANSI downsampling"
    - "Public Render() mirrors Diff() options pattern: defaultConfig() + option loop"

key-files:
  created:
    - render.go
    - render_test.go
  modified: []

key-decisions:
  - "resolveProfile checks cfg.noColor and NO_COLOR env before *os.File detection — explicit user intent wins"
  - "colorprofile.NewWriter used for ANSI downsampling even for non-file writers (env-only profile detection)"
  - "Non-file writers default to colorprofile.NoTTY → NoOp formatter → plain text output"

patterns-established:
  - "Public drift.Render() follows same options application pattern as drift.Diff()"
  - "resolveProfile extracted as package-level func (unexported) for testability"

requirements-completed: [REND-01, REND-05, REND-06, REND-07, REND-08, REND-09]

# Metrics
duration: 5min
completed: 2026-03-25
---

# Plan 03-05: Public drift.Render() API, WithNoColor, and color depth wire-up — Summary

**Public `drift.Render()` API wired to colorprofile detection, theme resolution, and the unified renderer with full NO_COLOR/noColor/NoTTY support**

## Performance

- **Duration:** ~5 min
- **Started:** 2026-03-25T22:23:00Z
- **Completed:** 2026-03-25T22:28:00Z
- **Tasks:** 3
- **Files modified:** 2

## Accomplishments
- `drift.Render()` and `drift.RenderWithNames()` implemented, wiring together `resolveProfile`, `DetectDarkBackground`, `DetectLexer`, `SelectTheme`, `FormatterForProfile`, and `render.Unified()`
- `resolveProfile()` correctly resolves: explicit `WithNoColor()` → `NO_COLOR` env var → `*os.File` detection → `NoTTY` fallback
- 8 integration tests passing covering all acceptance criteria: `TestRender_WithNoColor`, `TestRender_PlainWriter`, `TestRender_HunkHeaderFormat`, `TestRender_EqualInputsNoOutput`, `TestRender_WithLang`, `TestRender_WithTheme`, `TestRenderWithNames`, `TestRender_NoColorEnvVar`
- Full suite: `go test ./...`, `go test -race ./...`, `go vet ./...`, `go build ./...` all exit 0

## Task Commits

1. **Task 03-05-01: Implement render.go** — `7bad38f` (feat)
2. **Task 03-05-02: Integration tests for drift.Render()** — `2d84aad` (test)
3. **Task 03-05-03: Full suite verification** — (no commit needed — all checks passed)

## Files Created/Modified
- `render.go` — Public `Render()`, `RenderWithNames()`, and unexported `resolveProfile()` functions
- `render_test.go` — 8 integration tests for the public rendering API

## Decisions Made
- `resolveProfile` is unexported but extracted as a named func — keeps `Render()` readable while making the resolution logic auditable
- `colorprofile.NewWriter` wraps the writer for ANSI downsampling even for non-file writers; for non-TTY outputs the wrapper is a no-op pass-through (env-only detection)
- Non-file writers (`bytes.Buffer`, etc.) always use `colorprofile.NoTTY` regardless of environment — prevents color leakage into piped output

## Deviations from Plan
None — plan executed exactly as written.

## Issues Encountered
None — all tests passed on first run.

## User Setup Required
None — no external service configuration required.

## Next Phase Readiness
- Phase 3 (Unified Rendering) is now complete — all 5 plans done
- `drift.Render()` is the stable public rendering API; Phase 4 will add `drift.RenderSplit()` following the same pattern
- `render.go` and `internal/render/unified.go` provide the established patterns for Phase 4's split renderer

---
*Phase: 03-unified-rendering*
*Completed: 2026-03-25*
