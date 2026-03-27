---
phase: 01-foundation
plan: "03"
subsystem: diff-algorithm
tags: [tdd, myers, algorithm, go]
dependency_graph:
  requires: [01-02]
  provides: [myers-differ]
  affects: [internal/algo/myers]
tech_stack:
  added: []
  patterns: [tdd-red-green-refactor, myers-ses, compile-time-interface-check]
key_files:
  created:
    - internal/algo/myers/myers.go
    - internal/algo/myers/myers_test.go
  modified: []
decisions:
  - "Myers trace saved at END of d-loop after k-loop — off-by-one invisible on short inputs surfaces at 100+ line files"
  - "V array pre-allocated to 2*(N+M)+1 to avoid O(N^2) allocations during forward pass"
  - "Edge cases handled before main algorithm: both-empty, old-empty (all inserts), new-empty (all deletes)"
  - "Backtracking emits edits in reverse then reverses slice — simpler than prepending"
  - "1-indexed OldLine/NewLine applied during backtrack (not post-processing)"
metrics:
  duration: "3m 51s"
  completed: "2026-03-25"
  tasks_completed: 3
  files_created: 2
  files_modified: 0
---

# Phase 01 Plan 03: Myers Diff Algorithm Summary

**One-liner:** Myers SES algorithm implementing `algo.Differ` with TDD-verified correctness on toy inputs, paper examples, and real-world 50-line cross-validation against `diff -u`.

## What Was Built

`internal/algo/myers/` — a complete Myers diff algorithm implementation:

- **`myers.go`** (182 lines): Myers struct implementing `algo.Differ`, forward pass with V array pre-allocated to `2*(N+M)+1`, trace saved at END of d-loop, backtracking to produce ordered `[]drift.Edit`
- **`myers_test.go`** (410 lines): 9 top-level tests (18 sub-tests) including identical/empty edge cases, Myers 1986 paper example, precise edit sequence tests, cross-validation against system `diff -u`, and line invariant property tests on 9 cases including a 100-line large input

## TDD Phases

| Phase | Commit | Status |
|-------|--------|--------|
| RED — failing tests | `21c100c` | 18 tests written, 0 passing (no impl) |
| GREEN — implementation | `021d853` | All 18 tests pass |
| REFACTOR — comment clarity | `b274825` | All 18 tests pass, forward-pass trace comment expanded |

## Commits

| Commit | Type | Description |
|--------|------|-------------|
| `21c100c` | `test` | RED phase: 9 failing tests covering all cases |
| `021d853` | `feat` | GREEN phase: Myers algorithm, all 18 tests pass |
| `b274825` | `refactor` | REFACTOR: improve trace save comment clarity |

## Verification

```
go test ./internal/algo/myers/... -v     → 18 passed
go test ./internal/algo/myers/... -race  → no data races
go vet ./internal/algo/myers/...         → no issues
```

**Line invariant confirmed:** `Equal + Delete == len(old)` and `Equal + Insert == len(new)` for all 9 test cases including 100-line generated inputs.

## Key Technical Details

### The Critical Pitfall (Avoided)

The Myers algorithm requires saving a snapshot of the V array at the **END** of each d-loop iteration (after all k-diagonals are processed). Saving at the top gives the state *before* the current d's extensions, causing backtracking to reconstruct the wrong path. This bug is invisible on ≤10 line inputs and only surfaces at 100+ lines.

```go
// CORRECT: saved at END (after all k-diagonals update V)
for d := 0; d <= maxD; d++ {
    for k := -d; k <= d; k += 2 {
        // ... extend diagonal k ...
    }
    trace = append(trace, copyV(v))  // ← END of d-loop
}
```

### Algorithm Structure

1. **Edge cases first**: both-empty → `[]Edit{}`, old-empty → all inserts, new-empty → all deletes
2. **V array**: `make([]int, 2*(N+M)+1)` — pre-allocated, single allocation per call
3. **Forward pass**: d from 0 to N+M, k from -d to d step 2; snake extends on matches
4. **Trace**: `[][]int` snapshots of V at end of each d-iteration
5. **Backtrack**: walk d from len(trace)-1 down to 0; for each step emit snake (Equal) then edit (Insert/Delete)
6. **Reverse**: output is built in reverse order during backtracking, flipped at end

### Compile-Time Interface Check

```go
var _ algo.Differ = (*Myers)(nil)
```

Build fails immediately if the `Diff` signature diverges from the interface.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Backtracking logic produced wrong results on insert-middle case**
- **Found during:** GREEN phase first run (13/18 tests pass)
- **Issue:** Initial backtrack implementation used `x > prevX+1 && y > prevY+1` for snake, missing the last snake step; also had incorrect insertion index logic
- **Fix:** Rewrote backtracking to use `x > prevX && y > prevY` for snake, and correctly decrement y before emitting Insert, x before emitting Delete
- **Files modified:** `internal/algo/myers/myers.go`
- **Commit:** `021d853` (included in GREEN phase)

**2. [Rule 3 - Blocking] `range int` syntax (Go 1.22+) used for Go 1.21 target**
- **Found during:** GREEN phase implementation
- **Issue:** `for i := range M` requires Go 1.22+; project targets Go 1.21
- **Fix:** Changed to `for i := 0; i < M; i++` form
- **Files modified:** `internal/algo/myers/myers.go`
- **Commit:** `021d853`

## Known Stubs

None — implementation is complete and fully functional.

## Self-Check: PASSED

| Item | Status |
|------|--------|
| `internal/algo/myers/myers.go` | ✅ Found |
| `internal/algo/myers/myers_test.go` | ✅ Found |
| `.planning/phases/01-foundation/01-03-SUMMARY.md` | ✅ Found |
| commit `21c100c` (RED: failing tests) | ✅ Found |
| commit `021d853` (GREEN: implementation) | ✅ Found |
| commit `b274825` (REFACTOR: comment clarity) | ✅ Found |
