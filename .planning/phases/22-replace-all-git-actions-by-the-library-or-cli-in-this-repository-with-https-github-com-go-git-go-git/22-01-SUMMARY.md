---
phase: "22"
plan: "01"
subsystem: cmd/drift
tags: [go-git, git, dependency, refactor]
dependency_graph:
  requires: []
  provides: [go-git-gitworktree]
  affects: [cmd/drift/gitworktree.go, cmd/drift/dirwalk.go]
tech_stack:
  added: [github.com/go-git/go-git/v5 v5.17.2]
  patterns: [PlainOpenWithOptions, headCommitTree, tree.File, gitignore.NewMatcher]
key_files:
  created: []
  modified:
    - cmd/drift/gitworktree.go
    - go.mod
    - go.sum
decisions:
  - "go-git PlainOpenWithOptions(DetectDotGit=true) replaces git rev-parse subprocess"
  - "tree.Files() iterator replaces git ls-tree subprocess for deleted-file detection"
  - "gitignore.ReadPatterns + NewMatcher replaces git check-ignore subprocess"
  - "plumbing.ErrReferenceNotFound handled in gitShowHEADBlob for empty repo (no HEAD) case"
metrics:
  duration: 5
  completed: "2026-04-02"
  tasks: 2
  files: 3
---

# Phase 22 Plan 01: Add go-git Dependency and Rewrite gitworktree.go Summary

**One-liner:** Replaced all git subprocess calls with go-git library (PlainOpenWithOptions, tree.File, gitignore.NewMatcher), eliminating the runtime dependency on the git binary.

## What Was Built

- **`cmd/drift/gitworktree.go`** completely rewritten — zero `exec.Command` calls remain
- All public function signatures unchanged (`resolveGitWorkingTreeVsHEAD`, `gitDirectoryVsHEAD`, `filterGitIgnored`, `gitFilePair`)
- All internal signatures unchanged (`gitRevParseIsInsideWorkTree`, `gitRevParseShowToplevel`)
- New pure-Go helpers: `openRepoAt`, `headCommitTree`, `gitShowHEADBlob`
- Removed: `runGit`, `runGitStdin`, `gitEnv` (no longer needed)
- `go-git` promoted from indirect to direct dependency via `go mod tidy`

## Tasks Completed

| # | Task | Commit | Files |
|---|------|--------|-------|
| 1 | Add go-git dependency | 069f285 | go.mod, go.sum |
| 2 | Rewrite gitworktree.go with go-git | ddbbafb | cmd/drift/gitworktree.go, go.mod, go.sum |

## Decisions Made

- **go-git PlainOpenWithOptions(DetectDotGit=true)** replaces `git rev-parse --is-inside-work-tree` and `--show-toplevel` — single library call detects repo and returns worktree root
- **tree.Files() iterator** replaces `git ls-tree -r --name-only HEAD` — pure Go enumeration of HEAD blobs for deleted-file detection
- **gitignore.ReadPatterns + NewMatcher** replaces `git check-ignore -z --stdin` subprocess — gitignore patterns loaded from worktree filesystem directly
- **plumbing.ErrReferenceNotFound** handled in `gitShowHEADBlob` — empty repos (no HEAD commit) return `""` for blob content, not an error
- **Fail-open** preserved in `filterGitIgnored` — if repo cannot be opened, all paths returned unchanged (matches previous behavior)

## Deviations from Plan

None — plan executed exactly as written.

## Self-Check: PASSED

- [x] `cmd/drift/gitworktree.go` exists and has no `os/exec` import
- [x] `cmd/drift/gitworktree.go` imports `github.com/go-git/go-git/v5`
- [x] `go build ./...` succeeds
- [x] `runGit` and `runGitStdin` functions are gone
- [x] `go.mod` contains `github.com/go-git/go-git/v5 v5.17.2` as direct dependency
