---
phase: 16-fix-v1-blockers-hirschberg-myers-context-validation-goreleaser
plan: "03"
subsystem: distribution
tags: [goreleaser, distribution, cli, release-engineering]
dependency_graph:
  requires: []
  provides: [goreleaser-config]
  affects: [cmd/drift]
tech_stack:
  added: [goreleaser v2.15.0]
  patterns: [multi-platform cross-compile, CGO_ENABLED=0 static binaries, ldflags version injection]
key_files:
  created: [.goreleaser.yaml]
  modified: [.gitignore]
decisions:
  - "goreleaser v2 config uses formats (plural array) not format (singular) for archives"
  - "snapshot.version_template used (not deprecated snapshot.name_template)"
  - "linux/arm64 and windows/arm64 excluded from initial release pending CI arm64 runners"
  - "git remote origin added to repo to satisfy goreleaser SCM detection"
metrics:
  duration: "147s"
  completed_date: "2026-03-31"
  tasks_completed: 4
  files_modified: 2
---

# Phase 16 Plan 03: goreleaser Multi-Platform Release Config Summary

**One-liner:** goreleaser v2 config with darwin/amd64+arm64, linux/amd64, windows/amd64 cross-compile, sha256 checksums, and platform-appropriate archive formats.

## What Was Built

Added `.goreleaser.yaml` at the repository root to enable multi-platform binary releases of the `drift` CLI via goreleaser v2.

### Key Deliverables

- **`.goreleaser.yaml`** — goreleaser v2 config producing binaries for 4 platforms (darwin/amd64, darwin/arm64, linux/amd64, windows/amd64) with `CGO_ENABLED=0`, version injection via ldflags, tar.gz for Unix / zip for Windows archives, and sha256 checksums
- **`.gitignore`** — added `/dist/` to prevent goreleaser's build output from being committed

## Tasks Completed

| Task | Name | Commit | Files |
|------|------|--------|-------|
| 1 | Create .goreleaser.yaml | ea5c6d8 | .goreleaser.yaml |
| 2 | Validate config with goreleaser check | ea5c6d8 | (validation only) |
| 3 | Run goreleaser build --snapshot | ea5c6d8 | (verification only) |
| 4 | Ensure dist/ is in .gitignore | c4a2c7e | .gitignore |

## Verification Results

- `goreleaser check` → passes with 0 ERRO lines
- `goreleaser build --snapshot --clean` → builds succeed for all 4 platforms in ~6s
- `go test ./...` → 104 tests pass across 16 packages, no regressions
- `dist/` directory properly git-ignored

## Decisions Made

| Decision | Rationale |
|----------|-----------|
| linux/arm64 and windows/arm64 excluded | No CI arm64 runners yet; can be added when runners are available |
| `formats: [tar.gz]` (plural array) | goreleaser v2 deprecated singular `format` key |
| `snapshot.version_template` | goreleaser v2 deprecated `snapshot.name_template` |
| git remote origin added | goreleaser check requires a remote for SCM detection; URL matches module path |
| CGO_ENABLED=0 | Ensures fully static binaries that work on target platforms without libc |

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] git remote required for goreleaser check**
- **Found during:** Task 2
- **Issue:** `goreleaser check` failed with "no remote configured to list refs from" — the repo had no git remote set up
- **Fix:** Added `git remote add origin https://github.com/tbcrawford/drift.git` (matches module path `github.com/tbcrawford/drift`)
- **Files modified:** (git config only, no committed files)

## Self-Check

### Files exist:
- `.goreleaser.yaml` ✓
- `.gitignore` (updated) ✓

### Commits exist:
- `ea5c6d8` (Task 1: .goreleaser.yaml) ✓
- `c4a2c7e` (Task 4: .gitignore update) ✓

## Self-Check: PASSED
