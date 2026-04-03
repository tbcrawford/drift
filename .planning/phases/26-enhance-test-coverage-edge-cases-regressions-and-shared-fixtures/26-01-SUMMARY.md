---
phase: 26
plan: "01"
subsystem: testing
tags: [tests, builder-api, algo, coverage]
dependency_graph:
  requires: []
  provides: [builder-api-test-coverage, apply-offset-test-coverage]
  affects: [builder.go, internal/algo/algo.go]
tech_stack:
  added: []
  patterns: [table-driven tests, white-box unit tests]
key_files:
  created:
    - internal/algo/algo_test.go
  modified:
    - builder_test.go
decisions:
  - Appended 9 TestBuilder_* functions rather than restructuring existing file to minimize merge risk
  - Used `package algo` (white-box) for algo_test.go to access unexported ApplyOffset directly
  - Followed existing test inputs (old/new Go snippet) for consistency
metrics:
  duration: "~15 minutes"
  completed: "2026-04-03"
  tasks_completed: 2
  files_changed: 2
---

# Phase 26 Plan 01: Builder API + ApplyOffset Tests Summary

**One-liner:** 9 Builder method tests and 5 ApplyOffset tests covering previously-uncovered zero/positive/zero-line-number paths.

## What Was Built

Added direct test coverage for the 9 previously-uncovered Builder API methods and the `ApplyOffset` internal helper.

### builder_test.go additions
9 new `TestBuilder_*` functions appended to the existing `builder_test.go`:
- `TestBuilder_Context` — zero context (ok) and negative context (error with "non-negative")
- `TestBuilder_Lang` — lang hint produces valid output
- `TestBuilder_ThemeResolved` — callback invoked during Render
- `TestBuilder_Split` — split mode renders without error
- `TestBuilder_LineNumbers` — LineNumbers(true) produces more output than LineNumbers(false)
- `TestBuilder_WithoutLineNumbers` — same output as LineNumbers(false)
- `TestBuilder_LineDiffStyle` — LineDiffStyle(false) succeeds
- `TestBuilder_WordDiff` — WordDiff(false) succeeds
- `TestBuilder_RenderWithNames` — output contains supplied file names

### internal/algo/algo_test.go (new)
5 `TestApplyOffset_*` tests as `package algo`:
- `TestApplyOffset_emptySlice` — nil slice, no panic
- `TestApplyOffset_positiveLinesAdjusted` — OldLine>0 gains oldOff, NewLine>0 gains newOff
- `TestApplyOffset_zeroLinesSkipped` — zero lines not adjusted (insert/delete markers)
- `TestApplyOffset_multipleEdits` — 3-edit slice all get correct offsets
- `TestApplyOffset_zeroOffsets` — both offsets zero, lines unchanged

## Test Count

| Stage | Count |
|-------|-------|
| Before 26-01 | 328 |
| After 26-01 | 342 |
| New tests added | +14 |

## Commits

| Hash | Message |
|------|---------|
| fecf2a6 | test(26-01): cover Builder API methods and ApplyOffset |

## Deviations from Plan

None — plan executed exactly as written.

## Self-Check: PASSED

- `builder_test.go` modified: confirmed
- `internal/algo/algo_test.go` created: confirmed
- Commit fecf2a6 exists: confirmed
- `go test ./...` passes (342 tests): confirmed
