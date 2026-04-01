---
phase: 19-add-pager-support-for-large-diffs-that-automatically-gets-invoked-in-tty-terminal-instances
plan: "01"
subsystem: cli/pager
tags: [pager, tty, cli, iostreams, exec]
dependency_graph:
  requires: []
  provides: [cmd/drift/pager.go]
  affects: [cmd/drift]
tech_stack:
  added: []
  patterns: [TDD red-green, io.Pipe for subprocess stdin, term.IsTerminal TTY detection]
key_files:
  created:
    - cmd/drift/pager.go
    - cmd/drift/pager_test.go
  modified: []
decisions:
  - "io.Pipe used for pager stdin rather than os.Pipe — avoids fd leaks and integrates cleanly with exec.Command.Stdin"
  - "term.IsTerminal (charmbracelet/x/term) reused for TTY detection — already a direct dep from termwidth.go"
  - "shouldPage tests only cover false-path conditions — true-path requires real TTY unavailable in tests (documented in plan)"
metrics:
  duration: "61s"
  completed_date: "2026-04-01"
  tasks_completed: 1
  tasks_total: 1
  files_created: 2
  files_modified: 0
---

# Phase 19 Plan 01: Pager Primitives Summary

**One-liner:** Isolated pager primitives (`resolvePager`, `shouldPage`, `startPager`) in `cmd/drift/pager.go` with full unit test coverage using `cat` as fake pager subprocess.

## What Was Built

Three pure, injectable functions in `cmd/drift/pager.go`:

- **`resolvePager() string`** — resolves the pager command: `$PAGER` env → `less -R` (if `less` on PATH) → `more`
- **`shouldPage(out io.Writer, lineCount int, termHeight int, noPager bool) bool`** — returns `true` only when `out` is a TTY `*os.File`, `noPager=false`, `termHeight > 0`, and `lineCount > termHeight`
- **`startPager(pagerCmd string, streams IOStreams) (io.WriteCloser, func(), error)`** — launches the pager subprocess via `exec.Command` with `io.Pipe` for stdin; pager stdout/stderr wired to `IOStreams`

Unit tests in `cmd/drift/pager_test.go` cover:
- `TestPagerResolvePager`: `$PAGER` env override, empty `$PAGER`, fallback to `less -R`/`more`
- `TestPagerShouldPage`: `noPager=true`, non-`*os.File` writer, `lineCount <= termHeight`, `termHeight=0`, non-TTY `*os.File`
- `TestPagerStart`: `cat` as fake pager, writes bytes to `WriteCloser`, calls cleanup, verifies output in buffer

## TDD Flow

- **RED:** `pager_test.go` written first — build failed with 10 `undefined` errors
- **GREEN:** `pager.go` implemented — all 12 pager tests pass; full suite 32/32 pass; `go vet` clean

## Commits

| Hash | Type | Description |
|------|------|-------------|
| `165250a` | test | Add failing tests for pager resolution, shouldPage, startPager |
| `71450b3` | feat | Implement pager primitives in cmd/drift/pager.go |

## Deviations from Plan

None — plan executed exactly as written.

## Self-Check

- [x] `cmd/drift/pager.go` created
- [x] `cmd/drift/pager_test.go` created
- [x] `go test ./cmd/drift/ -run TestPager` — 12 tests pass
- [x] `go test ./cmd/drift/` — 32 tests pass (no regressions)
- [x] `go vet ./cmd/drift/` — no issues
- [x] Commits `165250a` and `71450b3` in git log
