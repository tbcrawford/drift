---
phase: "22"
plan: "02"
subsystem: cmd/drift
tags: [go-git, testing, test-infrastructure, refactor]
dependency_graph:
  requires: [22-01]
  provides: [go-git-test-infra]
  affects: [cmd/drift/gitworktree_test.go, cmd/drift/dirwalk_test.go, cmd/drift/input_test.go, cmd/drift/main_test.go]
tech_stack:
  added: []
  patterns: [git.PlainInit, wt.Add, wt.Commit, makeTestRepo helper]
key_files:
  created: []
  modified:
    - cmd/drift/gitworktree_test.go
    - cmd/drift/dirwalk_test.go
    - cmd/drift/input_test.go
    - cmd/drift/main_test.go
decisions:
  - "makeTestRepo() shared helper creates real on-disk repos using go-git PlainInit"
  - "testSig() provides fixed author signature for deterministic commits"
  - "dirwalk_test.go and input_test.go and main_test.go also required rewriting (deviation Rule 3)"
  - "All fake PATH manipulation (prependPath) and writeFakeGit helpers removed"
metrics:
  duration: 8
  completed: "2026-04-02"
  tasks: 2
  files: 4
---

# Phase 22 Plan 02: Rewrite gitworktree_test.go with go-git Test Repos Summary

**One-liner:** Replaced fake-git shell script test infrastructure with real go-git on-disk repos using PlainInit + wt.Commit, eliminating PATH manipulation and subprocess mocking.

## What Was Built

- **`cmd/drift/gitworktree_test.go`** rewritten — uses `makeTestRepo()` helper for real go-git repos
- **`cmd/drift/dirwalk_test.go`** rewritten — gitignore filtering tests use real repos with `.gitignore` committed
- **`cmd/drift/input_test.go`** rewritten — single-arg git mode test uses real repo
- **`cmd/drift/main_test.go`** rewritten — `TestRunCLI_gitSingleArg_differs` uses real repo
- `writeFakeGit` and `prependPath` helpers deleted from all test files
- All 286 tests pass across 16 packages; `go mod tidy` produces no changes

## Tasks Completed

| # | Task | Commit | Files |
|---|------|--------|-------|
| 1 | Rewrite gitworktree_test.go with go-git test repos | f0efe7a | gitworktree_test.go, dirwalk_test.go, input_test.go, main_test.go |
| 2 | Full module test verification and tidy | (no new files) | — |

## Decisions Made

- **`makeTestRepo(t, files map[string]string)`** shared helper in `gitworktree_test.go` (package-level, accessible to all test files in `package main`) — creates real git repos in temp dirs with PlainInit → WriteFile → wt.Add → wt.Commit
- **`testSig()`** returns a fixed `*object.Signature` for deterministic commits
- **Empty repo test** for `missingHEADBlob` uses bare `git.PlainInit` without any commits — `gitRevParseIsInsideWorkTree` returns `"true"` (repo exists), but `gitShowHEADBlob` returns `""` (no HEAD ref)

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] dirwalk_test.go, input_test.go, main_test.go also referenced writeFakeGit/prependPath**
- **Found during:** Task 1 (first test run)
- **Issue:** Three additional test files (`dirwalk_test.go`, `input_test.go`, `main_test.go`) referenced `writeFakeGit` and `prependPath` helpers that were removed from `gitworktree_test.go`; build failed with "undefined: writeFakeGit" errors
- **Fix:** Rewrote all three files to use real go-git repos (same pattern as `gitworktree_test.go`) or plain temp dirs where no git repo is needed
- **Files modified:** `cmd/drift/dirwalk_test.go`, `cmd/drift/input_test.go`, `cmd/drift/main_test.go`
- **Commit:** f0efe7a

## Self-Check: PASSED

- [x] `cmd/drift/gitworktree_test.go` has no `writeFakeGit` or `prependPath` functions
- [x] `cmd/drift/gitworktree_test.go` imports `github.com/go-git/go-git/v5`
- [x] `go test ./...` passes (exit 0) for all packages — 286 tests
- [x] No fake shell script manipulation of PATH in any test file
- [x] `go mod tidy` produces no changes (module is clean)
- [x] `go build ./...` succeeds
