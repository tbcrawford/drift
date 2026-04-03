---
phase: 26
plan: "03"
subsystem: testing
tags: [tests, cli, coverage, parseAlgorithm, fileHeaderName, exitCodeErr, streamThroughPager, resolveInputs, gitworktree]
dependency_graph:
  requires: []
  provides: [cli-function-test-coverage]
  affects: [cmd/drift/main.go, cmd/drift/exit.go, cmd/drift/input.go, cmd/drift/gitworktree.go]
tech_stack:
  added: []
  patterns: [package-main white-box tests, git repo fixture pattern]
key_files:
  created: []
  modified:
    - cmd/drift/main_test.go
    - cmd/drift/input_test.go
    - cmd/drift/gitworktree_test.go
decisions:
  - Added `drift` and `fmt` imports to main_test.go to support parseAlgorithm and streamThroughPager tests
  - Used `noPager: true` in streamThroughPager tests to exercise the buffer/non-TTY code path without a real terminal
  - Used makeTestRepo fixture helper (already in gitworktree_test.go) for gitShowHEADBlobFromTree tests
  - Added compile-time `var _ = (*object.Tree)(nil)` assertion to prevent unused import error in gitworktree_test.go
metrics:
  duration: "~25 minutes"
  completed: "2026-04-03"
  tasks_completed: 1
  files_changed: 3
---

# Phase 26 Plan 03: CLI Function Tests Summary

**One-liner:** Direct tests for parseAlgorithm (all 5 paths), fileHeaderName (9 cases), exitCodeErr.Error, streamThroughPager (3 paths), and resolveInputs/gitShowHEADBlobFromTree uncovered branches.

## What Was Built

### cmd/drift/main_test.go additions
Added `fmt` and `github.com/tbcrawford/drift` imports, then appended:

**parseAlgorithm tests:**
- `TestParseAlgorithm_validCases` — table-driven: all 4 valid values (auto/myers/patience/histogram), case-insensitive, trimmed
- `TestParseAlgorithm_invalid` — "bogus" returns *exitCodeErr with code=2

**exitCodeErr.Error tests:**
- `TestExitCodeErr_Error` — Error() returns the msg field
- `TestExitCodeErr_EmptyMsg` — empty msg returns ""

**fileHeaderName tests (9 cases):**
- `TestFileHeaderName_noArgs` — 0 args → ""
- `TestFileHeaderName_oneArgStdin` — ["-"] → ""
- `TestFileHeaderName_oneArgPath` — ["foo/bar.go"] → "foo/bar.go"
- `TestFileHeaderName_twoArgsBothStdin` — ["-","-"] → ""
- `TestFileHeaderName_twoArgsFirstStdin` — ["-","new.go"] → "new.go"
- `TestFileHeaderName_twoArgsSecondStdin` — ["old.go","-"] → "old.go"
- `TestFileHeaderName_twoArgsSame` — ["same.go","same.go"] → "same.go"
- `TestFileHeaderName_twoArgsDifferent` — ["old.go","new.go"] → "old.go → new.go"
- `TestFileHeaderName_threeArgs` — 3 args → "" (default case)

**streamThroughPager tests (noPager=true path):**
- `TestStreamThroughPager_noDiff` — renderFn returns false → nil error
- `TestStreamThroughPager_hasDiff` — renderFn returns true → *exitCodeErr code=1
- `TestStreamThroughPager_renderError` — renderFn returns error → *exitCodeErr code=2

### cmd/drift/input_test.go additions
- `TestResolveInputs_fileAndStdin` — b=="-" case: file as old, stdin as new; verifies content and names
- `TestResolveInputs_tooManyArgs` — 3 positional args → error containing "too many"

### cmd/drift/gitworktree_test.go additions
- `TestGitShowHEADBlobFromTree_nilTree` — nil tree → ("", nil)
- `TestGitShowHEADBlobFromTree_existingFile` — committed file → correct content
- `TestGitShowHEADBlobFromTree_notFound` — missing file → ("", nil) via ErrFileNotFound

## Test Count

| Stage | Count |
|-------|-------|
| Before 26-03 | 354 |
| After 26-03 | 375 |
| New tests added | +21 |

## Commits

| Hash | Message |
|------|---------|
| 9f8c986 | test(26-03): add CLI function tests for parseAlgorithm, fileHeaderName, exitCodeErr, streamThroughPager, resolveInputs, and gitShowHEADBlobFromTree |

## Deviations from Plan

None — plan executed exactly as written.

## Known Stubs

None.

## Self-Check: PASSED

- `cmd/drift/main_test.go` modified: confirmed
- `cmd/drift/input_test.go` modified: confirmed
- `cmd/drift/gitworktree_test.go` modified: confirmed
- Commit 9f8c986 exists: confirmed
- `go test ./...` passes (375 tests): confirmed
