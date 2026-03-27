---
phase: 12-restructure-project-layout-for-idiomatic-go-library-and-cli
plan: 12-01
subsystem: core-library
tags: [refactor, internal-layout, config, testhelpers]
dependency_graph:
  requires: []
  provides: [internal/testhelpers/apply.go, diffConfig, renderConfig]
  affects: [options.go, drift.go, render.go, drift_property_test.go, drift_algorithm_integration_test.go, options_test.go]
tech_stack:
  added: []
  patterns: [nested-config-structs, internal-test-helpers]
key_files:
  created:
    - internal/testhelpers/apply.go
  modified:
    - options.go
    - drift.go
    - render.go
    - drift_property_test.go
    - drift_algorithm_integration_test.go
    - options_test.go
  deleted:
    - testdata/apply.go
decisions:
  - "Apply() moved to internal/testhelpers to make test-only helpers clearly non-public"
  - "config struct split into diffConfig (algorithm, contextLines) + renderConfig (all render options) for self-documenting separation of concerns"
metrics:
  duration: 220s
  completed: "2026-03-27"
  tasks_completed: 2
  files_changed: 7
---

# Phase 12 Plan 01: Restructure Project Layout Summary

**One-liner:** Moved `Apply()` to `internal/testhelpers` and split flat `config` into nested `diffConfig`+`renderConfig` sub-structs for idiomatic Go library layout.

## What Was Built

Two concrete internal improvements that clean up the project layout without any public API changes:

1. **`internal/testhelpers/apply.go`** — The `Apply()` round-trip helper previously lived in the `testdata/` package, which Go treats as a special directory (fuzz corpus, txtar files). Moving it to `internal/testhelpers` makes it unambiguously a test-only internal package, not importable by external consumers.

2. **`config` struct split** — The single flat `config` struct was refactored into two named sub-structs: `diffConfig` (holds `algorithm` and `contextLines`) and `renderConfig` (holds all rendering options). This self-documents which options affect `Diff()` vs `Render()` and makes call sites in `drift.go` and `render.go` more readable via `cfg.diff.*` / `cfg.render.*` namespacing.

## Tasks Completed

| Task | Name | Commit | Files Changed |
|------|------|--------|---------------|
| 1 | Move Apply() to internal/testhelpers | 3ccc4f6 | internal/testhelpers/apply.go (+), testdata/apply.go (deleted), drift_property_test.go, drift_algorithm_integration_test.go |
| 2 | Split config struct into diffConfig + renderConfig | 2193f4b | options.go, drift.go, render.go, options_test.go |

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 2 - Missing Critical] Updated options_test.go field paths**
- **Found during:** Task 2
- **Issue:** `options_test.go` directly accessed `cfg.lang`, `cfg.theme`, and `cfg.noColor` — all now moved to `cfg.render.lang`, `cfg.render.theme`, `cfg.render.noColor` after the config struct refactor
- **Fix:** Updated all three test functions to use nested `cfg.render.*` field paths; updated doc comments to reflect the new paths
- **Files modified:** `options_test.go`
- **Commit:** 2193f4b (included in Task 2 commit)

## Verification Results

```
✅ go build ./...   — Clean
✅ go test ./...    — 219 tests passed (16 packages)
✅ go vet ./...     — No issues
✅ testdata/         — Contains only rapid/ (fuzz corpus), no apply.go
✅ internal/testhelpers/ — Contains apply.go
✅ grep testdata\.   — No matches in non-test source files
```

## Known Stubs

None — all data flows are wired. No placeholder values or TODO stubs introduced.

## Self-Check: PASSED

- `internal/testhelpers/apply.go` — EXISTS ✅
- `testdata/apply.go` — DELETED ✅
- Commit `3ccc4f6` — EXISTS ✅
- Commit `2193f4b` — EXISTS ✅
