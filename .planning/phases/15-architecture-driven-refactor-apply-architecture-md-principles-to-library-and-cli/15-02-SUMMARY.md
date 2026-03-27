---
phase: 15-architecture-driven-refactor-apply-architecture-md-principles-to-library-and-cli
plan: "02"
subsystem: cli
tags: [cobra, iostreams, global-state, refactor, architecture]

requires:
  - phase: 15-01
    provides: "IOStreams type + System() constructor; rootFlags + rootOptions structs; resolveRootOptions()"

provides:
  - "newRootCmd(streams IOStreams) constructor — no shared global cobra command"
  - "runRoot(opts *rootOptions) thin orchestrator — no flag parsing, pure orchestration"
  - "runCLI(streams IOStreams, args []string) — single injected IOStreams, fully testable"
  - "Zero package-level mutable variables in cmd/drift/main.go"
  - "Zero init() calls in cmd/drift/main.go"

affects: [future CLI plans, integration tests, testscript tests]

tech-stack:
  added: []
  patterns:
    - "newRootCmd constructor pattern — cobra command is local, not global"
    - "RunE two-liner: resolveRootOptions(flags, streams, args) → runRoot(opts)"
    - "runCLI accepts IOStreams value type — buffer swap replaces global mutation"

key-files:
  created: []
  modified:
    - cmd/drift/main.go
    - cmd/drift/main_test.go

key-decisions:
  - "runCLI signature changed from (stdout, stderr io.Writer, stdin io.Reader, args) to (streams IOStreams, args) — all tests updated accordingly"
  - "TestHelpListsAllFlags restructured to use runCLI(streams, []string{'--help'}) instead of global rootCmd.Execute() — eliminates last reference to removed global"
  - "validateRootArgs removed from main.go — validation fully handled by resolveRootOptions in flags.go (args conflict check lives there)"

patterns-established:
  - "No-global-cobra: newRootCmd() produces a fresh cobra.Command per invocation — shared state leaks between test runs become impossible"
  - "IOStreams injection: runCLI accepts IOStreams as first param — tests use buffer swap, no global mutation needed"

requirements-completed: [ARCH-01, ARCH-02, ARCH-03, ARCH-04]

duration: 4min
completed: "2026-03-27"
---

# Phase 15 Plan 02: Eliminate Global State and init() from cmd/drift/main.go Summary

**Rewrote cmd/drift/main.go to eliminate var stdinReader, var rootCmd, func init(), and buildDriftOptions() — replaced with newRootCmd(IOStreams) constructor, runRoot(opts) thin orchestrator, and runCLI(IOStreams, args) — all 219 tests pass with zero structural regressions**

## Performance

- **Duration:** 4 min
- **Started:** 2026-03-27T17:44:21Z
- **Completed:** 2026-03-27T17:48:29Z
- **Tasks:** 1
- **Files modified:** 2

## Accomplishments

- Eliminated all package-level mutable variables (`var stdinReader`, `var rootCmd`) from main.go
- Removed `func init()` flag registration and `buildDriftOptions()` function — both replaced by IOStreams-aware infrastructure from Plan 01
- `newRootCmd(streams IOStreams)` constructs a fresh cobra.Command per invocation, making repeated test calls safe without flag-reset hacks
- `runRoot(opts *rootOptions)` is a pure thin orchestrator: `resolveInputs` → `drift.Diff` → `drift.RenderWithNames`, no flag parsing
- `RunE` is now the canonical two-liner: `resolveRootOptions(flags, streams, args)` + `runRoot(opts)`
- Updated `main_test.go` to use new `runCLI(IOStreams, args)` signature; `TestHelpListsAllFlags` now calls `runCLI` instead of global `rootCmd.Execute()`
- All 219 tests pass; `go vet ./cmd/drift/...` clean

## Task Commits

Each task was committed atomically:

1. **Task 1: Rewrite cmd/drift/main.go — eliminate globals and init()** - `65f1da6` (feat)

**Plan metadata:** _(to be added by final docs commit)_

## Files Created/Modified

- `cmd/drift/main.go` — Complete rewrite: removed global state/init/buildDriftOptions; added newRootCmd/runRoot/runCLI with IOStreams injection
- `cmd/drift/main_test.go` — Updated all runCLI call sites to new IOStreams signature; restructured TestHelpListsAllFlags

## Decisions Made

- `runCLI` signature changed from `(stdout, stderr io.Writer, stdin io.Reader, args []string)` to `(streams IOStreams, args []string)` — cleaner, consistent with IOStreams-first architecture
- `TestHelpListsAllFlags` restructured to use `runCLI(streams, []string{"--help"})` — eliminates the last reference to the removed global `rootCmd`
- `validateRootArgs` removed entirely — validation is fully handled by `resolveRootOptions` in flags.go

## Deviations from Plan

None — plan executed exactly as written.

## Issues Encountered

None — the Plan 01 infrastructure (IOStreams, rootFlags, rootOptions, resolveRootOptions) was fully in place, making the main.go rewrite straightforward.

## Next Phase Readiness

- Phase 15 complete: ARCHITECTURE.md principles fully applied to cmd/drift
- No global state, no init() for anything other than blank imports, all output via injected IOStreams
- CLI is now fully testable at the unit level without subprocess spawning
- No blockers for subsequent phases

## Self-Check: PASSED

- ✅ `cmd/drift/main.go` — exists
- ✅ `cmd/drift/main_test.go` — exists
- ✅ `15-02-SUMMARY.md` — exists
- ✅ Commit `65f1da6` — verified in git log

---
*Phase: 15-architecture-driven-refactor-apply-architecture-md-principles-to-library-and-cli*
*Completed: 2026-03-27*
