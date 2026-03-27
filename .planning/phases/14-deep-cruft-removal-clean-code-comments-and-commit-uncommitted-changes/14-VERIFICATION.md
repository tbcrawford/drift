---
phase: 14-deep-cruft-removal-clean-code-comments-and-commit-uncommitted-changes
verified: 2026-03-27T15:10:00Z
status: gaps_found
score: 7/8 must-haves verified
re_verification: false
gaps:
  - truth: "CRUFT-01 and CRUFT-02 requirement IDs are defined in REQUIREMENTS.md"
    status: failed
    reason: "Both CRUFT-01 and CRUFT-02 are referenced in ROADMAP.md, 14-01-PLAN.md, 14-02-PLAN.md, and their SUMMARYs, but neither ID appears anywhere in .planning/REQUIREMENTS.md. They are fully orphaned — no traceability entry, no requirement definition."
    artifacts:
      - path: ".planning/REQUIREMENTS.md"
        issue: "CRUFT-01 and CRUFT-02 are absent; file ends at OSS-09 with only a single Phase 1 entry in the Traceability table"
    missing:
      - "Add CRUFT-01 requirement definition to REQUIREMENTS.md under a new 'Maintenance' section or equivalent"
      - "Add CRUFT-02 requirement definition to REQUIREMENTS.md"
      - "Add traceability entries mapping CRUFT-01 and CRUFT-02 to Phase 14 in the Traceability table"
human_verification:
  - test: "golangci-lint runs cleanly with the v2 config"
    expected: "golangci-lint run ./... exits 0 and produces no lint errors"
    why_human: "golangci-lint binary not available in the CI-less verification environment; config schema correctness is structurally verified but runtime execution requires the tool installed"
---

# Phase 14: Deep Cruft Removal — Verification Report

**Phase Goal:** Commit 6 pending working-tree changes accumulated during phases 11–13, and remove the dead exported function `DiffLineMutedBackgroundColour` from the internal highlight package.
**Verified:** 2026-03-27T15:10:00Z
**Status:** gaps_found
**Re-verification:** No — initial verification

---

## Goal Achievement

### Observable Truths

| #  | Truth                                                                     | Status     | Evidence                                                                                          |
|----|---------------------------------------------------------------------------|------------|---------------------------------------------------------------------------------------------------|
| 1  | All 6 pending modified files are committed to git                         | ✓ VERIFIED | Commit `3c27d73` staged and committed all 6 files (`.golangci.yml`, `justfile`, `main_test.go`, `diff_line.go`, `gutter.go`, `split_test.go`) |
| 2  | `go test ./...` passes with 0 failures                                    | ✓ VERIFIED | `219 passed in 16 packages` — confirmed via live run                                             |
| 3  | golangci-lint v2 config is valid (`version: "2"` key present)             | ✓ VERIFIED | `.golangci.yml` line 2: `version: "2"` confirmed; structure uses v2 `linters.exclusions` schema  |
| 4  | `justfile` has an `install:` recipe                                       | ✓ VERIFIED | `install:` recipe present at line 16 of justfile (`go install ./cmd/drift`)                      |
| 5  | `DiffLineMutedBackgroundColour` is removed from `diffcolors.go`           | ✓ VERIFIED | `grep -rn DiffLineMutedBackgroundColour --include=*.go` returns 0 matches                        |
| 6  | No references to `DiffLineMutedBackgroundColour` remain in any `.go` file | ✓ VERIFIED | Full-repo grep confirms absence across all 16 packages                                            |
| 7  | `go build ./...` succeeds after removal                                   | ✓ VERIFIED | `go build ./...` exits 0; `go vet ./...` exits clean                                             |
| 8  | CRUFT-01 and CRUFT-02 are defined in REQUIREMENTS.md                      | ✗ FAILED   | Both IDs appear in ROADMAP.md, PLANs, and SUMMARYs but are completely absent from REQUIREMENTS.md |

**Score:** 7/8 truths verified

---

### Required Artifacts

#### Plan 01 Artifacts

| Artifact                             | Expected                                        | Status      | Details                                                                                     |
|--------------------------------------|-------------------------------------------------|-------------|---------------------------------------------------------------------------------------------|
| `.golangci.yml`                      | golangci-lint v2 compatible config              | ✓ VERIFIED  | `version: "2"` at top; `linters.exclusions` block; `gosimple` absent; `formatters.exclusions` present |
| `justfile`                           | `install:` recipe                               | ✓ VERIFIED  | `install:\n    go install ./cmd/drift` found at expected position                          |
| `internal/render/gutter.go`          | `gutterColumnSeparator` constant + dim/high fg  | ✓ VERIFIED  | `const gutterColumnSeparator = " │"` (U+2502) at line 20; `dim`/`high` variables in `gutterStyleForCell` |

#### Plan 02 Artifacts

| Artifact                              | Expected                                    | Status      | Details                                                                          |
|---------------------------------------|---------------------------------------------|-------------|----------------------------------------------------------------------------------|
| `internal/highlight/diffcolors.go`    | Clean diffcolors without dead export        | ✓ VERIFIED  | 197 lines; `DiffLineMutedBackgroundColour` absent; `WordSpanBackgroundColour`, `terminalBaseRGB`, `blendColourTowardRGB` all retained |
| `.gitignore`                          | Root-anchored `/drift` binary entry         | ✓ VERIFIED  | Line 2: `/drift` (not broad `drift`); `.gitignore` fix committed in `793775c`    |

---

### Key Link Verification

#### Plan 01

