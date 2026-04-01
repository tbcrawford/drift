---
phase: 18-auto-algorithm-mode
plan: "01"
subsystem: diff-algorithm
tags: [algorithm, auto, heuristic, default, go]
dependency_graph:
  requires: []
  provides: [drift.Auto, selectAuto, auto-default]
  affects: [options.go, drift.go, cmd/drift/main.go, doc.go, tests]
tech_stack:
  added: []
  patterns: [O(N)-heuristic, dispatch-switch, property-testing]
key_files:
  created: []
  modified:
    - options.go
    - drift.go
    - cmd/drift/main.go
    - drift_algorithm_integration_test.go
    - drift_property_test.go
    - doc.go
decisions:
  - "Auto iota=3 appended after Myers=0, Patience=1, Histogram=2 — preserves backward-compat for existing callers using integer algorithm values"
  - "selectAuto thresholds: smallFileThreshold=2000 total lines, maxFreqThreshold=32 old-side line occurrences — from AUTO-ALGORITHM.md research"
  - "default: case in dispatch switch remains Myers for invalid/unknown values — safe fallback"
  - "Go 1.25 minimum allows integer range (range 1001) in TestAuto_SelectsMyers_LargeFile"
metrics:
  duration: "253 seconds (~4 min)"
  completed: "2026-04-01"
  tasks_completed: 7
  files_modified: 6
---

# Phase 18 Plan 01: Auto Algorithm Mode Summary

**One-liner:** O(N) heuristic Auto algorithm selects Histogram for small/clean files and Myers otherwise; Auto is now the default replacing Myers.

## What Was Built

Added `drift.Auto` as the fourth `Algorithm` constant (iota=3) and made it the new default for `drift.Diff()`. The `selectAuto()` heuristic chooses between Myers and Histogram at diff-time using an O(N) frequency scan: use Histogram for files ≤ 2000 total lines where no old-side line appears more than 32 times; use Myers otherwise. Updated the CLI `--algorithm` flag to default to `"auto"` and extended all test layers (integration, property) to cover the new algorithm.

## Commits

| Task | Commit | Description |
|------|--------|-------------|
| Task 1 | f79c502 | feat(18-01): add Auto constant and update defaultConfig() to use Auto |
| Task 2 | a893208 | feat(18-01): add Auto dispatch case and selectAuto() heuristic in drift.go |
| Task 3 | e8f523a | feat(18-01): update CLI parseAlgorithm() and --algorithm flag default to auto |
| Task 4 | 48e33ca | test(18-01): add Auto integration tests; rename default test |
| Task 5 | 0d4f7b0 | test(18-01): add TestProperty_RoundTrip_Auto to property tests |
| Task 6 | ba89ea5 | docs(18-01): update doc.go to document Auto algorithm as the new default |
| Task 7 | ce39a30 | chore(18-01): verify full test suite passes with Auto algorithm |

## Changes by File

### options.go
- Added `Auto Algorithm = iota` constant (value 3) after Histogram
- Changed `defaultConfig()` from `algorithm: Myers` to `algorithm: Auto`
- Updated `WithAlgorithm` godoc to mention Auto

### drift.go
- Added `case Auto: differ = selectAuto(oldLines, newLines)` to dispatch switch
- Updated `default:` comment to clarify it covers Myers and invalid values
- Added `selectAuto(old, new []string) algoInterface` function with O(N) heuristic

### cmd/drift/main.go
- Added `case "auto": return drift.Auto, nil` to `parseAlgorithm()` (listed first)
- Updated error message to list `auto, myers, patience, histogram`
- Changed `--algorithm` flag default from `"myers"` to `"auto"`
- Updated flag help text to list `auto` first

### drift_algorithm_integration_test.go
- Renamed `TestWithAlgorithm_Myers_StillDefault` → `TestDefault_Algorithm_RoundTrip`
- Updated error message in renamed test to reflect Auto default
- Added `drift.Auto` to `TestAllAlgorithmsCorrect` algos/algoNames slices
- Added `TestWithAlgorithm_Auto_RoundTrip`
- Added `TestAuto_SelectsHistogram_SmallCleanFile`
- Added `TestAuto_SelectsMyers_HighFrequency`
- Added `TestAuto_SelectsMyers_LargeFile`

### drift_property_test.go
- Added `TestProperty_RoundTrip_Auto` using rapid property-based testing

### doc.go
- Updated package description to list Auto alongside other algorithms
- Rewrote `WithAlgorithm` doc section to document all 4 algorithms
- Described Auto as the default with O(N) heuristic details

## Verification Results

```
go test ./...: 240 tests pass (16 packages)
go vet ./...:  clean
go build ./...: clean
--algorithm flag: (default "auto") in --help output
```

## Deviations from Plan

None — plan executed exactly as written.

## Known Stubs

None — all functionality is fully wired.

## Self-Check

- [x] `drift.Auto` defined as iota=3 in options.go
- [x] `defaultConfig()` returns `algorithm: Auto`
- [x] `selectAuto()` present in drift.go with correct thresholds
- [x] Auto case wired in Diff() switch
- [x] `parseAlgorithm("auto")` returns `drift.Auto`
- [x] `--algorithm` defaults to `"auto"` in CLI
- [x] `TestDefault_Algorithm_RoundTrip` passes (renamed)
- [x] All 4 new Auto tests pass
- [x] Property test covers Auto
- [x] 240 total tests pass
- [x] doc.go documents Auto as default

## Self-Check: PASSED
