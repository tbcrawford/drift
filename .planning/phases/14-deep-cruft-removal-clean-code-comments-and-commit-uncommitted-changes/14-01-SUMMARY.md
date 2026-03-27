---
phase: 14-deep-cruft-removal-clean-code-comments-and-commit-uncommitted-changes
plan: "01"
subsystem: infra

tags: [go, golangci-lint, justfile, chroma, lipgloss, gutter, word-diff]

requires:
  - phase: 13-refactor-project-layout-no-go-files-in-root-directory
    provides: Project layout with drift/ subdirectory and no Go files at root; working tests

provides:
  - Clean git working tree: 6 accumulated working-tree changes committed
  - golangci-lint v2 schema (version key, restructured exclusions)
  - justfile install recipe for go install ./cmd/drift
  - cmd/drift main_test.go covers --no-line-numbers flag
  - highlight.DiffLineStyle returns chroma.Colour (not lipgloss.Style)
  - render/gutter gutterColumnSeparator constant and split dim/high foreground
  - render/split_test updated NoColor separator assertion and word-diff test

affects: [phase-15, golangci-lint, testing, cli]

tech-stack:
  added: []
  patterns:
    - "gutterColumnSeparator = ' │' constant for Unicode box-drawing separator"
    - "Split dim (context) / high (delete/insert) foreground in gutter cells"
    - "DiffLineStyle returns chroma.Colour — callers format directly, no lipgloss.Style wrapper"
    - "golangci-lint v2 schema: version key at top, exclusions nested under linters/formatters"

key-files:
  created: []
  modified:
    - .golangci.yml
    - justfile
    - cmd/drift/main_test.go
    - internal/highlight/diff_line.go
    - internal/render/gutter.go
    - internal/render/split_test.go

key-decisions:
  - "golangci-lint v2 config: version key is required at top level; exclusions restructured under linters: and formatters: blocks"
  - "justfile install recipe uses go install ./cmd/drift (not go build)"
  - "DiffLineStyle returns chroma.Colour instead of lipgloss.Style — leaner API; callers apply directly to rendering pipeline"
  - "gutterColumnSeparator uses U+2502 BOX DRAWINGS LIGHT VERTICAL (│), not ASCII pipe (|)"

patterns-established:
  - "gutter.go: gutterColumnSeparator constant ensures single source of truth for separator character"
  - "gutter.go: dim vs high foreground distinguishes context rows from changed rows without a background fill"

requirements-completed:
  - CRUFT-01

duration: 3min
completed: "2026-03-27"
---

# Phase 14 Plan 01: Commit Accumulated Working-Tree Changes Summary

**golangci-lint migrated to v2 schema, justfile install recipe added, and 4 code quality changes across highlight/gutter committed — 6 files, 219 tests green, clean working tree**

## Performance

- **Duration:** ~3 min
- **Started:** 2026-03-27T14:44:00Z
- **Completed:** 2026-03-27T14:46:58Z
- **Tasks:** 2
- **Files modified:** 6

## Accomplishments

- Committed 6 accumulated working-tree changes that were working but unstaged
- golangci-lint v2 schema migration: `version: "2"` key, `exclusions:` restructured under `linters:` and `formatters:` blocks, `gosimple` removed (not in v2)
- justfile `install:` recipe enables `go install ./cmd/drift` via `just install`
- `render/gutter.go`: extracted `gutterColumnSeparator` constant and split dim/high foreground for context vs changed rows
- `highlight/diff_line.go`: `DiffLineStyle` now returns `chroma.Colour` directly, removing `ApplyDiffLineStyle` wrapper
- `render/split_test.go`: `TestSplit_NoColorSeparator` updated to expect `│` (U+2502); `TestSplit_WordDiffPairedDeleteInsert` added

## Task Commits

Each task was committed atomically:

1. **Task 1+2: Verify working state and commit accumulated changes** - `3c27d73` (chore)

## Files Created/Modified

- `.golangci.yml` — migrated to golangci-lint v2 schema (version key, restructured exclusions)
- `justfile` — added `install:` recipe
- `cmd/drift/main_test.go` — added `--no-line-numbers` to help smoke test
- `internal/highlight/diff_line.go` — `DiffLineStyle` returns `chroma.Colour`; removed `ApplyDiffLineStyle`
- `internal/render/gutter.go` — `gutterColumnSeparator` constant; split dim/high foreground
- `internal/render/split_test.go` — updated `NoColor` separator assertion; added `WordDiffPairedDeleteInsert` test

## Decisions Made

- golangci-lint v2 schema uses `version: "2"` at top; `exclusions:` block moves under `linters:` and `formatters:` sections separately
- `DiffLineStyle` returns `chroma.Colour` not `lipgloss.Style` — leaner, avoids unnecessary wrapping
- `gutterColumnSeparator` uses U+2502 (│) not ASCII pipe (|) — matches Unicode box-drawing convention for TUI borders

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] .gitignore `drift` pattern matched `cmd/drift/` directory**
- **Found during:** Task 2 (staging files)
- **Issue:** `.gitignore` line `drift` was broad enough to match the `cmd/drift/` path, causing `git add cmd/drift/main_test.go` via directory path to fail
- **Fix:** Used individual `git add` calls for each file to work around the pattern; all 6 files were successfully staged (the files were already tracked so they staged correctly)
- **Files modified:** None — workaround only; `.gitignore` fix is tracked as a known issue for future cleanup
- **Verification:** `git status` showed all 6 files staged; commit succeeded; `git status` after commit shows clean tracked files
- **Committed in:** 3c27d73 (task commit)

---

**Total deviations:** 1 (workaround for .gitignore breadth; no code changes required)
**Impact on plan:** Minimal — all 6 files committed successfully despite the staging hiccup.

## Issues Encountered

- `.gitignore` pattern `drift` is too broad — it matches `cmd/drift/` directory in addition to the root binary. Files are already tracked so staging works on individual files, but `git add cmd/drift/` would fail. The pattern should be `/drift` (root-anchored) to only match the root binary. Deferred to a separate cleanup since it doesn't affect tests or builds.

## Next Phase Readiness

- Working tree is clean — only `.planning/` files remain untracked/modified
- All 219 tests pass; `go vet` clean
- Ready for plan 02: deep cruft removal (code comments, naming, cleanup)
- `.gitignore` breadth issue (`drift` vs `/drift`) should be addressed in plan 02

---
*Phase: 14-deep-cruft-removal-clean-code-comments-and-commit-uncommitted-changes*
*Completed: 2026-03-27*
