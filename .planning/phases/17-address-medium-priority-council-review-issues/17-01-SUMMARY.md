---
phase: 17-address-medium-priority-council-review-issues
plan: "01"
subsystem: module-layout
tags: [go-modules, import-path, workspace, refactor]
dependency_graph:
  requires: []
  provides: [canonical-import-path, go-workspace, separate-cli-module]
  affects: [all-import-sites, examples, cmd/drift, internal packages]
tech_stack:
  added: [go.work, cmd/drift/go.mod, go workspace]
  patterns: [multi-module workspace, replace directive, git mv for history preservation]
key_files:
  created:
    - go.work
    - cmd/drift/go.mod
    - cmd/drift/go.sum
  modified:
    - benchmark_test.go (moved from drift/)
    - builder.go (moved from drift/)
    - builder_test.go (moved from drift/)
    - doc.go (moved from drift/)
    - drift.go (moved from drift/)
    - drift_algorithm_integration_test.go (moved from drift/)
    - drift_property_test.go (moved from drift/)
    - drift_test.go (moved from drift/)
    - options.go (moved from drift/)
    - options_test.go (moved from drift/)
    - render.go (moved from drift/)
    - render_test.go (moved from drift/)
    - types.go (moved from drift/)
    - cmd/drift/main.go (import path updated)
    - cmd/drift/flags.go (import path updated)
    - internal/testhelpers/apply.go (import path updated)
    - internal/edittype/edittype.go (godoc import path updated)
    - internal/algo/myers/myers_test.go (import path updated)
    - internal/hunk/hunk_test.go (import path updated)
    - examples/basic/main.go (import path updated)
    - examples/builder/main.go (import path updated)
    - .gitignore (removed /drift and !/drift/, added go.work.sum and /cmd/drift/drift)
decisions:
  - "go.work replace directive used for local development: cmd/drift/go.mod uses replace github.com/tylercrawford/drift => ../.. so go mod tidy resolves correctly without requiring published module"
  - "go.work committed (not gitignored); go.work.sum gitignored (machine-generated)"
  - "git mv used for all 13 library files to preserve file history in git log"
  - "cmd/drift/drift compiled binary added to .gitignore (produced by go build ./... in CLI module)"
metrics:
  duration: "~12 minutes"
  completed: "2026-03-31"
  tasks_completed: 3
  files_modified: 21
requirements_satisfied: [REVIEW-05]
---

# Phase 17 Plan 01: Move Library to Module Root (Canonical Import Path) Summary

**One-liner:** Moved all 13 library files from `drift/` to module root via `git mv`, created `go.work` workspace + separate `cmd/drift/go.mod`, updated 12 import sites — canonical import path is now `github.com/tylercrawford/drift`.

## What Was Built

Resolved REVIEW-05 (import path friction: `drift/drift` double-path) by moving the library package from the `drift/` subdirectory to the module root. This makes the canonical import path `import "github.com/tylercrawford/drift"` — exactly what PROJECT.md's Core Value statement requires.

The CLI (`cmd/drift`) was extracted into its own Go module with a `go.work` workspace for seamless local development. All 223 tests continue to pass (203 library + 20 CLI).

## Tasks Completed

| Task | Name | Commit | Key Changes |
|------|------|--------|-------------|
| 1 | Move library files and update all import sites | `1023f41` | 13 files moved via git mv, 12 import sites updated, drift/ removed |
| 2 | Create cmd/drift/go.mod and go.work workspace | `ac099b4` | go.work, cmd/drift/go.mod, cmd/drift/go.sum, .gitignore updated |
| 3 | Verify all tests pass in both modules | `6f37390` | 223 tests verified, /cmd/drift/drift binary gitignored |

## Decisions Made

1. **`replace` directive in cmd/drift/go.mod** — `go mod tidy` inside a workspace sub-module cannot resolve workspace-provided modules via network (the repo doesn't yet exist on GitHub). The `replace github.com/tylercrawford/drift => ../..` directive enables `go mod tidy` to work correctly for local development. When the repo is published, this can be removed and `cmd/drift` can reference the published version directly.

2. **`go.work` committed, `go.work.sum` gitignored** — Convention for multi-module repos. The workspace file is a dev contract, the sum file is machine-generated.

3. **`git mv` for history preservation** — All 13 library files moved using `git mv` so `git log --follow` preserves their full history from when they were originally created.

4. **`/cmd/drift/drift` compiled binary gitignored** — Running `go build ./...` in the CLI module produces a `drift` binary in the `cmd/drift/` directory. Added to `.gitignore` to prevent accidental commits.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 2 - Missing] Updated `.gitignore` for post-reorganization accuracy**
- **Found during:** Task 2
- **Issue:** `.gitignore` had `/drift` and `!/drift/` entries that were confusing now that `drift/` directory is removed. `go.work.sum` was not gitignored despite plan requiring it. Compiled CLI binary needed ignoring.
- **Fix:** Removed stale `/drift` and `!/drift/` entries; added `go.work.sum`, `/cmd/drift/drift`
- **Files modified:** `.gitignore`
- **Commit:** `ac099b4`, `6f37390`

**2. [Rule 2 - Missing] `go.work` go version reflects local toolchain (1.26.1 not 1.21)**
- **Found during:** Task 2
- **Issue:** `go work init` uses the current toolchain version; `go mod tidy` upgraded cmd/drift/go.mod to `go 1.25.0`. Both are still >= 1.21 minimum requirement.
- **Fix:** Accepted toolchain-determined versions; no manual override needed (overriding would cause go toolchain downgrade warnings)
- **Impact:** None — both versions exceed the 1.21 minimum specified in PROJECT.md

## Test Results

| Module | Tests | Status |
|--------|-------|--------|
| `github.com/tylercrawford/drift` (library) | 203 | ✅ All pass |
| `github.com/tylercrawford/drift/cmd/drift` (CLI) | 20 | ✅ All pass |
| **Total** | **223** | **✅ Matches baseline** |

## Known Stubs

None — all functionality is fully wired.

## Self-Check: PASSED

**Files verified:**
- ✅ `go.work` exists at repo root
- ✅ `cmd/drift/go.mod` exists (`module github.com/tylercrawford/drift/cmd/drift`)
- ✅ `types.go`, `drift.go` and 11 other library files at repo root
- ✅ `drift/` directory removed

**Commits verified:**
- ✅ `1023f41` — feat(17-01): move library from drift/ to module root for canonical import path
- ✅ `ac099b4` — chore(17-01): create go.work workspace and cmd/drift/go.mod separate module
- ✅ `6f37390` — chore(17-01): verify all tests pass and add compiled binary to gitignore
