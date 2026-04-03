---
phase: "25"
plan: "01"
subsystem: cmd/drift
tags: [parser, unified-diff, git-pager, tdd]
dependency_graph:
  requires: []
  provides: [parseUnifiedDiff, parsedFileDiff]
  affects: [cmd/drift/main.go]
tech_stack:
  added: []
  patterns: [state-machine, bufio.Scanner, strings.Builder]
key_files:
  created:
    - cmd/drift/unifieddiff.go
    - cmd/drift/unifieddiff_test.go
  modified: []
decisions:
  - "State machine with stateHeader/stateMeta/stateHunk using bufio.Scanner for line-by-line parsing"
  - "isNewFile/isDeletedFile booleans track null-side suppression — avoids writing content to empty placeholder side"
  - "flush() pattern finalizes and resets state for each file block"
  - "stripABPrefix removes a/ b/ git path prefixes; Name derived from NewName when available"
  - "No-newline-at-EOF marker strips trailing \\n from last appended line in both builders"
metrics:
  duration: "8min"
  completed: "2026-04-02"
  tasks_completed: 1
  files_changed: 2
---

# Phase 25 Plan 01: Unified Diff Parser for Git Pager Mode Summary

**One-liner:** State-machine unified diff parser reconstructs old/new file content from git diff streams, handling all format variants (added, deleted, binary, renamed, multi-hunk).

## What Was Built

`cmd/drift/unifieddiff.go` implements a `parseUnifiedDiff(r io.Reader) ([]parsedFileDiff, error)` function that reads a multi-file unified diff stream (as produced by `git diff`, `git show`, `git log -p`, etc.) and returns one `parsedFileDiff` per changed file. Each entry contains faithfully reconstructed old and new file content ready to feed into `drift.Diff`.

The parser uses a three-state machine (`stateHeader → stateMeta → stateHunk`) driven by `bufio.Scanner`. Context lines go to both old and new builders; `-` lines only to old; `+` lines only to new. Binary files are flagged via `IsBinary=true`.

## Decisions Made

- **State machine design**: Three states (`stateHeader/stateMeta/stateHunk`) cleanly separate "looking for next diff block", "parsing file headers", and "processing hunk content" — avoids complex nested if/else.
- **isNewFile/isDeletedFile flags**: Suppresses content reconstruction for the null side (`/dev/null`) — correctly produces empty OldContent for new files and empty NewContent for deleted files.
- **flush() pattern**: Finalizes and appends the current parsedFileDiff, then resets all buffers and flags — enables clean sequential multi-file processing.
- **Name derivation**: Strip `a/` or `b/` prefix from NewName when available, falling back to OldName for deleted files.
- **No-newline-at-EOF handling**: When `\` marker is seen, strips the trailing `\n` that was already appended to the builder — preserves exact file content.

## Tasks Completed

| # | Name | Commit | Files |
|---|------|--------|-------|
| 1 | Implement parsedFileDiff type + parseUnifiedDiff parser (TDD) | 90d5a2d | cmd/drift/unifieddiff.go, cmd/drift/unifieddiff_test.go |

## Verification Results

- `go test ./cmd/drift/ -run TestParseUnifiedDiff -v -count=1` — 11/11 subtests pass
- `go test ./...` — 316 tests pass, 0 regressions
- `go vet ./cmd/drift/` — no issues
- `parsedFileDiff` and `parseUnifiedDiff` are unexported (package-private) — correct

## Deviations from Plan

None — plan executed exactly as written.

## Known Stubs

None — all fields in parsedFileDiff are populated by the parser.

## Self-Check: PASSED

- [x] `cmd/drift/unifieddiff.go` exists
- [x] `cmd/drift/unifieddiff_test.go` exists
- [x] Commit 90d5a2d exists
