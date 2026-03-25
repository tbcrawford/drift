---
phase: 02-algorithms
plan: "02-04"
subsystem: testing
tags: [rapid, property-based-testing, histogram, patience, myers, diff]

# Dependency graph
requires:
  - phase: 02-algorithms/02-03
    provides: patience.New() and histogram.New() wired into drift.go dispatch
provides:
  - TestProperty_RoundTrip_Patience — 1000-run rapid property check for Patience algorithm
  - TestProperty_RoundTrip_Histogram — 1000-run rapid property check for Histogram algorithm
  - Histogram correctness fix: tagged stackItem union eliminates broken post-sort
affects: [phase-03-rendering]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - Tagged stackItem union (isEmit bool) for LIFO stacks that emit edits in file-traversal order — avoids post-sort correctness issues with Delete OldLine proxies

key-files:
  created: []
  modified:
    - drift_property_test.go
    - internal/algo/histogram/histogram.go

key-decisions:
  - "Histogram.Diff rewired to tagged stackItem union (same pattern as Patience) — eliminates post-sort entirely; OldLine-proxy sort key is fundamentally uncomputable without a full traversal pass, so sort is wrong by design"
  - "Delete-before-Equal tiebreak is insufficient fix — the correct fix is ordered emission via stack, not sort adjustment"

patterns-established:
  - "LIFO stacks that split work into before/match/after sub-regions MUST use tagged isEmit entries to preserve file-traversal emission order — never emit the match block eagerly then sort"

requirements-completed: [ALGO-01, ALGO-02, ALGO-03, ALGO-04]

# Metrics
duration: 15min
completed: 2026-03-25
---

# Plan 02-04: Extend property-based tests — all three algorithms satisfy `apply(diff(a, b), a) == b`

**1000-run rapid property checks for Patience and Histogram plus histogram correctness fix eliminating a broken post-sort**

## Performance

- **Duration:** ~15 min
- **Started:** 2026-03-25T16:10:00Z
- **Completed:** 2026-03-25T16:25:00Z
- **Tasks:** 2
- **Files modified:** 2

## Accomplishments
- Added `TestProperty_RoundTrip_Patience` and `TestProperty_RoundTrip_Histogram` to `drift_property_test.go`
- Fixed a histogram correctness bug: LIFO stack was emitting Equal edits immediately before pushing sub-regions, then relying on a post-sort that incorrectly used OldLine as a proxy for Delete's new-file position
- All 5 property tests pass with `-race -count=1`; full suite `go test ./... -race -count=1` exits 0

## Task Commits

Each task was committed atomically:

1. **Task 02-04-01: Add TestProperty_RoundTrip_Patience and TestProperty_RoundTrip_Histogram** - `d5795cf` (feat + fix: also rewrote histogram to use tagged stack)
2. **Task 02-04-02: Final phase validation** - `b63cbad` (chore: validation run)

## Files Created/Modified
- `drift_property_test.go` — appended two new property test functions
- `internal/algo/histogram/histogram.go` — rewrote Diff to use tagged `stackItem` union (isEmit bool), removing the broken `sortEdits`/`sortKey`/`editTieBreak` functions

## Decisions Made
- Histogram's post-sort approach is fundamentally broken: `OldLine` cannot serve as a proxy for a Delete's new-file position because that position depends on the full sequence of preceding Inserts and Equals. Fix: use the same `isEmit`-tagged LIFO stack pattern as Patience, emitting edits in file-traversal order without any post-sort.
- Initial attempt to fix the tiebreak (Delete=0, Equal=1, Insert=2) fixed one failure case but introduced a different one — confirmed the post-sort approach cannot be patched.

## Deviations from Plan

### Auto-fixed Issues

**1. [Correctness Bug] Histogram post-sort produces wrong edit order**
- **Found during:** Task 02-04-01 (running TestProperty_RoundTrip_Histogram)
- **Issue:** Histogram emitted Equal for matched block before pushing before/after sub-regions onto stack. Post-sort used OldLine as proxy for Delete's sort key. When Delete.OldLine == Equal.NewLine they tied; any tiebreak ordering was wrong for some input combination.
- **Fix:** Replaced `frame` struct + separate `[]edittype.Edit` append + `sortEdits` with tagged `stackItem` struct (isEmit bool). Equal block entries are now pushed in reverse onto the stack as `isEmit=true` items, after the after-region and before the before-region, so they pop in correct file order.
- **Files modified:** `internal/algo/histogram/histogram.go`
- **Verification:** `go test -run TestProperty_RoundTrip_Histogram ./... -race -count=1` exits 0 after 1000 iterations
- **Committed in:** `d5795cf` (Task 02-04-01 commit)

---

**Total deviations:** 1 auto-fixed (correctness bug in histogram.go)
**Impact on plan:** Required fix — property test correctly caught the bug. No scope creep.

## Issues Encountered
- Histogram property test failed on first run: `old="\n\x00"`, `new="\x00"`. Root cause was the post-sort architecture, not a simple off-by-one. Required architectural change (tagged stack) rather than a sort key tweak.

## Next Phase Readiness
- Phase 2 complete: Myers, Patience, Histogram all pass 1000-run rapid property checks
- All ALGO-01..04 requirements satisfied
- Ready for Phase 3: rendering / hunk display

---
*Phase: 02-algorithms*
*Completed: 2026-03-25*
