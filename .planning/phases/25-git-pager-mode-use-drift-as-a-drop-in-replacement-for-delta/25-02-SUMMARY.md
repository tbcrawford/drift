---
phase: "25"
plan: "02"
subsystem: cmd/drift
tags: [pager, git-pager, color-only, stdin-pipe, install-pager, tdd]
dependency_graph:
  requires: [25-01]
  provides: [runPagerMode, isStdinPipe, runColorOnlyMode, drift-install-pager]
  affects: [cmd/drift/main.go, cmd/drift/flags.go]
tech_stack:
  added: []
  patterns: [isTTYReader interface, ttyReader test helper, bufio.Scanner, cobra subcommand]
key_files:
  created: []
  modified:
    - cmd/drift/main.go
    - cmd/drift/flags.go
    - cmd/drift/main_test.go
decisions:
  - "isTTYReader interface injected into isStdinPipe for testable TTY detection without real *os.File"
  - "ttyReader test helper implements isTTYReader to signal TTY stdin to zero-arg tests"
  - "runPagerMode dispatched FIRST in runRoot before zero-arg git worktree branch (non-TTY + no args = pager mode)"
  - "colorOnly flag stored on rootOptions; runColorOnlyMode preserves exact line structure for git add -p"
  - "install-pager subcommand registered with cmd.AddCommand on root cobra command"
  - "Binary files in pager mode are skipped silently (no panic, no output)"
metrics:
  duration: "8min"
  completed: "2026-04-02"
  tasks_completed: 2
  files_changed: 3
---

# Phase 25 Plan 02: Git Pager Mode Wiring Summary

**One-liner:** Wired git pager mode so `drift` acts as a drop-in for `delta` — stdin pipe detection, per-file re-rendering, `--color-only` for `interactive.diffFilter`, and `drift install-pager` subcommand.

## What Was Built

### Task 1: Stdin Pipe Detection + runPagerMode
- `isStdinPipe(in io.Reader)` — detects whether stdin is a pipe vs TTY. Uses `isTTYReader` interface for testability (non-`*os.File` readers are treated as pipes; readers implementing `IsTTY() bool` are respected).
- `runPagerMode(r io.Reader, opts *rootOptions)` — parses unified diff from stdin via `parseUnifiedDiff`, re-renders each file with drift's full styling pipeline (file headers, syntax highlighting, gutter line numbers), then streams through pager.
- `runRoot` updated with pager mode detection at top: zero-arg + non-TTY stdin → `runPagerMode`. Zero-arg + TTY stdin → existing git worktree mode.
- `ttyReader` test helper wraps an `io.Reader` and implements `IsTTY() bool` so existing zero-arg tests work with TTY-simulated stdin.

### Task 2: --color-only + drift install-pager  
- `--color-only` flag added to `rootFlags`/`rootOptions`/`newRootCmd`. When set with piped stdin, calls `runColorOnlyMode`.
- `runColorOnlyMode` — scans stdin line-by-line, colorizes `+` lines green and `-` lines red (ANSI SGR), passes context lines through unchanged. Preserves exact line structure for `git add -p` compatibility.
- `colorizeUnifiedLine(line, noColor)` — pure function for per-line colorization; `--no-color` pass-through.
- `drift install-pager` cobra subcommand prints a ready-to-paste `~/.gitconfig` snippet with `core.pager = drift` and `interactive.diffFilter = drift --color-only`.

## Decisions Made

- **isTTYReader interface**: Checked in `isStdinPipe` before `*os.File` cast. Lets tests inject TTY-flagged readers without OS pipe setup. Clean separation: production path uses `*os.File` + `term.IsTerminal`; tests use `ttyReader`.
- **Pager mode order in runRoot**: Pager mode check comes BEFORE zero-arg git worktree check — both are zero-arg + no-from/to, so the stdin pipe test disambiguates them. TTY stdin → git worktree. Non-TTY stdin → pager mode.
- **Binary file skip**: `IsBinary=true` → `continue` in `runPagerMode` render loop. No output, no panic. Exit 0 if all files are binary.
- **colorOnly + isStdinPipe**: Both checked before any other routing in `runRoot`. `--color-only` + pipe → `runColorOnlyMode`. This comes before pager mode detection.

## Tasks Completed

| # | Name | Commit | Files |
|---|------|--------|-------|
| 1 | Stdin pipe detection + runPagerMode (TDD) | 5655bcc | cmd/drift/main.go, cmd/drift/flags.go, cmd/drift/main_test.go |
| 2 | --color-only flag + drift install-pager subcommand | 5655bcc | cmd/drift/main.go, cmd/drift/flags.go, cmd/drift/main_test.go |

## Verification Results

- `go test ./...` — 324 tests pass, 0 regressions
- `go vet ./...` — no issues
- `go build ./cmd/drift/` — success
- Manual: `git diff HEAD~1..HEAD | drift --no-color` → 508 lines of drift-rendered output with `▸ filename` headers, gutter line numbers, `@@` hunk markers; exit 1
- `drift install-pager` → prints `[core] pager = drift` + `[interactive] diffFilter = drift --color-only`; exit 0
- `drift --color-only < /dev/null` → exit 0 (empty input, no error)

## Deviations from Plan

**[Rule 1 - Bug] ttyReader test helper added for zero-arg test regression fix**
- **Found during:** Task 1 GREEN phase — 2 existing zero-arg tests regressed
- **Issue:** `isStdinPipe` returns `true` for all non-`*os.File` readers (including `strings.NewReader`), causing `TestRunCLI_zeroArg_notInRepo` and `TestRunCLI_zeroArg_hasDiff` to trigger pager mode instead of git worktree mode.
- **Fix:** Added `isTTYReader` interface to `isStdinPipe`; added `ttyReader` test helper implementing `IsTTY() bool = true`; updated 4 zero-arg tests to use `ttyStream()`.
- **Files modified:** cmd/drift/main.go, cmd/drift/main_test.go

## Known Stubs

None — all functionality fully wired.

## Self-Check: PASSED

- [x] `cmd/drift/main.go` contains `runPagerMode`, `isStdinPipe`, `runColorOnlyMode`
- [x] `cmd/drift/flags.go` contains `colorOnly` in both `rootFlags` and `rootOptions`  
- [x] `cmd/drift/main_test.go` contains `TestRunCLI_pagerMode_*`, `TestRunCLI_colorOnly`, `TestRunCLI_installPager`
- [x] Commit 5655bcc exists
- [x] 324 tests pass
