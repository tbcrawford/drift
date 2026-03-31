---
phase: 17-address-medium-priority-council-review-issues
plan: "04"
subsystem: internal/render
tags: [split-view, pairing, visual-quality, bottom-aligned]
dependency_graph:
  requires: [17-01]
  provides: [bottom-aligned-pair-hunk-lines]
  affects: [internal/render/split.go, internal/render/split_test.go]
tech_stack:
  added: []
  patterns: [bottom-aligned-delete-insert-pairing]
key_files:
  modified:
    - internal/render/split.go
    - internal/render/split_test.go
decisions:
  - "Bottom-aligned pairing: for D>I, top (D-I) deletes get blank right placeholder; bottom I pair with inserts. For I>D, top (I-D) inserts get blank left placeholder; bottom D pair with deletes. Matches git split-view convention."
  - "Existing TestPairHunkLines_MoreDeletesThanInserts and TestPairHunkLines_MoreInsertsThanDeletes updated to reflect new bottom-aligned semantics — the old top-aligned behavior was intentionally replaced."
metrics:
  duration: "4 minutes"
  completed: "2026-03-31"
  tasks_completed: 2
  files_modified: 2
requirements_satisfied: [REVIEW-08]
---

# Phase 17 Plan 04: Bottom-Aligned pairHunkLines Summary

**One-liner:** Rewrote `pairHunkLines` with git-style bottom-aligned insert pairing for asymmetric delete/insert blocks, matching git's split-view visual convention.

## What Was Built

Replaced the top-aligned (positional) `pairHunkLines` algorithm in `internal/render/split.go` with a bottom-aligned algorithm that matches git's split-view behavior:

- **For D > I (more deletes than inserts):** The top `(D-I)` delete rows receive blank right-side placeholders; the bottom `I` deletes pair with all inserts. The insert "anchors" to the bottom of the delete block.
- **For I > D (more inserts than deletes):** The top `(I-D)` insert rows receive blank left-side placeholders; the delete pairs with the last `D` inserts. The delete "anchors" to the bottom of the insert block.
- **Equal-size blocks and equal lines:** Unchanged — positional pairing still applies.

## Tasks Completed

| Task | Name | Commit | Files |
|------|------|--------|-------|
| 1 | Rewrite pairHunkLines with bottom-aligned pairing | 3b5bf54 | internal/render/split.go |
| 2 | Add TestPairHunkLines_BottomAligned and run full suite | 0f1dd2b | internal/render/split_test.go |

## Verification

```
go test ./internal/render/... -run TestPairHunkLines -v → 9 tests PASS
go test ./...                                           → 210 tests PASS
go vet ./internal/render/...                            → CLEAN
```

All 9 `TestPairHunkLines` tests pass:
- `TestPairHunkLines_EqualLines`
- `TestPairHunkLines_MoreDeletesThanInserts` (updated to bottom-aligned expectations)
- `TestPairHunkLines_MoreInsertsThanDeletes` (updated to bottom-aligned expectations)
- `TestPairHunkLines_OnlyDeletes`
- `TestPairHunkLines_OnlyInserts`
- `TestPairHunkLines_BottomAligned/3del_1ins`
- `TestPairHunkLines_BottomAligned/1del_3ins`
- `TestPairHunkLines_BottomAligned/2del_2ins_unchanged`

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Updated pre-existing tests that asserted old top-aligned behavior**
- **Found during:** Task 2 (pre-analysis before writing tests)
- **Issue:** `TestPairHunkLines_MoreDeletesThanInserts` and `TestPairHunkLines_MoreInsertsThanDeletes` both asserted the OLD top-aligned pairing behavior. They would have failed with the new algorithm.
- **Fix:** Updated both tests to assert the correct bottom-aligned expectations, in addition to appending the new `TestPairHunkLines_BottomAligned` test.
- **Files modified:** `internal/render/split_test.go`
- **Commit:** 0f1dd2b

## Known Stubs

None — all data flows are fully wired.

## Self-Check: PASSED

- ✅ `internal/render/split.go` — modified with bottom-aligned pairHunkLines
- ✅ `internal/render/split_test.go` — modified with updated + new tests
- ✅ Commit 3b5bf54 — exists in git log
- ✅ Commit 0f1dd2b — exists in git log
