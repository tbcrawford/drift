---
quick_id: 260331-rle
slug: add-github-workflow-ci-and-release-workf
description: add github workflow ci and release workflows
date: 2026-03-31
status: planned
---

# Quick Task 260331-rle: Add GitHub Workflow CI and Release Workflows

## Goal

Create production-quality GitHub Actions workflows for the `drift` project:
- **CI** — test, vet, lint on every push/PR
- **Release** — goreleaser on version tag push
- **Security** — dependency vulnerability scan (bonus)

## Context

- Module root: `github.com/tbcrawford/drift` (Go 1.25, single `go.mod`)
- CLI sub-module: `cmd/drift/` with its own `go.mod` + `go.work` workspace
- goreleaser config at `dist/config.yaml`
- Test command: `go test ./...` (library root) + `go test ./...` (cmd/drift)
- Lint: `golangci-lint run ./...`
- No existing `.github/` directory

## Tasks

### Task 1: Create `.github/workflows/ci.yml`

**Files:** `.github/workflows/ci.yml`

**Action:**
Create a CI workflow that runs on every `push` and `pull_request` to `main`.

Steps:
- Matrix: `ubuntu-latest` only (Go cross-compile handles other platforms at release time)
- Go versions: `['1.21.x', 'stable']` — validate minimum compat + latest
- Steps:
  1. `actions/checkout@v4`
  2. `actions/setup-go@v5` with `cache: true`
  3. `go vet ./...` (library)
  4. `go test -race ./...` (library)
  5. `cd cmd/drift && go test -race ./...` (CLI sub-module)
  6. `golangci-lint run ./...` via `golangci-lint-action@v6` (only on stable Go, skip on 1.21.x)

**Verify:** File exists, valid YAML, triggers on push/PR, uses go workspace correctly

---

### Task 2: Create `.github/workflows/release.yml`

**Files:** `.github/workflows/release.yml`

**Action:**
Create a release workflow that triggers on `push` to tags matching `v*.*.*`.

Steps:
  1. `actions/checkout@v4` with `fetch-depth: 0` (goreleaser needs full tag history)
  2. `actions/setup-go@v5` with `cache: true`
  3. `goreleaser/goreleaser-action@v6` with `distribution: goreleaser`, `version: latest`, `args: release --clean`, `GITHUB_TOKEN` env var

Config pointer: `--config dist/config.yaml` passed via args.

**Verify:** File exists, valid YAML, triggers only on version tags, passes GITHUB_TOKEN

---

### Task 3: Create `.github/workflows/security.yml`

**Files:** `.github/workflows/security.yml`

**Action:**
Create a security/dependency scan workflow using `govulncheck`:
- Trigger: weekly schedule (`0 9 * * 1`) + push to `main`
- Steps:
  1. `actions/checkout@v4`
  2. `actions/setup-go@v5`
  3. `golang/govulncheck-action@v1` — scans library and cmd/drift for known CVEs

**Verify:** File exists, valid YAML, scheduled trigger present

---

## Commit Strategy

Single atomic commit after all three files are written:
```
ci: add GitHub Actions workflows for CI, release, and security
```
