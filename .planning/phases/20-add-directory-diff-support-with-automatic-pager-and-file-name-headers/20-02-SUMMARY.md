---
phase: 20-add-directory-diff-support-with-automatic-pager-and-file-name-headers
plan: "02"
subsystem: cmd/drift
tags: [directory-diff, cli, tdd, pager]
dependency_graph:
  requires: [filePair, diffDirectories]
  provides: [runDirectoryDiff, isDir]
  affects: [cmd/drift/main.go, cmd/drift/input.go, cmd/drift/main_test.go]
tech_stack:
  added: []
  patterns: [os.ReadFile, drift.Diff, drift.RenderWithNames, bytes.Buffer accumulation, pager routing]
key_files:
  created: []
  modified:
    - cmd/drift/input.go
    - cmd/drift/main.go
    - cmd/drift/main_test.go
decisions:
  - "runDirectoryDiff is a separate function from runRoot for single-responsibility; runRoot dispatches by detecting both args are dirs via isDir()"
  - "Directory diff shares pager routing with single-file path by accumulating into bytes.Buffer then counting lines"
  - "isDir placed in input.go (not main.go) as it's an input resolution helper adjacent to resolveInputs"
  - "Test 7 (--from/--to incompatible) tests --from without --to (the clearest incompatible case)"
metrics:
  duration_seconds: 168
  completed_date: "2026-04-02"
  tasks_completed: 2
  files_created: 0
  files_modified: 3
requirements_validated:
  - DIR-01
  - DIR-02
  - DIR-03
---

# Phase 20 Plan 02: CLI Directory Diff Wiring — Summary

**One-liner:** `runDirectoryDiff` wired into `runRoot` with directory detection via `isDir`, per-file `=== name ===` headers, and automatic pager routing through existing Phase 19 buffer accumulation pattern.

## What Was Built

Extended `cmd/drift/main.go` and `cmd/drift/input.go` to deliver the user-visible directory diff feature:

- **`isDir(path string) bool`** in `cmd/drift/input.go` — uses `os.Stat` to detect whether a path is a directory; used by `runRoot` as the dispatch guard.
- **`runDirectoryDiff(pairs []filePair, opts *rootOptions, buf *bytes.Buffer) (hasDiff bool, err error)`** in `cmd/drift/main.go` — iterates each `filePair`, reads file content (`os.ReadFile`), calls `drift.Diff` and `drift.RenderWithNames`, writes `=== name ===` header before each diff; skips identical files; handles added (empty old) and removed (empty new) files.
- **`runRoot` extended** — directory detection block at top: `len(opts.args)==2 && isDir(args[0]) && isDir(args[1])` dispatches to `diffDirectories` + `runDirectoryDiff`; routes output through existing pager logic (buffer → line count → shouldPage → pager or direct write); returns `exitCode(1)` on diff, `nil` on identical dirs.
- **7 integration tests** in `cmd/drift/main_test.go` covering all behavior cases via TDD.

## Tasks Completed

| Task | Name | Commits | Files |
|------|------|---------|-------|
| 1 | Add isDir helper | `5da5b92` | `cmd/drift/input.go` |
| 2 (RED) | Failing tests for directory diff CLI | `6a05613` | `cmd/drift/main_test.go` |
| 2 (GREEN) | Implement runDirectoryDiff + wire into runRoot | `d8c42ef` | `cmd/drift/main.go` |

## Verification

```
go test ./cmd/drift/ -run "TestRunCLI_directory" -v  → 7 passed
go test ./...                                          → 270 passed (no regressions)
```

All 7 behavior cases pass:
1. Identical dirs → exit 0, empty stdout
2. File differs → exit 1, `=== a.txt ===` header, `@@` hunk
3. File added in new dir → exit 1, `=== added.txt ===` header, `+new content` line
4. File removed in new dir → exit 1, `=== removed.txt ===` header, `-old content` line
5. `--no-color` suppresses ANSI in directory diff output
6. Non-directory paths use existing two-file diff path (no regression)
7. `--from` without `--to` returns exit 2 (incompatible flags error)

## Deviations from Plan

None — plan executed exactly as written.

## Known Stubs

None.

## Self-Check

- [x] `cmd/drift/input.go` modified with `isDir` helper
- [x] `cmd/drift/main.go` modified with `runDirectoryDiff` and `runRoot` extension
- [x] `cmd/drift/main_test.go` modified with 7 directory diff tests
- [x] Commit `5da5b92` exists (isDir helper)
- [x] Commit `6a05613` exists (RED tests)
- [x] Commit `d8c42ef` exists (GREEN implementation)
- [x] `go test ./...` exits 0 with 270 tests passing
- [x] No new external dependencies added

## Self-Check: PASSED
