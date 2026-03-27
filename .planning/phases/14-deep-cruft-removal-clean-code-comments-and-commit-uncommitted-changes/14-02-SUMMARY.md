---
phase: 14-deep-cruft-removal-clean-code-comments-and-commit-uncommitted-changes
plan: "02"
subsystem: highlight
tags: [go, chroma, diffcolors, dead-code-removal, gitignore]

# Dependency graph
requires:
  - phase: 14-01
    provides: "golangci-lint config, comment cleanup, gutterColumnSeparator unicode fix"
provides:
  - "diffcolors.go with DiffLineMutedBackgroundColour removed (zero-caller dead export)"
  - "Root-anchored .gitignore entry for /drift binary"
affects:
  - future-highlight-callers
  - any-phase-using-diffcolors

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Dead exported functions identified via grep and removed promptly to keep internal API surface minimal"

key-files:
  created: []
  modified:
    - internal/highlight/diffcolors.go
    - .gitignore

key-decisions:
  - "DiffLineMutedBackgroundColour removed — had zero callers since phase 11 terrasort parity refactor"
  - ".gitignore: drift → /drift (root-anchored) to prevent accidental suppression of the drift/ library subdirectory"

patterns-established:
  - "Check for zero-caller exported functions after each refactor; remove promptly rather than keeping as 'reserved'"

requirements-completed:
  - CRUFT-02

# Metrics
duration: 1min
completed: "2026-03-27"
---

# Phase 14 Plan 02: Remove Dead DiffLineMutedBackgroundColour Summary

**Dead exported function `DiffLineMutedBackgroundColour` removed from `internal/highlight/diffcolors.go`; `.gitignore` root-anchored to prevent drift/ library suppression**

## Performance

- **Duration:** ~1 min
- **Started:** 2026-03-27T14:49:21Z
- **Completed:** 2026-03-27T14:50:18Z
- **Tasks:** 2
- **Files modified:** 2

## Accomplishments
- Removed `DiffLineMutedBackgroundColour` (lines 185–196 of diffcolors.go) — dead since the phase 11 terrasort parity refactor replaced it with a direct `DiffLineBackgroundColour` call path
- Confirmed `terminalBaseRGB`, `blendColourTowardRGB`, and `WordSpanBackgroundColour` are retained (still have active callers)
- Fixed `.gitignore` to use `/drift` instead of `drift` — prevents the broad pattern from suppressing the `drift/` library subdirectory

## Task Commits

Each task was committed atomically:

1. **Task 1+2: Remove DiffLineMutedBackgroundColour + verify + .gitignore fix** - `793775c` (refactor)

**Plan metadata:** *(docs commit below)*

## Files Created/Modified
- `internal/highlight/diffcolors.go` - Removed dead `DiffLineMutedBackgroundColour` export (12 lines removed)
- `.gitignore` - Root-anchored binary ignore: `drift` → `/drift`

## Decisions Made
- `DiffLineMutedBackgroundColour` confirmed zero callers via `grep -rn` across all `.go` files; safe to delete without backward-compat concerns (internal package, not public API)
- `.gitignore` fix included in this plan commit because it was flagged in wave 1 context and is a housekeeping item with zero risk

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 2 - Housekeeping] Root-anchor drift binary in .gitignore**
- **Found during:** Task 1 (pre-task environment check per wave 1 context note)
- **Issue:** `.gitignore` contained `drift` (matches any path containing "drift") instead of `/drift` (root-only binary). After phase 13 created the `drift/` library subdirectory, the broad pattern risks suppressing that directory in some git tooling contexts.
- **Fix:** Changed line 2 of `.gitignore` from `drift` to `/drift`
- **Files modified:** `.gitignore`
- **Verification:** File reviewed post-edit; both `go build ./...` and `go test ./...` unaffected
- **Committed in:** `793775c` (combined with main task)

---

**Total deviations:** 1 auto-fixed (housekeeping - .gitignore root-anchor)
**Impact on plan:** Minor housekeeping item flagged explicitly in wave 1 context; no scope creep.

## Issues Encountered
None — function removal was straightforward; build and all 219 tests passed immediately.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Phase 14 complete: all 2 plans executed
- Internal API surface is cleaner; `diffcolors.go` exports only functions with active callers
- `.gitignore` is now safe with the `drift/` library subdirectory present

---
*Phase: 14-deep-cruft-removal-clean-code-comments-and-commit-uncommitted-changes*
*Completed: 2026-03-27*
