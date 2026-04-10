---
phase: 27-add-function-context-to-hunk-header-where-possible
plan: "02"
subsystem: [internal/edittype, internal/render, cmd/drift]
tags: [rendering, git-pager, code-fragment, tdd, integration-test]
dependency_graph:
  requires: [27-01-SUMMARY.md]
  provides: [edittype.Hunk.CodeFragment, unified-renderer-codefragment, split-renderer-codefragment, runPagerMode-backfill]
  affects: [internal/edittype/edittype.go, internal/render/unified.go, internal/render/split.go, cmd/drift/main.go]
tech_stack:
  added: []
  patterns: [positional hunk backfill, conditional format string, TDD integration test]
key_files:
  created: []
  modified:
    - internal/edittype/edittype.go
    - internal/render/unified.go
    - internal/render/split.go
    - cmd/drift/main.go
    - cmd/drift/main_test.go
decisions:
  - "Match parsed hunks to rendered hunks by position (not start line) — simpler, safe for contiguous files"
  - "Guard with i < len(f.Hunks) per T-27-05 mitigation: no panic when hunk counts diverge"
  - "Integration test uses sentinel 'parseXYZ' absent from diff content — strict proof that code_fragment comes from @@ header backfill"
  - "Standalone drift.Diff() path unchanged: CodeFragment is zero-value ''; no options, no standalone regex"
metrics:
  duration: "~4 minutes"
  completed: "2026-04-10T19:23:09Z"
  tasks_completed: 2
  files_modified: 5
---

# Phase 27 Plan 02: Thread code_fragment Through Rendering Pipeline — Summary

**One-liner:** `edittype.Hunk.CodeFragment` field backfilled in git pager mode by positional matching from `parsedFileDiff.Hunks`, rendering `@@ ... @@ func_name` headers in both unified and split views.

## What Was Built

### Task 1: CodeFragment field + renderer updates

**`internal/edittype/edittype.go`** — `Hunk` struct gains `CodeFragment string`:
```go
type Hunk struct {
    OldStart     int
    OldLines     int
    NewStart     int
    NewLines     int
    Lines        []Line
    CodeFragment string // function/class context from git @@ header; "" in standalone mode
}
```

**`internal/render/unified.go`** — hunk header now conditionally includes code_fragment:
```go
var header string
if h.CodeFragment != "" {
    header = fmt.Sprintf("@@ -%d,%d +%d,%d @@ %s\n", ...)
} else {
    header = fmt.Sprintf("@@ -%d,%d +%d,%d @@\n", ...)
}
```

**`internal/render/split.go`** — same conditional pattern for side-by-side headers.

### Task 2: Backfill in runPagerMode + integration test

**`cmd/drift/main.go:runPagerMode`** — after `drift.Diff()`, backfills CodeFragment by position:
```go
for i := range result.Hunks {
    if i < len(f.Hunks) {
        result.Hunks[i].CodeFragment = f.Hunks[i].CodeFragment
    }
}
```

**`cmd/drift/main_test.go`** — two new integration tests:
- `TestRunCLI_pagerMode_codeFragment`: sentinel `parseXYZ` absent from diff content; presence in rendered output proves @@ header backfill works
- `TestRunCLI_pagerMode_noCodeFragment`: regression check — no trailing space when CodeFragment is empty

## TDD Flow

**RED**: `TestRunCLI_pagerMode_codeFragment` failed because `runPagerMode` returned output with correct `@@` lines but no code_fragment (backfill loop not yet added).

**GREEN**: Added backfill loop in `runPagerMode`. Test passed immediately.

## Test Results

- 383 total tests across 16 packages — all pass
- `go vet ./...` — no issues
- `go build ./...` — clean

## Commits

- `d2c51ed` — `feat(27-02): thread code_fragment through rendering pipeline for git pager mode`

## Deviations from Plan

**1. [Rule 1 - Bug] Strict integration test required sentinel-based approach**
- **Found during:** Task 2 (TDD RED phase)
- **Issue:** The first draft of `TestRunCLI_pagerMode_codeFragment` searched for "func ParseOptions" which appeared in the diff *content* lines, causing the test to pass without the backfill implemented
- **Fix:** Redesigned test to use sentinel `parseXYZ` that appears ONLY in the `@@` header, not in any content line. Now the test is a strict proof of the backfill path
- **Files modified:** `cmd/drift/main_test.go`
- **Commit:** d2c51ed (included in same commit)

## Threat Model Verification

| Threat ID | Status |
|-----------|--------|
| T-27-03 (code_fragment in terminal output) | Accepted — user's own source content |
| T-27-04 (function name disclosure) | Accepted — user already sees full diff |
| T-27-05 (hunk index mismatch) | Mitigated — `if i < len(f.Hunks)` guard present ✓ |

## Known Stubs

None — code_fragment is fully wired from git `@@` header through to rendered output.

## Self-Check: PASSED

- [x] `internal/edittype/edittype.go` Hunk struct has `CodeFragment string` field
- [x] `internal/render/unified.go` conditionally emits `@@ ... @@ CodeFragment`
- [x] `internal/render/split.go` conditionally emits `@@ ... @@ CodeFragment`
- [x] `cmd/drift/main.go:runPagerMode` contains CodeFragment backfill loop with bounds guard
- [x] `go test ./... -count=1` passes (383 tests)
- [x] `go vet ./...` passes
- [x] Standalone `drift.Diff()` hunks always have `CodeFragment=""` (no standalone-path changes)
- [x] `d2c51ed` commit exists in git log
