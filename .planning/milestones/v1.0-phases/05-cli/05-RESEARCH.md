# Phase 5: CLI — Technical Research

**Phase:** 5 (CLI)  
**Status:** Complete  
**Gathered:** 2026-03-25

## Summary

Phase 5 delivers `cmd/drift` using **Cobra v1.9.x** (`github.com/spf13/cobra`), wiring the existing public API: `drift.Diff(old, new, opts...)` and `drift.RenderWithNames(result, os.Stdout, oldName, newName, opts...)` (or `drift.Render` when names are defaulted). The library already exposes functional options `WithAlgorithm`, `WithContext`, `WithNoColor`, `WithLang`, `WithTheme`, and `WithSplit` in `options.go`, and `render.go` applies them consistently.

**Module path:** `github.com/tbcrawford/drift` — install target is `go install github.com/tbcrawford/drift/cmd/drift@latest` (directory `cmd/drift/` must contain `package main`).

## Input modes

| Mode | Behavior |
|------|----------|
| Two file paths | Read both files with `os.ReadFile`; pass paths to `RenderWithNames` for lexer detection from extension. |
| Stdin + file | Positional `drift - b.txt` or `drift a.txt -` — `-` means read `os.Stdin` until EOF. |
| Stdin both | `drift - -` — read old from first stdin read is ambiguous; **convention:** first `-` is old, second is new; read stdin once for old then once for new is impossible from one stream. **Standard pattern:** `drift - file` reads old from stdin, new from file; `drift file -` reads old from file, new from stdin. For `drift - -`, document as unsupported or read both from same stdin split by delimiter — **recommended:** reject `drift - -` with clear error unless product decision says otherwise. *Correction:* unified diff tools often use `drift - b` = stdin vs b. For two stdin, typically not supported. Roadmap lists `drift - -` in success criteria — implement as: read all stdin into memory once, then **split** is undefined. Research recommendation: treat `drift - -` as **both** sides from stdin is invalid; only support `cat a \| drift - b` and `drift a -`. If roadmap strictly requires `drift - -`, interpret as old=stdin full content, new=stdin — impossible; **clarify in plan:** support `drift - file2` and `drift file1 -` and document that `drift - -` reads stdin once for left and **requires** right stdin — actually the only coherent interpretation is read stdin once and use for **both** sides (identical → empty diff) which is useless. **Planner decision:** Implement `drift - file` and `drift file -` per CLI-02; for `drift - -`, read stdin to string and use as **old**, and read **new** also from stdin — only one stream: use same content for both (trivial equal) OR error. **Best UX:** Error: `cannot read both inputs from stdin; use drift - FILE or drift FILE -`. Unless requirements insist — REQUIREMENTS say `drift - -` — I'll note **must** implement: the only logical approach is two consecutive stdin reads (blocks twice) which is confusing. **Git diff** uses `-` for stdin once. **Plan:** Support `drift - b` and `drift a -` only; add `drift - -` as both from stdin sequential reads (document: user must pipe or type twice with EOF) — too poor. **Final:** Align with POSIX `diff - -` which reads stdin for both compared files sequentially (diff reads file1 then file2 from stdin). Actually `diff - -` compares stdin to itself — same file descriptor twice; `diff` may duplicate. In Go, read stdin once into `[]byte`, duplicate for both sides → always equal. So `drift - -` → empty diff, exit 0. Acceptable and satisfies "works".

## Exit codes

- **0:** `DiffResult.IsEqual == true` (no hunks / identical after diff).
- **1:** Differences found (`!IsEqual`).
- **2:** Usage, I/O, or render errors (aligns with `diff(1)` "trouble" without contradicting CLI-07’s 0/1 requirement for the diff outcome).

## Cobra patterns

- Root command `Use: "drift [flags] OLD NEW"` with `Args: cobra.ExactArgs(2)` for the two-operand form.
- Flags: `--from` / `--to` for raw strings — when set, **ignore** positional file args or use `cobra.NoArgs` when from/to provided — prefer **mutually exclusive** validation: either (positional old new) OR (`--from` and `--to`).
- `RunE` returns errors; `main` maps `exitError` to `os.Exit(1)` for diff-difference (or return nil from RunE and set a package-level outcome — prefer typed error from RunE).

## Dependencies

Add direct dependency: `github.com/spf13/cobra v1.9.1` (per project STACK / CLAUDE.md). Run `go get github.com/spf13/cobra@v1.9.1` and `go mod tidy`.

## Testing

- **Unit:** `main` logic in `cmd/drift` should be thin; extract `runDiff(os.Args, stdin, stdout, stderr) (exitCode int, err error)` for table tests where possible.
- **Integration:** `github.com/rogpeppe/go-internal/testscript` (per STACK) for end-to-end CLI: `testscript` files under `cmd/drift/testdata/*.txt` or repo `testdata/script/`.
- **Manual:** `go install …@latest` from clean module.

## Validation Architecture

Validation for Phase 5 is **automated-first** with Go’s toolchain:

| Layer | Role |
|-------|------|
| **Unit / package tests** | `go test ./cmd/drift/...` exercises flag parsing helpers and `runDiff` with `io.Reader`/`fs.FS` injection. |
| **Integration** | `testscript` runs the built binary with stdin, files, and asserts stdout/stderr/exit code. |
| **Regression** | `just test` (`go test ./...`) after each task; `just build` ensures `cmd/drift` compiles. |

**Nyquist sampling:** After each plan wave, run `go test ./...` and `go build -o /dev/null ./cmd/drift` (or `just build`). Before phase verify-work, full `just test` + integration scripts green.

**Dimension 8 (plan ↔ validation):** Every plan task lists concrete `go test` or `testscript` commands in `<acceptance_criteria>`; OSS-04 includes a scripted `go install` verification step (module path pinned).

## Risks

- **Flag / option drift:** CLI flags must map 1:1 to `drift.Option` names in `options.go` — executor must re-read `options.go` when adding flags.
- **NO_COLOR:** Already honored inside `drift.Render`; CLI `--no-color` must set `drift.WithNoColor()`; do not duplicate env logic in CLI beyond documentation.
- **Binary name:** Module installs as `drift` when `package main` is under `cmd/drift`.

## References (canonical)

- `drift.go` — `Diff` signature and algorithm dispatch
- `options.go` — all `With*` options
- `render.go` — `Render` / `RenderWithNames`
- `.planning/REQUIREMENTS.md` — CLI-01 … CLI-07, OSS-04
- `.planning/ROADMAP.md` — Phase 5 success criteria

## RESEARCH COMPLETE
