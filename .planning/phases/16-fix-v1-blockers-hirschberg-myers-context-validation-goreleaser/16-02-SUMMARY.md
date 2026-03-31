---
phase: 16-fix-v1-blockers-hirschberg-myers-context-validation-goreleaser
plan: "02"
subsystem: drift-library
tags: [validation, options, config, testing]
dependency_graph:
  requires: []
  provides: [WithContext-validation, config-validate]
  affects: [drift/options.go, drift/drift.go, drift/options_test.go]
tech_stack:
  added: []
  patterns: [validate-on-use, functional-options, config-validation]
key_files:
  created: []
  modified:
    - drift/options.go
    - drift/drift.go
    - drift/options_test.go
    - .gitignore
decisions:
  - "validate() called at Diff() time (not WithContext() time) — standard Go functional-options pattern: validate on use, not on set"
  - "negative contextLines returns error from Diff(), not panic — consistent with existing (DiffResult, error) signature"
metrics:
  duration: "406s"
  completed: "2026-03-31T14:01:52Z"
  tasks_completed: 4
  files_modified: 4
---

# Phase 16 Plan 02: WithContext Validation Summary

**One-liner:** `validate()` method on `*config` rejects negative `WithContext` values before diff work begins, returning a descriptive error from `Diff()`.

## What Was Built

Added input validation for `WithContext(n int)` to prevent a silent bug where negative context values expanded the diff output to the entire file. The fix validates `contextLines >= 0` inside `Diff()` using a new `validate()` method on `*config`, consistent with the existing `(DiffResult, error)` return signature.

### Key Changes

**`drift/options.go`** — Added `validate()` method:
```go
func (c *config) validate() error {
    if c.diff.contextLines < 0 {
        return fmt.Errorf("drift: WithContext value must be non-negative, got %d", c.diff.contextLines)
    }
    return nil
}
```

**`drift/drift.go`** — Added validate call at top of `Diff()`:
```go
if err := cfg.validate(); err != nil {
    return DiffResult{}, err
}
```

**`drift/options_test.go`** — Added two new tests:
- `TestWithContextNegative`: confirms `Diff()` with `WithContext(-1)` returns a non-nil error containing "non-negative"  
- `TestWithContextZero`: confirms `Diff()` with `WithContext(0)` succeeds (zero context lines is valid)

**`.gitignore`** — Fixed pre-existing bug: `/drift` pattern was blocking `git add` for all `drift/` package files. Added `!/drift/` negation to allow tracked source files to be staged.

## Commits

| Task | Commit | Description |
|------|--------|-------------|
| Task 2 | c4a2c7e | feat: validate() + Diff() validate call (committed by 16-03 agent) |
| Task 3 + Rule 3 | b91d55e | fix: .gitignore + TestWithContextNegative/TestWithContextZero |

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Fixed .gitignore blocking git add for drift/ directory**
- **Found during:** Task 3 (attempting to stage drift/options_test.go)
- **Issue:** `.gitignore` entry `/drift` matched both the `drift` compiled binary and the `drift/` source directory. Git refused to stage any file in `drift/` with "The following paths are ignored by one of your .gitignore files: drift"
- **Fix:** Added `!/drift/` negation pattern after `/drift` in `.gitignore`
- **Files modified:** `.gitignore`
- **Commit:** b91d55e

**2. [Parallel execution] Task 2 changes committed by 16-03 agent**
- **Found during:** Task 3 verification
- **Issue:** A parallel agent (16-03, goreleaser) committed `drift/drift.go` and `drift/options.go` as part of `chore(16-03): add dist/ to .gitignore for goreleaser build output` (c4a2c7e) — the 16-03 agent's `git add .gitignore` window included my in-progress working tree edits.
- **Impact:** My Task 2 changes were committed (correctly) but under a different commit message. All code is correct and in the repo.
- **Resolution:** Documented here; no code changes needed.

**3. [Out of scope] myers package test failures**
- **Found during:** Task 4 full test suite run
- **Issue:** `internal/algo/myers` has 2 failing tests (`TestMyersPaperExample`, `TestCrossValidateWithSystemDiff`) from in-progress Hirschberg implementation work by parallel agent 16-01
- **Status:** Out of scope — these are plan 16-01 failures, not caused by 16-02 changes
- **Logged to:** Deferred (plan 16-01 responsibility)

## Test Results

```
drift/... → 29 passed (all drift package tests pass, including new WithContext tests)
go vet ./... → clean
```

## Known Stubs

None — all implemented functionality is fully wired.

## Self-Check

- [x] `drift/options.go` has `validate()` method with `contextLines < 0` check
- [x] `drift/drift.go` calls `cfg.validate()` and returns `DiffResult{}, err` on invalid config
- [x] `TestWithContextNegative` passes — error returned for `WithContext(-1)`
- [x] `TestWithContextZero` passes — diff succeeds for `WithContext(0)`
- [x] `.gitignore` fixed — `drift/` source files can be staged again
- [x] `go test ./drift/...` → 29 passed
- [x] `go vet ./...` → clean
- [x] Commits b91d55e and c4a2c7e exist in repo

## Self-Check: PASSED
