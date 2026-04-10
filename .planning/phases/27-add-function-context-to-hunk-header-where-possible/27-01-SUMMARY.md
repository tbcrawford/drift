---
phase: 27-add-function-context-to-hunk-header-where-possible
plan: "01"
subsystem: cmd/drift/unifieddiff
tags: [parsing, git-pager, code-fragment, tdd]
dependency_graph:
  requires: []
  provides: [hunkFragment, parsedFileDiff.Hunks, parseHunkHeader]
  affects: [cmd/drift/main.go, internal/render]
tech_stack:
  added: []
  patterns: [strings parsing, TDD red-green]
key_files:
  created: []
  modified:
    - cmd/drift/unifieddiff.go
    - cmd/drift/unifieddiff_test.go
decisions:
  - "Used strings.Index to find ' @@' closing marker; no regex per CONTEXT.md D-03"
  - "fmt.Sscanf for integer parsing from range strings — simple and idiomatic"
  - "strings.TrimSpace on code_fragment so whitespace-only yields empty string"
  - "parseHunkHeader returns ok=false for malformed lines; backfill is optional (no panic path)"
metrics:
  duration: "~4 minutes"
  completed: "2026-04-10T19:19:00Z"
  tasks_completed: 1
  files_modified: 2
---

# Phase 27 Plan 01: Extract code_fragment from git @@ hunk header lines — Summary

**One-liner:** TDD extraction of git hunk header code_fragment into `hunkFragment` struct with `parseHunkHeader` helper and 5-subtest coverage.

## What Was Built

`parseUnifiedDiff` in `cmd/drift/unifieddiff.go` now captures the `code_fragment` that git appends after the closing `@@` on hunk header lines (e.g. `@@ -12,7 +12,9 @@ func ParseOptions(args []string)`).

### New Types

```go
// hunkFragment holds parsed @@ header metadata per hunk.
type hunkFragment struct {
    OldStart     int
    NewStart     int
    CodeFragment string // "" when git omitted it
}
```

### Extended parsedFileDiff

```go
type parsedFileDiff struct {
    // ... existing fields ...
    Hunks []hunkFragment // populated by parseUnifiedDiff per @@ line
}
```

### New Helper: parseHunkHeader

`parseHunkHeader(line string) (oldStart, newStart int, codeFragment string, ok bool)` — parses the range fields and extracts code_fragment using `strings.Index` on the closing ` @@` marker. Returns `ok=false` for malformed lines (safe fail-open).

### Updated @@ Case in parseUnifiedDiff

The `case strings.HasPrefix(line, "@@ "):` switch branch now calls `parseHunkHeader` and appends the result to `current.Hunks`. Content reconstruction is unchanged.

## Test Coverage

`TestParseUnifiedDiff_CodeFragment` (5 subtests):
1. **single hunk with code_fragment** — `func ParseOptions(args []string)` extracted, NewStart=12 verified
2. **hunk without code_fragment** — `""` returned cleanly (no regression)
3. **multi-hunk with different code_fragments** — `func Alpha()` and `func Beta()` each in correct Hunks slot
4. **ANSI-colored @@ line** — ANSI already stripped by `ansi.Strip` before parsing; code_fragment extracted correctly
5. **whitespace-only code_fragment** — `strings.TrimSpace` normalizes to `""`

All 12 original `TestParseUnifiedDiff` subtests pass unchanged.

## Commit

- `53ed9ee` — `feat(27-01): extract code_fragment from git @@ hunk header lines`

## Deviations from Plan

None — plan executed exactly as written.

## Self-Check: PASSED

- [x] `cmd/drift/unifieddiff.go` defines `hunkFragment` struct
- [x] `parsedFileDiff` has `Hunks []hunkFragment` field  
- [x] `parseUnifiedDiff` populates `Hunks` per `@@ ` line
- [x] `parseHunkHeader` correctly extracts code_fragment or ""
- [x] `go test ./cmd/drift/... -run TestParseUnifiedDiff -count=1` passes (18 tests)
- [x] All pre-existing TestParseUnifiedDiff subtests still pass
- [x] `go build ./cmd/drift/...` compiles cleanly
