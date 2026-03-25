---
phase: 7
slug: support-diffs-from-git-that-is-if-a-single-file-is-provided-and-the-file-is-in-a-git-repo-drift-will-show-the-current-changes
status: passed
verified_at: 2026-03-25
verifier: cursor-agent
---

# Phase 7 — Verification Report

## Overall Status: PASSED

Both plans (07-01, 07-02) have SUMMARY files. Single-path git mode is implemented in `cmd/drift` with fake-`git` tests; two-path and `--from`/`--to` behavior is covered by existing and new tests.

---

## Build & Test

| Command | Exit Code | Result |
|---------|-----------|--------|
| `go test ./... -count=1` | 0 | PASS |
| `go vet ./...` | 0 | PASS |
| `go test ./cmd/drift/... -run TestResolve -count=1` | 0 | PASS |
| `go run ./cmd/drift --help \| grep git` | 0 | Long text mentions git repository |

---

## Regression (prior phases)

`go test ./...` runs all library and CLI packages including testscript under `cmd/drift`. No regressions after Phase 7 changes.

---

## Success criteria (ROADMAP intent)

1. **Git worktree helper** — `resolveGitWorkingTreeVsHEAD` uses `rev-parse` + `show HEAD:<relpath>` with `GIT_TERMINAL_PROMPT=0`; missing HEAD blob → empty OLD.
2. **CLI** — One positional file inside a worktree diffs working tree vs HEAD; outside repo → error containing `git` and `two`; `Args` allows 0 args only with `--from`/`--to`, else 1 or 2 positionals.
3. **Docs** — README CLI example with `working tree` / `HEAD`; `drift --help` describes single-path git mode.

---

## Must-haves (by plan)

### 07-01

- [x] `gitworktree.go` + `gitworktree_test.go` (fake git on PATH)
- [x] `GIT_TERMINAL_PROMPT=0`, `HEAD:`, `filepath.ToSlash`

### 07-02

- [x] `resolveInputs` one-arg branch; `validateRootArgs` on `rootCmd`
- [x] `input_test.go` + `main_test.go` (`TestRunCLI_gitSingleArg_differs`)
- [x] README example line

---

## Human verification

None required for this phase (automated coverage sufficient).
