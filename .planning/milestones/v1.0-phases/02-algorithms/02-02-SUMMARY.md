---
phase: 02-algorithms
plan: "02"
subsystem: testing

# Dependency graph
requires:
  - phase: 02-algorithms/02-01
    provides: algo.Differ interface, Myers fallback pattern, iterative stack approach
provides:
  - Histogram diff algorithm satisfying algo.Differ
  - histogramMaxOccurrences=64 cutoff constant
  - Myers fallback for regions where all lines appear >64 times
  - Comprehensive test suite covering edge cases, fallback, round-trip, and line invariant
affects: [02-03-patience, 02-04-integration, 03-rendering, root-drift-package]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - Iterative stack-based recursive diff with LIFO push order
    - findBestMatch() anchors on lowest-occurrence-count old lines
    - Sort edits by new-file position after stack unwind to guarantee file order
    - Myers fallback fires when no histogram match found (all lines > 64 occurrences)

key-files:
  created:
    - internal/algo/histogram/histogram.go
    - internal/algo/histogram/histogram_test.go
  modified: []

key-decisions:
  - "histogramMaxOccurrences=64 — matches jgit/Git implementation (research confirmed)"
  - "sortEdits() by NewLine (primary) post-stack-unwind to fix out-of-order Equal emission"
  - "Inserts sort after Equal/Delete at same new-file position via tiebreak"
  - "Myers fallback invoked when findBestMatch returns found=false (no match within threshold)"

patterns-established:
  - "sortEdits pattern: use NewLine as primary sort key; OldLine for Deletes (NewLine=0)"
  - "applyOffset pattern: adjusts Myers sub-region line numbers back to full-file coordinates"
  - "histogram_test applyEditsWithNew: applies edits using both old and new slices for round-trip verification"

requirements-completed: [ALGO-02, ALGO-04]

# Metrics
duration: 20min
completed: 2026-03-25
---

# Plan 02-02: Histogram Diff Algorithm Summary

**Histogram diff in `internal/algo/histogram/` with jgit-style findBestMatch, Myers fallback at 64-occurrence cutoff, and full test suite**

## Performance

- **Duration:** ~20 min
- **Started:** 2026-03-25T20:10:00Z
- **Completed:** 2026-03-25T20:30:00Z
- **Tasks:** 2
- **Files modified:** 2

## Accomplishments
- Implemented `internal/algo/histogram/histogram.go` satisfying `algo.Differ` with compile-time check
- `findBestMatch()` follows jgit HistogramDiff: anchors on unique/low-frequency old lines, widens contiguous matches
- Myers fallback fires correctly when all old lines in a region exceed `histogramMaxOccurrences=64`
- Full test suite: edge cases, line invariant (8 table cases), Myers fallback with 200-line identical input, round-trip, no-panic pathological test

## Task Commits

1. **Task 02-02-01: histogram.go implementation** — `5e09afe` (feat)
2. **Task 02-02-02: histogram_test.go** — `2069e15` (test)

## Files Created/Modified
- `internal/algo/histogram/histogram.go` — Histogram struct, Diff(), findBestMatch(), applyOffset(), sortEdits()
- `internal/algo/histogram/histogram_test.go` — 8 test functions covering all acceptance criteria

## Decisions Made
- **sortEdits() required**: The iterative stack emits Equal edits for a match before processing the before-region, so edits arrive out of order. Added `sortEdits()` using NewLine as primary key to restore canonical file order.
- **Sort key**: Use `NewLine` for Equal/Insert; use `OldLine` for Delete (NewLine=0). This correctly places Inserts before the Equal that maps to the next new-file position.

## Deviations from Plan

### Auto-fixed Issues

**1. [Ordering] Edit sort key uses NewLine (not OldLine) as primary**
- **Found during:** Task 02-02-02 (TestRoundTrip_Histogram failure)
- **Issue:** Emitting Equal edits immediately when a match is found, before sub-regions are processed, produces out-of-order edits. Initial sort by OldLine still wrong for Insert-before-Equal cases.
- **Fix:** `sortEdits()` uses `NewLine` as primary sort key (present for Equal/Insert); `OldLine` as fallback for Delete. Insert tiebreaks after Equal/Delete at same position.
- **Verification:** `TestRoundTrip_Histogram` passes; all 8 TestLineInvariant subtests pass.
- **Committed in:** `5e09afe` (histogram.go)

---

**Total deviations:** 1 auto-fixed (sort key selection)
**Impact on plan:** Essential for correctness. No scope creep.

## Issues Encountered
- Out-of-order edits from iterative stack required post-processing sort — resolved by switching primary sort key from OldLine to NewLine.

## Next Phase Readiness
- Histogram satisfies `algo.Differ` — ready for integration in the algorithm dispatcher (plan 02-04)
- Myers and Histogram both implemented; Patience (02-03) is the remaining algorithm

---
*Phase: 02-algorithms*
*Completed: 2026-03-25*
