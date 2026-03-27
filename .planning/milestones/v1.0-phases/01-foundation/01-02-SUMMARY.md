---
phase: 01-foundation
plan: 02
subsystem: core-types
tags: [types, options, algo-interface, public-api, godoc]
dependency_graph:
  requires: [01-01]
  provides: [types.go, options.go, internal/algo/algo.go, doc.go]
  affects: [hunk-builder, renderers, cli, all-consumers]
tech_stack:
  added: []
  patterns: [functional-options, iota-enum, interface-contract, internal-package]
key_files:
  created:
    - types.go
    - options.go
    - doc.go
    - internal/algo/algo.go
  modified:
    - drift.go
decisions:
  - "Op enum uses iota starting at Equal=0 — consistent with zero-value meaning 'no change'"
  - "Line.Spans []Span kept nil in v1.0; reserved for v1.x intra-line word diff"
  - "internal/algo imports root package drift for Edit type — canonical Go internal layout pattern"
  - "defaultConfig() sets contextLines=3 matching git diff -U3 default"
  - "Algorithm enum defined in options.go (not internal/algo) — it's part of the public API"
metrics:
  duration: "83 seconds"
  completed: "2026-03-25"
  tasks_completed: 2
  files_created: 4
  files_modified: 1
---

# Phase 01 Plan 02: Public Type Contract & Algo Interface Summary

**One-liner:** Stable public API contract — `Op`/`Edit`/`Hunk`/`Line`/`DiffResult` types + functional options + `internal/algo.Differ` interface — zero implementation, zero breaking changes needed later.

## What Was Built

The complete exported data model and functional options pattern for `package drift`, establishing the stable public contract that every downstream layer (hunk builder, renderers, CLI) will consume.

### Files Created

| File | Purpose | Key Exports |
|------|---------|-------------|
| `types.go` | Core data model | `Op`, `Edit`, `Hunk`, `Line`, `Span`, `DiffResult` |
| `options.go` | Functional options + Algorithm enum | `Algorithm`, `Option`, `config`, `defaultConfig`, `WithAlgorithm`, `WithContext`, `WithNoColor`, `WithLang`, `WithTheme` |
| `doc.go` | Canonical package-level godoc | Package documentation with quick-start example |
| `internal/algo/algo.go` | Algorithm interface contract | `Differ` interface |

### Files Modified

| File | Change |
|------|--------|
| `drift.go` | Stripped duplicate package comment (godoc now lives in `doc.go`) |

## Architecture Pipeline

```
Algorithm layer → []drift.Edit → Hunk Builder → DiffResult → Renderer → io.Writer
     ↑
internal/algo.Differ interface (defined here)
```

All three algorithm implementations (Myers, Patience, Histogram) will satisfy `Differ.Diff(oldLines, newLines []string) []drift.Edit`.

## Decisions Made

| Decision | Rationale |
|----------|-----------|
| `Op` starts at `Equal=0` (iota) | Zero value means "unchanged" — safe default |
| `Line.Spans []Span` is nil in v1.0 | Reserved field for v1.x intra-line word diff; no cost now, no breaking change later |
| `Algorithm` enum in `options.go` (public) | Algorithm selection is a public API concern, not an internal one |
| `internal/algo` imports root `drift` | Canonical Go pattern: internal packages use root types; no circular dependency |
| `defaultConfig()` returns `contextLines=3` | Matches `git diff -U3` — the de facto standard |
| `doc.go` holds canonical package comment | Separates godoc from stub code; `drift.go` kept clean for future top-level functions |

## Verification

- [x] `go build ./...` — PASSED (zero errors)
- [x] `go vet ./...` — PASSED (no issues)
- [x] `go test ./...` — PASSED (no tests yet — declaration-only plan)
- [x] All exported types have godoc comments
- [x] `Op` enum has `Equal`/`Insert`/`Delete` constants
- [x] `Line.Spans` field exists as `[]Span` (nil by default)
- [x] `internal/algo.Differ` interface accepts `[]string`, returns `[]drift.Edit`
- [x] `defaultConfig()` returns `contextLines=3`, `algorithm=Myers`

## Deviations from Plan

None — plan executed exactly as written.

## Self-Check: PASSED

**Files exist:**
- `types.go` ✓
- `options.go` ✓
- `doc.go` ✓
- `internal/algo/algo.go` ✓

**Commits exist:**
- `1d53084` — feat(01-02): define exported data model types ✓
- `ac50fee` — feat(01-02): define functional options and internal algo interface ✓
