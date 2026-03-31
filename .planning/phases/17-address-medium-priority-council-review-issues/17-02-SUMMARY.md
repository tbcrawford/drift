---
phase: 17-address-medium-priority-council-review-issues
plan: "02"
subsystem: public-api
tags: [api-cleanup, types, public-api, v1-readiness]
dependency_graph:
  requires: [17-01]
  provides: [clean-line-struct, span-internal-only]
  affects: [types.go, internal/edittype/edittype.go]
tech_stack:
  added: []
  patterns: [type-alias-removal, zero-value-stub-removal]
key_files:
  created: []
  modified:
    - internal/edittype/edittype.go
    - types.go
decisions:
  - "Span struct retained in internal/edittype as internal-only type for future word-diff work (not deleted)"
  - "Line struct reduced to 4 fields: Op, Content, OldNum, NewNum — no stub fields at v1.0"
metrics:
  duration: "4min"
  completed: "2026-03-31T17:13:00Z"
  tasks: 2
  files: 2
---

# Phase 17 Plan 02: Remove Spans Field and Span Public Alias Summary

**One-liner:** Removed always-nil `Spans []Span` field from `Line` struct and `type Span` public alias — clean 4-field Line at v1.0 with Span retained as internal-only.

## What Was Built

Cleaned the public API surface by removing a zero-value stub (`Spans []Span`) from the `Line` struct and dropping `type Span = edittype.Span` from `types.go`. The `Span` struct itself is retained in `internal/edittype` for future word-diff work (D-06).

### Changes

**`internal/edittype/edittype.go`:**
- Removed `Spans []Span` field and its godoc comment from the `Line` struct
- Updated package-level godoc to remove `drift.Span` from the exported types list
- `Span` struct definition kept as internal-only type

**`types.go`:**
- Removed `type Span = edittype.Span` type alias and its godoc comment block
- Removed the `Line` godoc note about Spans field (no longer relevant)

## Verification

All success criteria met:
- `Line` struct has exactly 4 fields: `Op`, `Content`, `OldNum`, `NewNum`
- `Span` struct still defined in `internal/edittype` (internal-only)
- `types.go` has no `Span` type alias
- `doc.go` had no `Span` references (nothing to change)
- `grep -rn "\.Spans"` finds zero references in non-definition files
- `go build ./...` — clean
- `go vet ./...` — clean
- `go test ./...` — 210 tests pass (root module, 15 packages)
- `go test ./...` (cmd/drift) — 20 tests pass
- Total: 230 tests pass (exceeds the 223 expected)

## Commits

| Task | Commit | Description |
|------|--------|-------------|
| 1    | de9ad0f | refactor(17-02): remove Spans field from Line and Span alias from public API |
| 2    | (no files) | Verification-only — all 230 tests pass, no new files |

## Deviations from Plan

None — plan executed exactly as written.

## Known Stubs

None — this plan *removed* the only stub (always-nil `Spans` field). No new stubs introduced.

## Self-Check

- [x] `internal/edittype/edittype.go` modified — verified via Read tool
- [x] `types.go` modified — verified via Read tool
- [x] `go build ./...` clean — verified via Bash
- [x] `go test ./...` 210+20=230 tests passing — verified via Bash
- [x] `de9ad0f` commit exists — verified via `git rev-parse --short HEAD`

## Self-Check: PASSED
