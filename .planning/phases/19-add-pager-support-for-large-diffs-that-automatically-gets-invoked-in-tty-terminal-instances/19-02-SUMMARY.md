---
phase: 19-add-pager-support-for-large-diffs-that-automatically-gets-invoked-in-tty-terminal-instances
plan: "02"
subsystem: cli/pager
tags: [pager, tty, cli, flags, runRoot, buffer-render]
dependency_graph:
  requires: [cmd/drift/pager.go]
  provides: [--no-pager flag, pager wiring in runRoot]
  affects: [cmd/drift/flags.go, cmd/drift/main.go, cmd/drift/main_test.go]
tech_stack:
  added: []
  patterns: [TDD red-green, buffer-first render, pager fallback on subprocess error]
key_files:
  created: []
  modified:
    - cmd/drift/flags.go
    - cmd/drift/main.go
    - cmd/drift/main_test.go
decisions:
  - "runRoot renders to bytes.Buffer before deciding pager vs direct write — enables line counting without consuming the output"
  - "term.GetSize used for terminal height; returns 0 on non-TTY so shouldPage returns false automatically"
  - "startPager failure falls back to direct write — pager subprocess failure is non-fatal"
metrics:
  duration: "147s"
  completed_date: "2026-04-01"
  tasks_completed: 2
  tasks_total: 2
  files_created: 0
  files_modified: 3
---

# Phase 19 Plan 02: Pager CLI Wiring Summary

**One-liner:** `--no-pager` flag added and pager primitives wired into `runRoot` via buffer-first rendering with automatic TTY detection and line count threshold.

## What Was Built

### Task 1: --no-pager flag in flags.go

- `noPager bool` field added to `rootFlags` struct
- `noPager bool` field added to `rootOptions` struct
- `resolveRootOptions` passes `flags.noPager` through to `rootOptions.noPager`
- `--no-pager` flag registered in `newRootCmd` (after `--show-theme`)

### Task 2: Pager wiring in runRoot

`runRoot` in `cmd/drift/main.go` rewritten to:

1. Resolve inputs (same as before)
2. Call `drift.Diff(...)` (same as before)
3. Return nil if `result.IsEqual` (same as before)
4. **Render to `bytes.Buffer`** (not directly to `streams.Out`)
5. **Count lines**: `strings.Count(buf.String(), "\n")`
6. **Get terminal height**: `term.GetSize(f.Fd())` when `streams.Out` is `*os.File`; else `termHeight=0`
7. **Call `shouldPage()`**: returns false for non-TTY, `noPager=true`, or line count ≤ terminal height
8. **If shouldPage=true**: call `startPager(resolvePager(), streams)` → write buffer → cleanup; fallback to direct write on subprocess error
9. **If shouldPage=false**: write buffer directly to `streams.Out`

### Tests Added

- `TestHelpListsAllFlags` updated to include `"--no-pager"` in expected flags list
- `TestRunCLI_noPagerFlag`: verifies exit 1 and `"@@"` hunk header with `--no-pager` flag
- `TestRunCLI_pagerSkippedOnNonTTY`: verifies output arrives in buffer (non-TTY path never invokes pager)

## TDD Flow

- **RED:** Test file updated with `--no-pager` expectations and two new test functions; all tests pass trivially because flag was already registered (Task 1 prerequisite)
- **GREEN:** `runRoot` rewritten with buffer-first render path, line counting, and `shouldPage`/`startPager` calls — all 254 tests pass

## Commits

| Hash | Type | Description |
|------|------|-------------|
| `1de7d4a` | feat | Add --no-pager flag to rootFlags/rootOptions |
| `d4ec5a4` | test | Add failing tests for --no-pager flag and non-TTY pager skip |
| `2f9984b` | feat | Wire pager into runRoot — buffer render, line count, shouldPage/startPager |

## Deviations from Plan

None — plan executed exactly as written.

## Self-Check

- [x] `cmd/drift/flags.go` — `noPager` field in `rootFlags` and `rootOptions`, wired through `resolveRootOptions`
- [x] `cmd/drift/main.go` — `runRoot` renders to buffer, calls `shouldPage`, invokes pager or writes direct
- [x] `cmd/drift/main_test.go` — `TestHelpListsAllFlags` includes `--no-pager`, two new tests added
- [x] `go build ./cmd/drift/` — exits 0
- [x] `go test ./... -count=1` — 254 tests pass (no regressions)
- [x] `go vet ./cmd/drift/` — no issues
- [x] `drift --help` — includes `--no-pager` flag
- [x] Commits `1de7d4a`, `d4ec5a4`, `2f9984b` in git log
