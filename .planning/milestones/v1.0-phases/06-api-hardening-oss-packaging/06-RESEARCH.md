# Phase 6: API Hardening & OSS Packaging — Technical Research

**Gathered:** 2026-03-25  
**Status:** Ready for planning  
**Context note:** No `06-CONTEXT.md` present; research derived from ROADMAP, REQUIREMENTS, STATE, and current `github.com/tylercrawford/drift` source.

---

## 1. Current public API (baseline)

| Surface | Location | Notes |
|---------|----------|-------|
| `Diff(old, new string, opts ...Option) (DiffResult, error)` | `drift.go` | Functional API; applies `config` via options |
| `Render`, `RenderWithNames` | `render.go` | Same `Option` set as `Diff` for theme/lang/split/no-color |
| `Option` + `With*` funcs | `options.go` | `WithAlgorithm`, `WithContext`, `WithNoColor`, `WithLang`, `WithTheme`, `WithSplit` |
| Exported types | `types.go`, `options.go` | `Algorithm`, `Op`, `Edit`, `Span`, `Line`, `Hunk`, `DiffResult` (aliases) |
| Package doc | `doc.go` | References `examples/` but examples do not exist yet |

**Implication:** Builder API is additive. Recommended pattern: unexported slice of `Option` on a small exported type (e.g. `Differ` or `Builder`) with methods that append options and terminal methods `Diff(string, string) (DiffResult, error)` and `Render(DiffResult, io.Writer) error` delegating to package-level `Diff` / `Render`. Avoid duplicating `defaultConfig` logic.

---

## 2. Builder / fluent API (CORE-05)

**Industry pattern (Go):** Method chaining returns the same pointer receiver; each method appends or wraps `Option` values; terminal methods expand `opts...` into existing package functions.

**Concrete API target (from ROADMAP success criteria):**

```go
result, err := drift.New().Algorithm(drift.Histogram).WithTheme("github").Diff(a, b)
```

**Naming:** `New()` is idiomatic. Method names should mirror functional options where possible: `Algorithm(Algorithm)`, `Context(int)`, `NoColor()`, `Lang(string)`, `Theme(string)`, `Split()` — each delegates to `With*` internally to avoid drift between builder and functional APIs.

**Testing:** Table-driven test proving builder output matches `Diff` / `Render` with equivalent `With*` options (same `DiffResult` or same rendered bytes with `WithNoColor()`).

**Stability:** v1.0.0 — builder is a second entry point only; do not remove or change existing `Diff`/`Render` signatures.

---

## 3. Godoc completeness (OSS-02)

**Tooling:**

- `go doc github.com/tylercrawford/drift` — package summary
- `go doc -all github.com/tylercrawford/drift` — every exported symbol
- `staticcheck` / `golangci-lint` may include comment checks if enabled

**Convention:** Every exported name needs a comment starting with the name (`// Diff computes...`). Constants in a block can share a lead comment plus per-value comments where non-obvious (`Myers`, `Patience`, `Histogram` already partially documented).

**Audit scope:** Root package only for OSS-02 as stated in requirements; internal packages may stay minimally documented unless exported test helpers exist.

---

## 4. Examples (OSS-03)

**Layout (REQUIREMENTS):**

- `examples/basic/main.go` — functional API: `Diff` + `Render` to stdout
- `examples/builder/main.go` — builder: `New()` chain + `Diff` + `Render`

**Constraints:** Each must be `go run`-able from repo root; use `WithNoColor()` or document `NO_COLOR` so CI/log output is stable. Prefer small hardcoded strings that visibly differ (add/remove line).

**go.mod:** Standard library for examples is enough if they only import `drift`; no separate module required for v1 unless examples need third-party deps (avoid for simplicity).

---

## 5. Benchmarks (OSS-07)

**Target:** 10,000 lines per side, diff + render unified and split; ROADMAP requires &lt; 1s per bench iteration (interpret as `ns/op` or wall time — use `b.N` with sub-benchmarks; document threshold as `go test -bench=. -benchtime=1x` or assert order-of-magnitude in plan acceptance).

**Implementation notes:**

- Generate reproducible `old`/`new` strings in `testing.B` setup (e.g. repeated lines with a known edit block) — avoid I/O in the loop
- Use `bytes.Buffer` or `io.Discard` as `io.Writer` for `Render`
- Use `WithNoColor()` to avoid terminal detection overhead in benchmarks
- Split benchmark must set `WithSplit()` and provide reasonable `TermWidth` behavior — `render.TerminalWidth` may return default when writer is not TTY; verify split path is exercised (read `internal/render/termwidth.go`)

**File placement:** `*_test.go` next to `drift.go` / `render.go` or dedicated `bench_test.go` in root package.

---

## 6. README (OSS-06)

**Sections (minimum):**

1. **Installation** — `go get`, `go install .../cmd/drift@latest`
2. **CLI** — copy-paste from `drift --help` or representative invocations
3. **Library (functional)** — `Diff` + `Render` snippet
4. **Library (builder)** — chain snippet matching ROADMAP
5. **Rendering** — unified vs split, `WithTheme` / `WithLang` mention

**Assets:** ROADMAP allows snippets if screenshots unavailable; optional terminal PNGs deferred unless already present.

---

## 7. Risks and mitigations

| Risk | Mitigation |
|------|------------|
| Builder API diverges from `Option` semantics | Implement builder only by appending `With*` options |
| Godoc churn | Single audit pass with checklist tied to `go doc -all` |
| Benchmark flakiness on CI | Use large enough `benchtime`; document local threshold |
| Examples break module tidy | Run `go run ./examples/...` in acceptance |

---

## Validation Architecture

**Dimension 8 — Automated verification for Phase 6**

| Layer | Command | When |
|-------|---------|------|
| Unit / compile | `go test ./...` | After every task |
| Builder parity | `go test ./... -run Builder` (or named test) | After builder implementation |
| Examples | `go run ./examples/basic/` and `go run ./examples/builder/` | After examples plan |
| Benchmarks | `go test -bench=. -benchmem ./...` (subset flags) | After bench plan |
| Godoc spot-check | `go doc -all github.com/tylercrawford/drift \| head` | After godoc plan |

**Nyquist sampling:** Run `go test ./...` after each committed task; run full bench suite once per wave before merge.

---

## RESEARCH COMPLETE

Phase 6 planning can proceed with five plans aligned to ROADMAP items 06-01 … 06-05.
