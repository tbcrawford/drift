---
phase: 17-address-medium-priority-council-review-issues
plan: "03"
subsystem: testing/rendering
tags: [golden-tests, snapshot-testing, goldie, rendering, ci]
dependency_graph:
  requires: [17-01]
  provides: [rendering-regression-coverage]
  affects: [golden_test.go, testdata/golden/, go.mod, options.go, render.go]
tech_stack:
  added: [github.com/sebdah/goldie/v2 v2.8.0]
  patterns: [snapshot-testing, goldie-fixture-dir, WithNoColor-for-ci-portable-fixtures]
key_files:
  created:
    - golden_test.go
    - testdata/golden/unified_go.golden
    - testdata/golden/split_go.golden
    - testdata/golden/nocolor_basic.golden
  modified:
    - go.mod
    - go.sum
    - options.go
    - render.go
decisions:
  - "WithTermWidth(w int) added to public API to allow deterministic split-view width in tests and non-TTY environments"
  - "Golden fixtures use WithNoColor() — plain text, no ANSI, CI-portable; goldie.WithFixtureDir(testdata/golden) used for explicit fixture location"
  - "buildRenderPipeline now honours cfg.render.termWidth when non-zero, falling back to automatic TerminalWidth detection"
metrics:
  duration: "~8 minutes"
  completed: "2026-03-31"
  tasks_completed: 2
  files_changed: 8
requirements_satisfied: [REVIEW-07]
---

# Phase 17 Plan 03: Goldie Snapshot Tests for Rendering Pipeline Summary

**One-liner:** Goldie v2 snapshot tests for unified/split/no-color rendering paths with plain-text CI-portable fixtures and `WithTermWidth` option for deterministic output.

## What Was Built

Added `github.com/sebdah/goldie/v2 v2.8.0` as a test dependency and created `golden_test.go` with three snapshot tests that cover the core rendering paths:

- **TestGolden_UnifiedRenderer** — unified diff of a Go import refactor (syntax highlighting path)
- **TestGolden_SplitRenderer** — 120-col split diff of the same Go change (fixed-width layout path)
- **TestGolden_NoColorOutput** — plain-text unified diff for a minimal line change (no-color path)

All three tests use `WithNoColor()` so the golden fixtures are ANSI-free and reproducible across CI environments. Fixtures live at `testdata/golden/` adjacent to the library source.

As a prerequisite deviation (Rule 2), `WithTermWidth(w int) Option` was added to the public API in `options.go` and wired through `buildRenderPipeline` in `render.go`. This allows tests to produce deterministic split-view output at a fixed column width rather than depending on terminal detection.

## Tasks Completed

| Task | Description | Commit |
|------|-------------|--------|
| 1 | Add goldie v2, WithTermWidth option, create golden_test.go | 23b45e7 |
| 2 | Generate golden snapshots with -update, verify CI mode passes | 3797bbf |

## Verification Results

```
✅ go.mod contains github.com/sebdah/goldie/v2 v2.8.0
✅ golden_test.go: TestGolden_UnifiedRenderer, TestGolden_SplitRenderer, TestGolden_NoColorOutput
✅ testdata/golden/ contains unified_go.golden, split_go.golden, nocolor_basic.golden
✅ go test -run TestGolden_ . passes (3 snapshot tests, without -update)
✅ Fixtures contain no ANSI escape codes (plain text, CI-portable)
✅ go test ./... passes 206 tests (203 baseline + 3 new golden tests)
✅ go vet ./... clean
```

## Deviations from Plan

### Auto-added Missing Functionality

**1. [Rule 2 - Missing Critical Functionality] Added WithTermWidth option to public API**
- **Found during:** Task 1 implementation
- **Issue:** The plan's `TestGolden_SplitRenderer` uses `drift.WithTermWidth(120)` for deterministic split output, but this option did not exist in `options.go`. Without it, split rendering width depends on terminal detection, making snapshots non-deterministic across environments.
- **Fix:** Added `WithTermWidth(w int) Option` to `options.go`; updated `renderConfig` struct to hold `termWidth int`; updated `buildRenderPipeline` in `render.go` to use `cfg.render.termWidth` when non-zero.
- **Files modified:** `options.go`, `render.go`
- **Commit:** 23b45e7

## Known Stubs

None. All three snapshot tests are fully wired to real rendering paths.

## Self-Check: PASSED

- ✅ `golden_test.go` exists
- ✅ `testdata/golden/unified_go.golden` exists (415B)
- ✅ `testdata/golden/split_go.golden` exists (1.6KB)
- ✅ `testdata/golden/nocolor_basic.golden` exists (122B)
- ✅ Commit 23b45e7 exists
- ✅ Commit 3797bbf exists
