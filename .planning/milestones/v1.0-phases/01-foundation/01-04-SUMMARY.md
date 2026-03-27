---
phase: 01-foundation
plan: 04
subsystem: diff-engine
tags: [go, diff, hunk, myers, crlf, import-cycle, type-alias]

requires:
  - phase: 01-foundation/01-03
    provides: Myers diff algorithm returning []Edit from two string slices

provides:
  - internal/hunk package: Build(edits, oldLines, newLines, contextLines) → []Hunk
  - drift.Diff(old, new string, ...Option) (DiffResult, error) public API
  - internal/edittype package: canonical type definitions breaking import cycle
  - Windows CRLF normalization before line splitting
  - Configurable context window (default 3, WithContext(0) supported)

affects:
  - 01-05 (property-based testing uses drift.Diff)
  - Phase 2 (renderers consume DiffResult.Hunks)
  - Phase 3 (CLI uses drift.Diff as entry point)

tech-stack:
  added: []
  patterns:
    - "internal/edittype: canonical shared types to break root↔internal import cycles"
    - "type aliases in root (type Edit = edittype.Edit) re-export internal types publicly"
    - "TDD: hunk builder written RED→GREEN; drift.Diff written directly with tests"
    - "algoInterface local definition in drift.go avoids importing internal/algo"

key-files:
  created:
    - internal/edittype/edittype.go
    - internal/hunk/hunk.go
    - internal/hunk/hunk_test.go
    - drift_test.go
  modified:
    - drift.go
    - types.go
    - internal/algo/algo.go
    - internal/algo/myers/myers.go

key-decisions:
  - "internal/edittype package introduced to break import cycle (drift root → internal/algo/myers → drift root)"
  - "All public types (Op, Edit, Span, Line, Hunk, DiffResult) are now type aliases pointing to edittype; public API is unchanged"
  - "hunk.Build() operates on edit-sequence indices (not line numbers) for context expansion and merge — natural unit of iteration"
  - "drift.Diff fast path: identical inputs (post-normalization) short-circuit before splitting lines"
  - "Patience/Histogram stubs fall through to Myers — Phase 2 will replace"

patterns-established:
  - "Import cycle pattern: when root imports internal and internal imports root, extract shared types to internal/edittype and re-export via type aliases"
  - "Hunk context expansion in edit-space: expand changed edit indices by contextLines, merge overlapping ranges, then build Line slices per merged range"

requirements-completed: [CORE-02, CORE-04, CORE-06, CORE-07]

duration: 6min
completed: 2026-03-25
---

# Phase 01 Plan 04: Hunk Builder and Public Diff API Summary

**Hunk builder (edit-space context expansion + merge) and drift.Diff() public API wired with CRLF normalization; import cycle resolved via internal/edittype type-alias pattern**

## Performance

- **Duration:** ~6 min
- **Started:** 2026-03-25T19:00:46Z
- **Completed:** 2026-03-25T19:06:39Z
- **Tasks:** 2 (Task 1: hunk builder TDD; Task 2: drift.Diff() API)
- **Files modified:** 8

## Accomplishments

- `internal/hunk.Build()` converts `[]Edit` → `[]Hunk` with configurable context window, merging overlapping/adjacent ranges
- `drift.Diff(old, new string, ...Option) (DiffResult, error)` wires Myers → hunk builder → DiffResult in one call
- `internal/edittype` package resolves the root↔internal import cycle; all public types remain stable via type aliases
- Windows CRLF normalization (`\r\n → \n`) applied before diffing; trailing empty-line stripping from `strings.Split`

## Task Commits

1. **Task 1 RED: Add failing hunk builder tests** - `f0b682e` (test)
2. **Task 1 GREEN: Implement hunk builder** - `80f168e` (feat)
3. **Task 2: Wire drift.Diff() public API** - `32aff6c` (feat)

## Files Created/Modified

- `internal/edittype/edittype.go` - Canonical Op, Edit, Span, Line, Hunk, DiffResult types; imported by internal packages to break cycle
- `internal/hunk/hunk.go` - Build() algorithm: expand changed indices, merge ranges, walk edits to build Lines, compute hunk header
- `internal/hunk/hunk_test.go` - 6 behavior tests: no changes, single change middle/start, two distant hunks, merged hunk, zero context
- `drift.go` - Public Diff() function with CRLF normalization, splitLines(), algoInterface, Myers dispatch
- `drift_test.go` - 5 integration tests: identical inputs, CRLF normalization, single added line, context=0, trailing newline
- `types.go` - Re-exports all types as aliases from internal/edittype (unchanged public API)
- `internal/algo/algo.go` - Updated Differ interface to use edittype.Edit (no drift root import)
- `internal/algo/myers/myers.go` - Updated to use edittype types (removed drift root import)

## Decisions Made

- **import cycle resolution**: The plan's note about avoiding cycles was correct in spirit but the implementation required extracting ALL shared types (`Op`, `Edit`, `Span`, `Line`, `Hunk`, `DiffResult`) to `internal/edittype`. Root re-exports via type aliases — public API is identical. This pattern will be needed any time a new internal package needs both root's exported types AND is imported by root.
- **edit-space context expansion**: `hunk.Build()` tracks *edit-sequence indices* (not line numbers) for context. This handles inserts (which have no OldLine) and deletes (no NewLine) uniformly without special-casing.
- **algoInterface in drift.go**: A local `algoInterface` in `drift.go` satisfies the `*myers.Myers` return value from `myers.New()` without importing `internal/algo`, keeping the import graph acyclic.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Resolved import cycle between drift root and internal packages**
- **Found during:** Task 2 (implementing drift.Diff())
- **Issue:** Plan's suggested code imported `internal/algo/myers` from root `drift.go`, but `internal/algo/myers` imports root `drift` for `drift.Edit`. Go rejects this cycle.
- **Fix:** Created `internal/edittype` package containing all shared types. Root `types.go` re-exports via type aliases (`type Edit = edittype.Edit`). Both `myers.go` and `hunk.go` import `edittype` instead of `drift`.
- **Files modified:** `internal/edittype/edittype.go` (new), `types.go`, `internal/algo/algo.go`, `internal/algo/myers/myers.go`, `internal/hunk/hunk.go`
- **Verification:** `go build ./...` passes; all 23 tests pass; `go vet ./...` clean
- **Committed in:** `32aff6c` (Task 2 commit)

---

**Total deviations:** 1 auto-fixed (Rule 1 - Bug: import cycle)
**Impact on plan:** Necessary for compilation. Public API is unchanged — `drift.Edit`, `drift.Op`, etc. work identically to plan. The `internal/edittype` pattern is now established for future internal packages.

## Issues Encountered

- Import cycle was the only issue. Resolved cleanly via type-alias re-export pattern. No scope creep.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Core diff engine is complete: `drift.Diff(a, b)` → `DiffResult{Hunks, IsEqual}` ✓
- Ready for Phase 01-05: property-based testing and cross-validation of the full pipeline
- The `internal/edittype` import pattern is documented for Phase 2 renderer packages

---
*Phase: 01-foundation*
*Completed: 2026-03-25*
