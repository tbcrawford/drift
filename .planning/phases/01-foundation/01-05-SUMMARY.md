---
phase: 01-foundation
plan: 05
subsystem: testing
tags: [property-testing, fuzz-testing, rapid, myers, correctness]
dependency_graph:
  requires: [01-04]
  provides: [round-trip-property-tests, myers-fuzz-test]
  affects: [drift_property_test.go, testdata/apply.go, internal/algo/myers/myers_fuzz_test.go]
tech_stack:
  added: [pgregory.net/rapid v1.2.0]
  patterns: [property-based-testing, fuzz-testing, canonical-normalization]
key_files:
  created:
    - drift_property_test.go
    - testdata/apply.go
    - internal/algo/myers/myers_fuzz_test.go
  modified:
    - go.mod
    - go.sum
decisions:
  - "Property tests compare canonical (normalized) text, not raw line slices, to handle drift's trailing-newline stripping"
  - "Apply() uses 0-indexed cursor through oldLines tracking hunk coverage, not line-number-set lookup"
  - "Fuzz test verifies edit-level round-trip (applyEdits) rather than hunk-level (Apply) for maximum coverage"
  - "SliceOfN(elem, -1, 50) used instead of SliceOf + MaxLen (rapid v1.2.0 API is SliceOfN not options pattern)"
metrics:
  duration: 432
  completed_date: "2026-03-25"
  tasks_completed: 3
  files_created: 5
---

# Phase 01 Plan 05: Property-Based and Fuzz Testing Summary

**One-liner:** Property-based tests with pgregory.net/rapid verifying `Apply(Diff(old,new), oldLines) == newLines` for 100 random input pairs, plus Myers fuzz test running 30s without panics.

## What Was Built

Three test files that complete Phase 1's OSS-08 correctness requirement:

1. **`testdata/apply.go`** — `Apply(DiffResult, []string) []string` helper that reconstructs the new file from a diff result and the original old lines. Uses a sequential cursor through `result.Hunks`, emitting:
   - Old lines before each hunk (context not covered by any hunk)
   - Equal and Insert lines within hunks
   - Skipping Delete lines

2. **`drift_property_test.go`** — Three `rapid.Check` property tests:
   - `TestProperty_RoundTrip`: `Apply(Diff(old,new), canonicalLines(old))` joined == `canonicalLines(new)` joined (100 runs)
   - `TestProperty_IdenticalInputs`: `Diff(text, text).IsEqual == true` and `len(Hunks) == 0` (100 runs)
   - `TestProperty_HunkAccounting`: `sum(h.OldLines) <= len(oldLines)` and `sum(h.NewLines) <= len(newLines)` (100 runs)

3. **`internal/algo/myers/myers_fuzz_test.go`** — `FuzzMyers` with 10 seed corpus entries. The fuzz body:
   - Calls `myers.New().Diff(...)` — must not panic
   - Validates edit sequence structure (line number bounds, Equal content match)
   - Verifies edit-level round-trip invariant

## Verification Results

- `go test -run TestProperty ./...` — 3 properties × 100 runs = PASS
- `go test -fuzz=FuzzMyers -fuzztime=30s ./internal/algo/myers/...` — 0 panics, 0 failures
- `go mod tidy` — clean, rapid promoted from indirect to direct dependency
- `go test ./...` — all 43 tests pass

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] rapid.MaxLen does not exist in v1.2.0**
- **Found during:** Initial test writing
- **Issue:** Plan's `<implementation>` described `rapid.SliceOf(elem, rapid.MaxLen(50))` but rapid v1.2.0's API is `rapid.SliceOfN(elem, minLen, maxLen int)` — no options-style modifiers exist
- **Fix:** Used `rapid.SliceOfN(rapid.StringMatching(...), -1, 50)` throughout
- **Files modified:** `drift_property_test.go`
- **Commit:** 6947074

**2. [Rule 1 - Bug] Round-trip invariant fails at trailing-newline boundary**
- **Found during:** First test run (rapid shrunk to `oldLines=[], newLines=["",""]`)
- **Issue:** `strings.Join(["",""], "\n") = "\n"` but `drift.splitLines("\n") = [""]`. The plan's stated invariant `Apply(Diff(join(old), join(new)), oldLines) == newLines` breaks when `newLines` has a trailing empty element that drift normalizes away. E.g., `newLines=[""]` and `drift` reconstructs `[""]`, but `strings.Join([""],"") = ""` ≠ `"\n"` = `strings.Join(["",""],"\n")`.
- **Fix:** Test compares `strings.Join(Apply(...), "\n")` against `strings.Join(canonicalLines(newText), "\n")` where `canonicalLines` mirrors `drift.splitLines`. Added `canonicalLines()` helper. Used canonical `oldLines` (from `canonicalLines(oldText)`) as `Apply()` input.
- **Files modified:** `drift_property_test.go`
- **Commit:** 6947074

## Known Stubs

None — all files produce real verified output with no placeholder data.

## Self-Check: PASSED

- [x] `drift_property_test.go` exists and compiles
- [x] `testdata/apply.go` exists and compiles  
- [x] `internal/algo/myers/myers_fuzz_test.go` exists and compiles
- [x] Commit 6947074 exists (property tests + apply helper)
- [x] Commit a8c5a0a exists (fuzz test)
- [x] Commit 04868f3 exists (go mod tidy)
- [x] All 43 tests pass: `go test ./...`
