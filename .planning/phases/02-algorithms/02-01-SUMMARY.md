---
phase: 02-algorithms
plan: 02-01
subsystem: algorithms
tags: [patience-diff, myers-fallback, go, diff-algorithm, iterative-stack]

requires:
  - phase: 01-foundation
    provides: Myers diff implementation (algo.Differ interface, edittype package)

provides:
  - internal/algo/patience package implementing algo.Differ
  - Patience diff algorithm with iterative stack and Myers fallback
  - 8-test suite covering invariants, fallback, round-trip, and superiority cases

affects: [02-03, 02-04, drift.go WithAlgorithm dispatch]

tech-stack:
  added: []
  patterns:
    - "Iterative stack with tagged union (isEmit bool) for interleaving emit and frame items"
    - "Frequency map unique-element detection: lines appearing exactly once in both ranges"
    - "O(N·M) LCS DP over unique-line slices for anchor computation"
    - "applyOffset() pattern for converting sub-range Myers line numbers to full-file 1-indexed"

key-files:
  created:
    - internal/algo/patience/patience.go
    - internal/algo/patience/patience_test.go
  modified: []

key-decisions:
  - "Iterative stack with stackItem tagged union (isEmit/frame) avoids Go recursion depth issues and correctly interleaves anchor Equal edits with gap frame processing"
  - "Push tail edits first (reversed), then gap/anchor pairs reversed — ensures left-to-right pop order without any post-sort step"
  - "Myers called on sub-slices directly (old[f.os:f.oe]) with applyOffset() applied to results, consistent with PLAN spec"

patterns-established:
  - "Stack-based patience: push in reverse (right gap first, left gap last) so left is processed first"
  - "applyEditsWithNew helper in tests takes both old and new slices for correct Insert line reconstruction"

requirements-completed: [ALGO-01, ALGO-04]

duration: 7min
completed: 2026-03-25
---

# Phase 02 Plan 01: Patience Diff Algorithm Summary

**Patience diff implemented as `internal/algo/patience` package with iterative stack, O(N·M) LCS anchor computation, and Myers fallback for sub-ranges with no unique common lines**

## Performance

- **Duration:** 7 min
- **Started:** 2026-03-25T20:00:00Z
- **Completed:** 2026-03-25T20:07:24Z
- **Tasks:** 2
- **Files modified:** 2

## Accomplishments

- `Patience` struct implementing `algo.Differ` with compile-time interface check
- Iterative stack using a tagged-union `stackItem` (isEmit/frame) to correctly order anchor Equal edits between gap frames
- Head/tail trim, frequency-map unique-element detection, O(N·M) LCS DP for anchors
- Myers fallback for sub-ranges with no unique common lines, with `applyOffset()` correcting sub-range line numbers
- 8 tests all passing with `-race`: edge cases, invariants, fallback, round-trip, and canonical C function-move example

## Task Commits

Each task was committed atomically:

1. **Task 02-01-01: Implement patience.go** - `dd07c11` (feat)
2. **Task 02-01-02: Write patience_test.go** - `010b2a0` (test)

## Files Created/Modified

- `internal/algo/patience/patience.go` — Patience struct, Diff(), lcsAnchors(), applyOffset(), buildFreqMap()
- `internal/algo/patience/patience_test.go` — 8-test suite: TestBothEmpty, TestOldEmptyAllInserts, TestNewEmptyAllDeletes, TestIdenticalInputs, TestLineInvariant_Patience, TestPatienceFallback_NoUniqueLines, TestRoundTrip_Patience, TestPatienceSuperiority_FunctionMove

## Decisions Made

- Used tagged-union `stackItem` struct (isEmit bool + edit/frame fields) to allow pre-built Equal edits for anchors to be pushed onto the same stack as gap frames. This avoids any post-sort step and guarantees correct forward order.
- Pushed tail edits first (in reverse) so they sit at the bottom of the deferred items, emitted last.
- Gap/anchor pairs pushed in reverse (gap[n], anchor[n-1], gap[n-1], ..., anchor[0], gap[0]) so popping yields left-to-right order.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Type mismatch between local `lineIdx` and `lcsAnchors` anonymous struct parameter**

- **Found during:** Task 1 (go build)
- **Issue:** `lcsAnchors()` was declared with anonymous `[]struct{line string; idx int}` parameters but called with named `lineIdx` type — Go doesn't unify these
- **Fix:** Promoted `lineIdx` to a package-level named type and updated `lcsAnchors` signature accordingly
- **Files modified:** internal/algo/patience/patience.go
- **Verification:** `go build ./internal/algo/patience/` exits 0
- **Committed in:** dd07c11

---

**Total deviations:** 1 auto-fixed (1 bug)
**Impact on plan:** Minor compile fix; no scope change.

## Issues Encountered

None — all tests passed on first run after the type fix.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Patience package complete; `drift.go` can now import `internal/algo/patience` and wire `case Patience: differ = patience.New()`
- 02-02 (Histogram) already completed by parallel agent
- Ready for 02-03 (WithAlgorithm wiring) and 02-04 (property tests)

---
*Phase: 02-algorithms*
*Completed: 2026-03-25*
