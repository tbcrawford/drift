---
phase: 17-address-medium-priority-council-review-issues
status: pending
plans_complete: 0/5
---

# Phase 17 Verification: Address Medium-Priority Council Review Issues

**Phase Goal:** Resolve all five medium-priority issues from the council review
(`.reviews/drift-library/REVIEW.md`) — library-to-root migration, `Line.Spans`
removal, golden file tests, bottom-aligned split pairing, and dual term dep
cleanup.

**Verdict: PENDING**

---

## Must-Have Truths (All Plans)

### Plan 01 — Library-to-Root Migration

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | All 13 library files exist at repo root (moved from drift/ via git mv) | PENDING | — |
| 2 | drift/ subdirectory is removed | PENDING | — |
| 3 | cmd/drift/go.mod exists as a separate module: module github.com/tylercrawford/drift/cmd/drift | PENDING | — |
| 4 | go.work at repo root uses . and ./cmd/drift | PENDING | — |
| 5 | Zero files import "github.com/tylercrawford/drift/drift" | PENDING | — |
| 6 | All 223 existing library tests pass | PENDING | — |
| 7 | CLI tests pass (cd cmd/drift && go test ./...) | PENDING | — |

### Plan 02 — Line.Spans Removal

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | Line struct in internal/edittype/edittype.go has no Spans []Span field | PENDING | — |
| 2 | Span type still exists in internal/edittype/edittype.go (internal-only) | PENDING | — |
| 3 | types.go (repo root) does not export Span as a public type alias | PENDING | — |
| 4 | No callers reference .Spans anywhere in the codebase | PENDING | — |
| 5 | All 223 tests pass | PENDING | — |

### Plan 03 — Golden File Tests

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | github.com/sebdah/goldie/v2 v2.8.0 in go.mod | PENDING | — |
| 2 | golden_test.go at repo root with at least 3 snapshot tests | PENDING | — |
| 3 | testdata/golden/ directory with unified_go.golden, split_go.golden, nocolor_basic.golden | PENDING | — |
| 4 | TestGolden_ tests pass without -update | PENDING | — |
| 5 | Golden fixture files contain no ANSI escape codes | PENDING | — |

### Plan 04 — pairHunkLines Bottom-Aligned Pairing

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | 3del/1ins: delete[0] and delete[1] get blank right-side placeholders; delete[2] pairs with insert[0] | PENDING | — |
| 2 | 1del/3ins: insert[0] and insert[1] get blank left-side placeholders; delete[0] pairs with insert[2] | PENDING | — |
| 3 | Equal-size blocks: positional pairing preserved (delete[i] pairs with insert[i]) | PENDING | — |
| 4 | TestPairHunkLines_BottomAligned all subtests pass | PENDING | — |
| 5 | All existing split renderer tests pass | PENDING | — |

### Plan 05 — Dual Term Dependencies

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | Only one of golang.org/x/term or charmbracelet/x/term is a direct dep in go.mod | PENDING | — |
| 2 | The retained direct dep is the one actually imported by project source files | PENDING | — |
| 3 | The removed package is absent or appears only as // indirect | PENDING | — |
| 4 | go mod tidy produces no further changes | PENDING | — |
| 5 | All 223+ tests pass | PENDING | — |

---

## Regression Gate

| Check | Status | Evidence |
|-------|--------|----------|
| go test ./... passes (library module) | PENDING | — |
| cd cmd/drift && go test ./... passes | PENDING | — |
| go vet ./... passes | PENDING | — |
| go build ./... passes | PENDING | — |
| go install github.com/tylercrawford/drift/cmd/drift@latest works | PENDING | — |
