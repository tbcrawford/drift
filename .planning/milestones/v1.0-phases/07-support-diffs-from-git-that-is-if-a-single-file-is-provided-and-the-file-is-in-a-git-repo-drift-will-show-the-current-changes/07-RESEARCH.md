# Phase 7 — Research: Git “single file = working tree changes”

**Phase:** 7 — support diffs from git (single path in a repo → show current changes)  
**Gathered:** 2026-03-25  
**Question answered:** What do we need to know to plan CLI + resolver changes safely?

---

## 1. Product semantics (“current changes”)

**Roadmap intent:** One positional path, file lives in a Git worktree → user sees how that file differs from version control without passing a second path.

**Baseline choice (recommended):** Compare **working tree file contents** to **`HEAD` blob** for the same path (repository-relative). This matches the common mental model “what changed in this file since the last commit,” including staged+unstaged net change vs `HEAD`, consistent with `git diff HEAD -- <path>` *conceptually* (we do not parse `git diff` output—we reconstruct two full strings and feed `drift.Diff`).

| Scenario | OLD (left) | NEW (right) | Notes |
|----------|------------|-------------|--------|
| Tracked file, modified | `git show HEAD:<relpath>` | `os.ReadFile(abs)` | Core case |
| New file (untracked) | empty string | `os.ReadFile(abs)` | Same as diff vs `/dev/null`; hunks show all lines added |
| Tracked, deleted on disk | `git show HEAD:<relpath>` | empty string | All lines removed |
| File outside repo / not inside worktree | — | — | **Not** git-single mode; error: need two paths or a path inside a Git worktree |
| Repo exists but `git show` fails (e.g. path not in `HEAD`) | Treat as empty OLD if Git reports “path not in tree” | worktree | Align with new-file-at-that-path-in-index cases—planner should pick one consistent rule and test it |

**Explicit non-goals for v1 of this feature:**

- Do not shell out to `git diff` and parse unified patches (duplicates logic, breaks highlighting ownership).
- Do not add `libgit2` / `go-git` unless later phases require it; **`os/exec` + porcelain plumbing** is enough and matches “stdlib-first” stack.

---

## 2. Git discovery and paths (Go + `git` CLI)

**Detect repo:** From the **absolute** path of the user’s file, run Git with appropriate working directory:

- `git -C <parentDir> rev-parse --is-inside-work-tree` → expect `true`
- `git -C <parentDir> rev-parse --show-toplevel` → absolute repo root

**Repository-relative path:** `rel, err := filepath.Rel(topLevel, absFile)` then `filepath.ToSlash(rel)` for `git show HEAD:path` (Git expects forward slashes in object paths even on Windows).

**Resolve `absFile`:** `filepath.Clean` + `filepath.Abs` so `.`, `..`, and symlinks behave predictably; document that behavior matches `os.ReadFile`.

**Submodules / worktrees:** If `rev-parse` succeeds, same logic applies; edge cases (file is gitlink) can return a clear error (“cannot diff submodule path”)—planner may defer to “best effort + readable error.”

---

## 3. Commands to implement (exact `git` argv)

Use **non-interactive** invocations; set `cmd.Env` append `GIT_TERMINAL_PROMPT=0` to avoid hangs if Git ever prompts.

| Purpose | Suggested argv |
|---------|----------------|
| Inside worktree? | `git -C <dir> rev-parse --is-inside-work-tree` |
| Repo root | `git -C <dir> rev-parse --show-toplevel` |
| HEAD version of file | `git -C <repoRoot> show HEAD:<relpath>` — stdout is raw file bytes |

**Errors:** If `git` is not on `PATH`, wrap error with hint: `git not found in PATH`. If not a worktree, single-arg mode should **fail** with message containing `git` and `two paths` so tests can grep.

---

## 4. CLI / Cobra integration points

**Current state:** `rootCmd` uses `Args: cobra.MaximumNArgs(2)` and `resolveInputs` requires exactly two operands unless `--from`/`--to`.

**Target state:**

- **0 args:** invalid (unless both `--from` and `--to`—unchanged).
- **1 arg:** enter **git-single** resolver when path is a file inside a Git worktree; otherwise error listing valid patterns.
- **2 args:** existing `resolveInputs` behavior (files / `-`).

**Mutual exclusivity:** `--from`/`--to` with any positional arg stays invalid (existing rule).

**Display names for render:** e.g. `filepath.Base(path)` for both sides, or `base` vs `base (HEAD)`—planner picks strings that stay stable in golden tests; must be documented in plan acceptance criteria.

---

## 5. Testing strategy (no network, minimal flake)

- **Unit tests** for path / `ToSlash` / “is inside toplevel” logic with `testing/fstest` **cannot** mock `git`; use **fake `git` binary** in `PATH` (pattern used elsewhere in Go ecosystem): temp dir, write script that echoes expected outputs, prepend to `PATH` in tests.
- **Integration test** optional: skip if `git` missing in CI (use `t.Skip`); local dev runs real Git.
- Preserve existing `input_test.go` coverage for two-arg modes.

---

## 6. Library vs CLI

**Recommendation:** Keep Git orchestration in **`cmd/drift`** (or `cmd/drift/internal/...` package) only—library remains pure string-in/string-out. Phase 7 does **not** add `drift.WithGit()` unless explicitly requested later.

---

## 7. Documentation touchpoints

- `README.md` CLI section: one example `drift ./pkg/foo.go` inside a repo.
- `--help` long description one line for single-path git mode.

---

## RESEARCH COMPLETE

Findings are sufficient to plan resolver changes, Cobra arg validation, `git` invocation wrapper, and test layout without new third-party dependencies.

---

## Validation Architecture

**Nyquist / execution validation**

| Dimension | Approach |
|-----------|----------|
| **Automated unit** | `go test ./cmd/drift/...` — table tests for single-arg resolution with fake `git` on `PATH`; grep-verifiable error substrings |
| **Automated integration** | Optional `go test` with `-run Integration` and `git` present; or `testscript` if repo adopts it for CLI |
| **Manual smoke** | In any Git repo: edit file, run `drift path/to/file`; compare mentally to `git diff HEAD -- path` |
| **Regression guard** | All Phase 5 two-path and `--from`/`--to` tests remain green |

**Quick command:** `go test ./cmd/drift/...`  
**Full command:** `just test` (or `go test ./...` if that is repo standard)

**Sampling:** Run `go test ./cmd/drift/...` after each plan’s tasks touching `cmd/drift`; full `./...` before phase verify.
