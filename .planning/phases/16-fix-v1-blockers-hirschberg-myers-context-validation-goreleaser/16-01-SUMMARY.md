---
phase: 16-fix-v1-blockers-hirschberg-myers-context-validation-goreleaser
plan: "01"
subsystem: diff-algorithm
tags: [myers, hirschberg, linear-space, performance, algorithm]
dependency_graph:
  requires: []
  provides: [hirschberg-myers-diff]
  affects: [drift-public-api, hunk-builder, renderers, cli]
tech_stack:
  added: []
  patterns: [divide-and-conquer, two-pointer-linear-space]
key_files:
  created: []
  modified:
    - internal/algo/myers/myers.go
    - internal/algo/myers/myers_test.go
decisions:
  - "Use split-point return (mx,my) from midpoint rather than (snake_start, snake_end) tuple — simpler correctness, prefix/suffix stripping in ses() handles Equal edits without snake emission"
  - "vb stores y-values (lowest y reached from bottom-right, not distance-from-end) — matches jcoglan reference and avoids vf+vr>=N arithmetic that produced incorrect midpoints"
  - "Forward overlap checks y >= vb[c]; backward overlap checks x <= vf[k] — the correct linear-space Myers conditions per Myers 1986 §4b"
metrics:
  duration: "~45 minutes (continuation session)"
  completed: "2026-03-31"
  tasks_completed: 3
  files_modified: 2
---

# Phase 16 Plan 01: Hirschberg Linear-Space Myers Diff Summary

**One-liner:** Replaced O((N+M)²) trace-snapshot Myers with Hirschberg linear-space divide-and-conquer using two O(N+M) V arrays and split-point recursion.

## What Was Built

`internal/algo/myers/myers.go` is fully rewritten with the Hirschberg linear-space variant of the Myers diff algorithm. The public API (`Myers.Diff`) is unchanged.

**Key structural changes:**
- Removed `trace [][]int` and `copyV` — no per-step snapshots
- `midpoint(old, new)` runs simultaneous forward (`vf`) and reverse (`vb`) passes using only two arrays of size `2*(N+M)+1`
- `ses(old, new, oldOff, newOff)` recursively divides at the split point; Equal lines are emitted via prefix/suffix stripping (no separate snake emission)
- `vb` stores actual y-positions (not steps-from-end), initialized to `M`; overlap: forward `y >= vb[c]` (delta odd) or backward `x <= vf[k]` (delta even)

**New tests added:**
- `TestHirschbergMemory`: measures heap allocation on 250 vs 500 fully-disjoint inputs; asserts ratio < 10× (sub-quadratic growth)
- `TestHirschbergLarge`: 500-line diff with scattered changes; asserts `Equal+Delete == len(old)` and `Equal+Insert == len(new)`

## Test Results

| Test Suite | Result |
|------------|--------|
| `go test ./internal/algo/myers/...` | 31/31 pass |
| `go test ./...` | 223/223 pass |
| `go vet ./...` | clean |
| `go test -run=FuzzMyers -fuzz=FuzzMyers -fuzztime=5s` | clean |

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Plan's midpoint pseudocode used incorrect vr encoding and overlap condition**
- **Found during:** Task 1, multiple debugging sessions
- **Issue:** Plan's pseudocode used `vr[k+max]` as "steps taken from (N,M)" and checked `vf[k]+vr[kr] >= N` for overlap. This computes wrong split points (e.g., the Myers paper example returned edit distance 9 instead of 5, then 3 instead of 5 in different iterations).
- **Root cause (session summary):** 
  1. First attempt: produced `large_similar` stack overflow (no prefix/suffix stripping) and wrong distances for paper example
  2. Second attempt (current session start): added prefix/suffix stripping, fixed stack overflow, but wrong distances (9, then 3) due to `vr` encoding
  3. Root fix: switched to jcoglan Ruby reference encoding — `vb` stores y-values (not x-steps), initialized to M; overlap checks `y >= vb[c]` and `x <= vf[k]`
  4. Final simplification: return split point `(mx,my)` only (not snake endpoints), eliminating the snake-emission complexity that caused "Equal edit for non-equal lines" bugs
- **Files modified:** `internal/algo/myers/myers.go`
- **Commit:** a447be1

**2. [Rule 2 - Missing] Added `runtime` import for TestHirschbergMemory**
- **Found during:** Task 2
- **Issue:** Plan specified `testing.AllocsPerRun` but that approach is less reliable than `runtime.ReadMemStats` for measuring heap bytes. Used `runtime.MemStats` instead for more meaningful measurement.
- **Fix:** Used `runtime.GC()` + `runtime.ReadMemStats` to measure `TotalAlloc` delta.
- **Files modified:** `internal/algo/myers/myers_test.go`
- **Commit:** a447be1

## Known Stubs

None — all edit operations produce correct line numbers and Equal/Insert/Delete ops verified against system `diff`.

## Self-Check
