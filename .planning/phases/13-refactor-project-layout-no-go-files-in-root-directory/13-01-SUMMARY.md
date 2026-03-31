---
phase: 13-refactor-project-layout-no-go-files-in-root-directory
plan: "01"
subsystem: project-layout
tags: [refactor, layout, import-paths, go-modules]
dependency_graph:
  requires: []
  provides: [drift-subpackage-layout, clean-root-directory]
  affects: [cmd/drift, internal/testhelpers, internal/algo/myers, internal/hunk, examples]
tech_stack:
  added: []
  patterns: [go-subpackage-layout, git-mv-for-history-preservation]
key_files:
  created:
    - drift/drift.go
    - drift/options.go
    - drift/render.go
    - drift/builder.go
    - drift/types.go
    - drift/doc.go
    - drift/drift_test.go
    - drift/drift_algorithm_integration_test.go
    - drift/drift_property_test.go
    - drift/render_test.go
    - drift/benchmark_test.go
    - drift/builder_test.go
    - drift/options_test.go
  modified:
    - cmd/drift/main.go
    - internal/testhelpers/apply.go
    - internal/algo/myers/myers_test.go
    - internal/hunk/hunk_test.go
    - examples/basic/main.go
    - examples/builder/main.go
decisions:
  - "Used git mv to preserve file history when moving library files to drift/ subdirectory"
  - "Compiled drift binary at root was untracked (not in git) so was safely moved aside before creating drift/ directory"
  - "All 13 library source files moved atomically in a single git mv command"
metrics:
  duration: "144 seconds"
  completed: "2026-03-27"
  tasks_completed: 3
  files_changed: 19
---

# Phase 13 Plan 01: Move Library Files to drift/ Subdirectory Summary

**One-liner:** Moved 13 root-level Go library files into `drift/` subdirectory and updated all 10 dependent import paths from `github.com/tbcrawford/drift` to `github.com/tbcrawford/drift/drift`.

## What Was Built

Relocated all library source files from the Go module root to a `drift/` subdirectory, achieving a clean "no Go files in root" layout. Updated every consumer of the library to use the new import path. All 219 tests continue to pass with `go build ./...` and `go test ./...` both clean.

## Tasks Completed

| # | Task | Commit | Files |
|---|------|--------|-------|
| 1 | Move library files from root to `drift/` | `3dc4314` | 13 files moved via `git mv` |
| 2 | Update all import paths from drift to drift/drift | `f87a851` | 10 files updated |
| 3 | Run full test suite to verify refactor | — (verification only) | 219 tests pass |

## Verification Results

| Check | Result |
|-------|--------|
| `ls *.go 2>/dev/null \| wc -l` | `0` ✅ |
| `ls drift/*.go \| wc -l` | `13` ✅ |
| `go build ./...` | exit 0 ✅ |
| `go test ./...` | 219 tests pass ✅ |
| `go vet ./...` | no issues ✅ |
| Old import path count | `0` ✅ |
| New import path count | `10` (≥10 required) ✅ |

## Files Created/Moved

### Moved to `drift/` (formerly at module root)
- `drift/drift.go` — `Diff()` function entry point
- `drift/options.go` — Algorithm/Option/config types and `WithXxx()` functions
- `drift/render.go` — `Render()`/`RenderWithNames()` functions
- `drift/builder.go` — Builder struct and `New()`
- `drift/types.go` — Exported type aliases (Op, Edit, Span, Line, Hunk, DiffResult)
- `drift/doc.go` — Package-level godoc with `#` section headers
- `drift/drift_test.go` — Main library integration tests
- `drift/drift_algorithm_integration_test.go` — Algorithm-specific tests
- `drift/drift_property_test.go` — Property-based round-trip tests
- `drift/render_test.go` — Render function tests
- `drift/benchmark_test.go` — Benchmarks (white-box)
- `drift/builder_test.go` — Builder API tests (white-box)
- `drift/options_test.go` — Options function tests (white-box)

### Updated Import Paths
- `drift/drift_test.go` — `drift_test` package
- `drift/drift_algorithm_integration_test.go` — `drift_test` package
- `drift/drift_property_test.go` — `drift_test` package
- `drift/render_test.go` — `drift_test` package
- `cmd/drift/main.go` — CLI entry point
- `internal/testhelpers/apply.go` — Test helper
- `internal/algo/myers/myers_test.go` — Myers algorithm tests
- `internal/hunk/hunk_test.go` — Hunk builder tests
- `examples/basic/main.go` — Basic example
- `examples/builder/main.go` — Builder API example

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Compiled drift binary at root blocked drift/ directory creation**
- **Found during:** Task 1
- **Issue:** An untracked compiled `drift` binary (9.4MB, Mach-O arm64) existed at the module root. This would have blocked `git mv` from creating a `drift/` directory since a file named `drift` already existed at that path.
- **Fix:** Moved the untracked binary to `/tmp/drift_binary_backup` before executing `git mv`. Binary is a build artifact and should not be tracked in git.
- **Files modified:** None (binary was untracked)
- **Commit:** No separate commit needed (not a tracked file)

## Key Decisions

1. **Used `git mv` for history preservation** — Moving files via `git mv` retains rename history in `git log --follow`, which is valuable for tracing the evolution of library files.

2. **Single atomic `git mv` command** — All 13 files moved in one operation to keep the rename-to-move history clean and atomic.

3. **`internal/` directory structure unchanged** — Per Go's internal package visibility rules, `github.com/tbcrawford/drift/drift` can still import `github.com/tbcrawford/drift/internal/...` packages because the drift package's module root is `github.com/tbcrawford/drift`.

## Known Stubs

None — all existing functionality was preserved exactly; this was a pure file relocation and import path update.

## Self-Check: PASSED

All created/moved files verified present; commits verified in git log.
