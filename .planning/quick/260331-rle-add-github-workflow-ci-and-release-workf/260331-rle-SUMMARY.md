---
quick_id: 260331-rle
slug: add-github-workflow-ci-and-release-workf
description: add github workflow ci and release workflows
date: 2026-03-31
status: complete
commit: 06637c3
tags: [ci, github-actions, goreleaser, govulncheck, golangci-lint]
key-files:
  created:
    - .github/workflows/ci.yml
    - .github/workflows/release.yml
    - .github/workflows/security.yml
    - dist/config.yaml
  modified:
    - .gitignore
decisions:
  - golangci-lint runs only on stable Go matrix leg (skip 1.21.x to avoid version noise)
  - goreleaser --config dist/config.yaml explicit flag to override default dist/ path
  - .gitignore updated to /dist/* + !/dist/config.yaml so goreleaser build output excluded but config tracked
  - release.yml includes permissions.contents:write for GitHub Release asset upload
duration: ~8 min
---

# Quick Task 260331-rle: Add GitHub Workflow CI and Release Workflows — Summary

**One-liner:** GitHub Actions CI (matrix Go test + golangci-lint), goreleaser release on version tags, and weekly govulncheck security scan via three production-ready workflow files.

## What Was Done

Created three GitHub Actions workflows covering the full development → release → security lifecycle for the `drift` project:

### Task 1 — `.github/workflows/ci.yml`

Matrix CI across Go `1.21.x` (minimum compat) and `stable` (latest) on `ubuntu-latest`. Steps:
1. `actions/checkout@v4`
2. `actions/setup-go@v5` with `cache: true`
3. `go vet ./...` — static analysis
4. `go test -race ./...` — library tests via `go.work`
5. `go test -race ./...` in `working-directory: cmd/drift` — CLI sub-module explicit
6. `golangci-lint-action@v6` — conditional on `matrix.go == 'stable'` to avoid 1.21.x linter noise

### Task 2 — `.github/workflows/release.yml`

Triggers on `push` to `v*.*.*` tags. Steps:
1. `actions/checkout@v4` with `fetch-depth: 0` (goreleaser needs full tag history for changelog)
2. `actions/setup-go@v5` with `cache: true`
3. `goreleaser/goreleaser-action@v6` with `args: release --clean --config dist/config.yaml` and `GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}`

Includes `permissions: contents: write` at job level for GitHub Release asset upload.

### Task 3 — `.github/workflows/security.yml`

Weekly (`0 9 * * 1` — every Monday 09:00 UTC) + `push` to `main` triggers. Steps:
1. `actions/checkout@v4`
2. `actions/setup-go@v5`
3. `golang/govulncheck-action@v1` scanning `./...`

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 2 - Missing] Track dist/config.yaml in git**
- **Found during:** Staging files for commit
- **Issue:** `.gitignore` had `/dist/` which excluded the whole directory, so `dist/config.yaml` (goreleaser config source file) was never committed to the repo
- **Fix:** Updated `.gitignore` from `/dist/` to `/dist/*` + `!/dist/config.yaml` — build output still ignored, config source file now tracked
- **Files modified:** `.gitignore`, `dist/config.yaml` (now tracked)
- **Commit:** 06637c3

**2. [Rule 1 - Bug] Fix goreleaser owner: tylercrawford → tbcrawford**
- **Found during:** Reading dist/config.yaml (specified in constraints)
- **Issue:** `release.github.owner: tylercrawford` was incorrect — GitHub org/user is `tbcrawford`
- **Fix:** Changed to `owner: tbcrawford` in `dist/config.yaml`
- **Files modified:** `dist/config.yaml`
- **Commit:** 06637c3

## Files Created / Modified

| File | Status | Purpose |
|------|--------|---------|
| `.github/workflows/ci.yml` | Created | Matrix CI — test + vet + lint on push/PR |
| `.github/workflows/release.yml` | Created | GoReleaser release on version tags |
| `.github/workflows/security.yml` | Created | Weekly govulncheck vulnerability scan |
| `dist/config.yaml` | Created (now tracked) | GoReleaser config with owner fix |
| `.gitignore` | Modified | Allow dist/config.yaml through gitignore |

## Commit

```
06637c3  ci: add GitHub Actions workflows for CI, release, and security
```

Single atomic commit as required, covering all 5 file changes.

## Self-Check

- [x] `.github/workflows/ci.yml` exists and is valid YAML with matrix strategy
- [x] `.github/workflows/release.yml` exists, triggers on `v*.*.*` tags, passes `GITHUB_TOKEN`
- [x] `.github/workflows/security.yml` exists, has scheduled cron trigger
- [x] `dist/config.yaml` has `owner: tbcrawford` (verified via grep)
- [x] Commit `06637c3` exists in git log
- [x] golangci-lint only runs on `stable` Go matrix leg
- [x] goreleaser action passes `--config dist/config.yaml`
- [x] `fetch-depth: 0` set in release workflow for full tag history
- [x] `GITHUB_TOKEN` uses `secrets.GITHUB_TOKEN` (built-in, no custom secret)
