---
phase: 15-architecture-driven-refactor-apply-architecture-md-principles-to-library-and-cli
plan: "01"
subsystem: cmd/drift
tags: [cli, iostreams, flags, options, architecture, refactor]
dependency_graph:
  requires: []
  provides: [IOStreams, rootFlags, rootOptions, resolveRootOptions]
  affects: [cmd/drift/main.go]
tech_stack:
  added: []
  patterns: [Flags → Options → run() lifecycle, IOStreams injection]
key_files:
  created:
    - cmd/drift/iostreams.go
    - cmd/drift/flags.go
  modified: []
decisions:
  - "IOStreams defined in its own file (iostreams.go) for focused single-responsibility contract"
  - "rootFlags maps 1:1 to init() flag definitions — no business logic, just storage"
  - "resolveRootOptions centralizes all I/O decisions including show-theme stderr callback using streams.Err"
  - "parseAlgorithm reused from main.go — no duplication needed since same package"
metrics:
  duration: "69s"
  completed: "2026-03-27T17:42:11Z"
  tasks_completed: 2
  files_created: 2
  files_modified: 0
---

# Phase 15 Plan 01: IOStreams and Flags → Options Lifecycle Contracts Summary

**One-liner:** IOStreams I/O abstraction and rootFlags/rootOptions/resolveRootOptions lifecycle types added as new files, establishing the architecture contracts Plan 02 will wire in.

## What Was Built

Two new files that define the architectural contracts for the drift CLI refactor:

1. **`cmd/drift/iostreams.go`** — `IOStreams` struct holding `In`, `Out`, `Err` channels plus `System()` constructor that wires to real OS I/O. Establishes the contract that no code below `main()` accesses `os.Std*` directly.

2. **`cmd/drift/flags.go`** — Three types forming the Flags → Options lifecycle:
   - `rootFlags`: raw cobra-parsed values, one field per flag, no logic
   - `rootOptions`: fully resolved values ready for execution (streams + drift.Option slice)
   - `resolveRootOptions()`: the single place all I/O decisions are made, including routing `show-theme` output through `streams.Err` instead of `os.Stderr` directly

## Tasks Completed

| Task | Name | Commit | Files |
|------|------|--------|-------|
| 1 | Create cmd/drift/iostreams.go | c5624b2 | cmd/drift/iostreams.go |
| 2 | Create cmd/drift/flags.go | 38ec3c4 | cmd/drift/flags.go |

## Verification

```
✓ go build ./cmd/drift/...   — clean
✓ go test ./cmd/drift/...    — 20 passed
✓ go test ./...              — 219 passed (no regressions)
✓ go vet ./cmd/drift/...     — no issues
```

## Decisions Made

1. **`IOStreams` in its own file** — single-responsibility; the contract is a focused type with one purpose (I/O abstraction).

2. **`rootFlags` maps 1:1 to `init()` flags** — fields mirror the exact cobra flag names from `main.go`'s `init()`, making Plan 02's wiring mechanical.

3. **`resolveRootOptions` owns all decisions** — algorithm parsing, context validation, option building, and the `show-theme` stderr callback all live here. Nothing is deferred to `runRoot()`.

4. **`streams.Err` for show-theme output** — previously `fmt.Fprintf(os.Stderr, ...)` in `buildDriftOptions()`. Now uses the injected stream, making it testable with a buffer swap.

5. **No duplication of `parseAlgorithm`** — it already exists in `main.go` and is accessible within the same `package main`, so `flags.go` reuses it directly.

## Deviations from Plan

None — plan executed exactly as written.

## What's Next

Plan 02 will wire these types into `main.go`:
- Replace `init()` flag registration with a `rootFlags` struct wired to cobra
- Replace `buildDriftOptions(cmd)` with `resolveRootOptions(flags, streams, args)`  
- Replace `runRoot(cmd, args)` with `runRoot(opts *rootOptions)`
- Eliminate `stdinReader` global variable
- Eliminate direct `os.Stderr` writes

## Self-Check

- [x] `cmd/drift/iostreams.go` exists and exports `IOStreams` and `System()`
- [x] `cmd/drift/flags.go` exists and exports `rootFlags`, `rootOptions`, `resolveRootOptions`
- [x] Commit `c5624b2` exists (Task 1)
- [x] Commit `38ec3c4` exists (Task 2)
- [x] All 219 tests pass
