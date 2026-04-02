---
phase: 21-respect-gitignore-rules-where-necessary
plan: "01"
subsystem: cli/diff-walkers
tags: [gitignore, directory-diff, git-integration, cli]
dependency_graph:
  requires: [phase-20-directory-diff]
  provides: [gitignore-aware-directory-diff, filterGitIgnored-helper]
  affects: [cmd/drift/gitworktree.go, cmd/drift/dirwalk.go]
tech_stack:
  added: []
  patterns: [git-check-ignore-stdin, fail-open-git-integration, per-side-gitignore-filtering]
key_files:
  modified:
    - cmd/drift/gitworktree.go
    - cmd/drift/gitworktree_test.go
    - cmd/drift/dirwalk.go
    - cmd/drift/dirwalk_test.go
decisions:
  - filterGitIgnored uses NUL-separated stdin/stdout with git check-ignore -z --stdin for robust path handling
  - workSet rebuilt after filtering in gitDirectoryVsHEAD so deleted-file detection excludes ignored paths
  - isInsideGitRepo added to dirwalk.go for per-side gitignore detection in diffDirectories
  - Fail-open on all git errors (not found, not a repo, unexpected exit) — walk all files
  - newDir entries collected as slice then filtered before pair construction (avoids mid-walk mutation)
metrics:
  duration: 213s
  completed: "2026-04-02"
  tasks_completed: 3
  files_modified: 4
requirements_satisfied: [GITIGNORE-01, GITIGNORE-02]
---

# Phase 21 Plan 01: Gitignore-Aware Directory Diff Summary

**One-liner:** Added `filterGitIgnored` helper using `git check-ignore -z --stdin` and wired it into both `gitDirectoryVsHEAD` and `diffDirectories` for fail-open, per-side gitignore filtering.

## What Was Built

### filterGitIgnored (cmd/drift/gitworktree.go)

A shared helper that runs `git check-ignore -z --stdin` in a given repo root and returns the input paths with any gitignored entries removed. NUL-delimited I/O makes it correct for paths containing spaces.

**Fail-open contract:**
- Empty input → returns nil, no subprocess
- `git` not found → returns original paths unchanged
- `git check-ignore` exits 1 with no output (nothing ignored) → returns all paths
- Any unexpected exit code → returns all paths

### runGitStdin (cmd/drift/gitworktree.go)

A variant of `runGit` that feeds a string as stdin to the git subprocess, needed for the `check-ignore --stdin` protocol.

### isInsideGitRepo (cmd/drift/dirwalk.go)

A small helper that wraps `gitRevParseIsInsideWorkTree` + `gitRevParseShowToplevel` and returns the repo root if the directory is inside a worktree, or ("", false) otherwise.

### gitDirectoryVsHEAD updated

After `filepath.WalkDir` collects `workFiles`, the list is filtered through `filterGitIgnored` before the comparison loop. `workSet` (used for deleted-file detection) is rebuilt from the filtered list, so ignored paths don't block detection of truly deleted tracked files.

### diffDirectories updated

- Collects `oldFiles` map then filters through `filterGitIgnored` if `oldDir` is inside a git repo (removes ignored keys from map)
- Collects `newEntries` as a slice (instead of processing inline during walk), then filters through `filterGitIgnored` if `newDir` is inside a git repo
- Each side uses its own repo root for independent filtering
- Non-git dirs walk all files unchanged

## Tests Added

**gitworktree_test.go (4 new tests):**
- `TestFilterGitIgnored_emptyInput` — no subprocess, returns nil
- `TestFilterGitIgnored_noIgnored` — exit 1 → all paths returned
- `TestFilterGitIgnored_someIgnored` — dist/app NUL response → removed from result
- `TestFilterGitIgnored_gitNotFound` — empty PATH → fail-open
- `TestGitDirectoryVsHEAD_skipsIgnored` — dist/app ignored by fake git → absent from pairs; keep.go present

**dirwalk_test.go (4 new tests):**
- `TestDiffDirectories_gitignore_skipsIgnoredInOld` — ignored file on old side excluded
- `TestDiffDirectories_gitignore_skipsIgnoredInNew` — ignored file on new side excluded
- `TestDiffDirectories_gitignore_noRepo_walksAll` — not a git repo → all files included
- `TestDiffDirectories_gitignore_gitNotFound_walksAll` — no git binary → all files included

## Commits

| Task | Commit | Description |
|------|--------|-------------|
| Task 1 | fa4f0c4 | feat(21-01): add filterGitIgnored helper and wire into gitDirectoryVsHEAD |
| Task 2 | 29c5b39 | feat(21-01): update diffDirectories with gitignore-aware per-side filtering |

## Deviations from Plan

None — plan executed exactly as written.

## Known Stubs

None — all gitignore filtering is fully wired and tested.

## Self-Check: PASSED

- [x] `cmd/drift/gitworktree.go` exists with `filterGitIgnored` and `runGitStdin`
- [x] `cmd/drift/dirwalk.go` exists with `isInsideGitRepo` and updated `diffDirectories`
- [x] Commits fa4f0c4 and 29c5b39 exist in git log
- [x] `go test ./... -count=1` → 286 passed
- [x] `go vet ./...` → no issues
- [x] `go build ./cmd/drift/` → success