| From             | To                           | Via                            | Status     | Details                                                                 |
|------------------|------------------------------|--------------------------------|------------|-------------------------------------------------------------------------|
| `.golangci.yml`  | `golangci-lint run ./...`    | `version: "2"` selects v2 schema | ✓ WIRED  | Structural schema verified; runtime check flagged for human (see below) |

#### Plan 02

| From                         | To                          | Via                        | Status     | Details                                                                                   |
|------------------------------|-----------------------------|----------------------------|------------|-------------------------------------------------------------------------------------------|
| `diffcolors.go`              | `internal/render/wordline.go` | `WordSpanBackgroundColour` | ✓ WIRED  | `wordline.go:180` calls `highlight.WordSpanBackgroundColour`; function present at line 188 |
| `diffcolors.go` (terminalBaseRGB / blendColourTowardRGB) | `WordSpanBackgroundColour` | internal call chain | ✓ WIRED | Both helpers called from `WordSpanBackgroundColour` lines 194–196; no dangling helpers after removal |

---

### Data-Flow Trace (Level 4)

Not applicable — phase 14 is a housekeeping/commit phase. No new data-rendering pipelines were added; existing rendering pipelines were already verified in phases 11–13. The artifacts modified here (gutter constants, DiffLineStyle return type, dead function removal) do not introduce new dynamic data flows requiring trace verification.

---

### Behavioral Spot-Checks

| Behavior                                      | Command                                                                          | Result                        | Status   |
|-----------------------------------------------|----------------------------------------------------------------------------------|-------------------------------|----------|
| All tests pass after phase changes            | `go test ./...`                                                                  | 219 passed, 0 failed          | ✓ PASS   |
| Build succeeds after dead function removal    | `go build ./...`                                                                 | Exit 0                        | ✓ PASS   |
| `go vet` clean                                | `go vet ./...`                                                                   | No issues                     | ✓ PASS   |
| `DiffLineMutedBackgroundColour` fully absent  | `grep -rn DiffLineMutedBackgroundColour --include=*.go .`                        | 0 matches                     | ✓ PASS   |
| `golangci.yml` has v2 key                     | `grep -c 'version: "2"' .golangci.yml`                                           | 1 match                       | ✓ PASS   |
| `justfile` has install recipe                 | `grep -c 'install:' justfile`                                                    | 1 match                       | ✓ PASS   |
| `.gitignore` root-anchored                    | `grep -c '^/drift$' .gitignore`                                                  | 1 match                       | ✓ PASS   |
| `gutterColumnSeparator` constant present       | `grep -c 'gutterColumnSeparator = " │"' internal/render/gutter.go`              | 1 match                       | ✓ PASS   |

---

### Requirements Coverage

| Requirement | Source Plan   | Description                                                                 | Status            | Evidence                                                                                                    |
|-------------|---------------|-----------------------------------------------------------------------------|-------------------|-------------------------------------------------------------------------------------------------------------|
| CRUFT-01    | 14-01-PLAN.md | Commit 6 pending working-tree changes                                       | ✓ SATISFIED       | Commit `3c27d73` contains all 6 files; git log confirmed                                                   |
| CRUFT-02    | 14-02-PLAN.md | Remove dead `DiffLineMutedBackgroundColour` export from `diffcolors.go`     | ✓ SATISFIED       | Function absent from `diffcolors.go`; full-repo grep clean; tests pass                                     |
| **ORPHANED** | —            | Neither CRUFT-01 nor CRUFT-02 appear in `.planning/REQUIREMENTS.md`         | ✗ ORPHANED        | ROADMAP.md lists both IDs under Phase 14; REQUIREMENTS.md has no CRUFT section and no traceability entries |

---

### Anti-Patterns Found

| File                                      | Pattern                | Severity | Impact                           |
|-------------------------------------------|------------------------|----------|----------------------------------|
| `.planning/REQUIREMENTS.md`               | Missing CRUFT-01/02 requirement definitions and traceability entries | ⚠️ Warning | Planning artifact gap — no blocking code issue, but requirements traceability is broken for this phase |

No code anti-patterns found in any of the 6 modified source files. No TODOs, FIXMEs, placeholders, or hollow stubs detected in the modified files.

---

### Human Verification Required

#### 1. golangci-lint v2 Runtime Validation

**Test:** With `golangci-lint` v2 installed, run `golangci-lint run ./...` from the repo root.
**Expected:** Exit 0, no lint errors reported. The `version: "2"` key selects the v2 schema which restructures `exclusions` blocks under `linters:` and `formatters:` sections separately.
**Why human:** The `golangci-lint` binary was not available in the verification environment. The config file structure was verified structurally (correct YAML keys, correct nesting), but the actual linter engine's acceptance of the schema requires a runtime check.

---

### Gaps Summary

**1 gap blocking full verification:** CRUFT-01 and CRUFT-02 are referenced by two PLANs, both SUMMARYs, and ROADMAP.md, but are entirely absent from `.planning/REQUIREMENTS.md`. This is a documentation/traceability gap — the actual work was completed correctly (commits exist, dead code removed, tests pass). However, the requirements file is the source of truth for what the project commits to deliver, and these two IDs have no definition there.

**Recommended fix:** Add a `### Maintenance` section to REQUIREMENTS.md with:
- `[x] **CRUFT-01**: Commit 6 accumulated working-tree changes from phases 11–13`
- `[x] **CRUFT-02**: Remove dead exported function DiffLineMutedBackgroundColour from internal/highlight/diffcolors.go`

And add traceability rows for both in the Traceability table mapping to Phase 14.

This is a planning artifact fix, not a code fix. All code goals were fully achieved.

---

*Verified: 2026-03-27T15:10:00Z*
*Verifier: the agent (gsd-verifier)*
