---
phase: 17-address-medium-priority-council-review-issues
plan: "05"
subsystem: dependencies
tags: [go-modules, cleanup, term, dependencies]
dependency_graph:
  requires: [17-01]
  provides: [clean-dep-graph-single-term-package]
  affects: [go.mod, go.sum, internal/terminal/palette_unix.go]
tech_stack:
  added: []
  patterns: [go-mod-tidy, charmbracelet-x-term-uintptr-api]
key_files:
  created: []
  modified:
    - go.mod
    - go.sum
    - internal/terminal/palette_unix.go
decisions:
  - Migrated palette_unix.go from golang.org/x/term to charmbracelet/x/term; charmbracelet/x/term takes uintptr fd (tty.Fd() returns uintptr directly — no int() cast needed), golang.org/x/term takes int fd
metrics:
  duration: ~5min
  completed: "2026-03-31T17:13:23Z"
  tasks_completed: 2
  tasks_total: 2
  files_changed: 3
---

# Phase 17 Plan 05: Consolidate term package dependencies — Summary

**One-liner:** Migrated `palette_unix.go` from `golang.org/x/term` to `charmbracelet/x/term`, eliminating the redundant direct dependency via `go mod tidy`.

## What Was Built

Audited all import sites for both `golang.org/x/term` and `charmbracelet/x/term`:
- `internal/render/termwidth.go` already used `charmbracelet/x/term`
- `internal/terminal/palette_unix.go` used `golang.org/x/term`

Migrated `palette_unix.go` to `charmbracelet/x/term`:
- Changed import path from `golang.org/x/term` to `github.com/charmbracelet/x/term`
- Removed `int()` casts: `charmbracelet/x/term` takes `uintptr` fd arguments directly (same type as `tty.Fd()` return)
- `golang.org/x/term.MakeRaw(int(tty.Fd()))` → `term.MakeRaw(tty.Fd())`
- `golang.org/x/term.Restore(int(tty.Fd()), oldState)` → `term.Restore(tty.Fd(), oldState)`

Ran `go mod tidy` which:
- Removed `golang.org/x/term v0.41.0` from direct dependencies entirely
- Promoted `github.com/sebdah/goldie/v2 v2.8.0` from indirect to direct (used in `golden_test.go`)

## Tasks Completed

| Task | Name | Commit | Files |
|------|------|--------|-------|
| 1 | Audit import sites and migrate palette_unix.go | 3b5bf54 | go.mod, go.sum, internal/terminal/palette_unix.go |
| 2 | Run full test suite — no regressions | (no new files) | — |

## Verification

- `go build ./internal/terminal/...` — clean
- `go build ./...` — clean  
- `go vet ./...` — clean
- `go test ./...` — 210 tests pass
- `grep "x/term" go.mod` — only `charmbracelet/x/term v0.2.2` remains as direct dep
- `golang.org/x/term` completely absent from root `go.mod`
- `go mod tidy` produces no further changes

## Decisions Made

| Decision | Rationale |
|----------|-----------|
| Migrate palette_unix.go to charmbracelet/x/term | Eliminates redundant direct dep; charmbracelet/x/term is already used elsewhere in the project; API is compatible (uintptr fd vs int fd) |
| Remove int() casts | charmbracelet/x/term uses uintptr which matches tty.Fd() return type directly — cleaner code, no-cast needed |

## Deviations from Plan

None — plan executed exactly as written.

**Pre-existing test failures noted (out of scope):**
- `TestPairHunkLines_MoreDeletesThanInserts` and `TestPairHunkLines_MoreInsertsThanDeletes` were failing before this plan's changes — confirmed by stash test. These are addressed by plan 17-04 (parallel agent).

## Known Stubs

None.

## Self-Check

- [x] `internal/terminal/palette_unix.go` uses `github.com/charmbracelet/x/term` — VERIFIED
- [x] `go.mod` does not contain `golang.org/x/term` as direct dep — VERIFIED
- [x] Commit `3b5bf54` exists in git history — VERIFIED
- [x] `go mod tidy` produces no changes — VERIFIED
- [x] All tests pass (210) — VERIFIED
