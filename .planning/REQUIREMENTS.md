# Requirements: drift

**Defined:** 2026-03-25
**Core Value:** A Go developer can `import "github.com/tylercrawford/drift"` and get a production-quality, richly-rendered diff with one function call — the same quality they see in GitHub's PR review UI but in the terminal.

## v1 Requirements

### Core Engine

- [x] **CORE-01**: Library correctly diffs any two multi-line strings using the Myers algorithm
- [x] **CORE-02**: Hunk builder merges adjacent edits into context windows (default 3 lines, configurable)
- [x] **CORE-03**: Library exports structured data model: `Op`, `Edit`, `Hunk`, `Line`, `DiffResult` types
- [x] **CORE-04**: Library exposes functional API: `drift.Diff(a, b string, opts ...Option) DiffResult`
- [ ] **CORE-05**: Library exposes builder/fluent API: `drift.New().Algorithm(drift.Myers).Diff(a, b)`
- [x] **CORE-06**: `Diff()` returns an empty result immediately when both inputs are identical (no allocation)
- [x] **CORE-07**: Library normalizes line endings on input (`\r\n` → `\n`); handles Windows files correctly

### Algorithms

- [ ] **ALGO-01**: Library implements Patience diff algorithm in `internal/algo/patience/`
- [ ] **ALGO-02**: Library implements Histogram diff algorithm in `internal/algo/histogram/`
- [ ] **ALGO-03**: Caller selects algorithm via `drift.WithAlgorithm(drift.Myers|drift.Patience|drift.Histogram)` option
- [ ] **ALGO-04**: Both Patience and Histogram implementations fall back to Myers for edge cases (high-repetition inputs, inter-anchor gaps)

### Rendering

- [ ] **REND-01**: Library renders unified diff output with Git-style `@@ -a,b +c,d @@` hunk headers and `+`/`-` line prefixes
- [ ] **REND-02**: Library renders side-by-side split diff with left/right panels via Lip Gloss, terminal-width-aware
- [ ] **REND-03**: Library applies Chroma v2 syntax highlighting per line, with diff colors (red/green) layered on top
- [ ] **REND-04**: Library auto-detects terminal light/dark theme and selects the best matching Chroma theme
- [ ] **REND-05**: Caller can override the Chroma theme via `drift.WithTheme("monokai")` option
- [ ] **REND-06**: Library auto-detects language from file extension; falls back to content analysis via `lexers.Analyse()`
- [ ] **REND-07**: Caller can override detected language via `drift.WithLang("go")` option
- [ ] **REND-08**: Library detects terminal color depth (TrueColor / 256-color / 16-color / none) and degrades gracefully
- [ ] **REND-09**: Library and CLI fully disable colors when `NO_COLOR` env var is set or `drift.WithNoColor()` option is passed

### CLI

- [ ] **CLI-01**: CLI accepts two file path arguments: `drift file1 file2`
- [ ] **CLI-02**: CLI supports stdin piping: `cat a.txt | drift - b.txt` or `drift - -`
- [ ] **CLI-03**: CLI supports raw string arguments: `drift --from 'text a' --to 'text b'`
- [ ] **CLI-04**: CLI supports `--split` flag to switch to side-by-side output (default is unified)
- [ ] **CLI-05**: CLI supports `--algorithm myers|patience|histogram` flag
- [ ] **CLI-06**: CLI supports `--lang`, `--theme`, `--no-color`, and `--context N` flags
- [ ] **CLI-07**: CLI exits with code `1` when inputs differ, `0` when identical (matches POSIX `diff` semantics)

### OSS Packaging

- [x] **OSS-01**: Project has a valid `go.mod` with module path `github.com/tylercrawford/drift`
- [ ] **OSS-02**: All exported types, functions, and options have godoc comments
- [ ] **OSS-03**: `examples/` directory contains runnable `basic/` and `builder/` examples
- [ ] **OSS-04**: `go install github.com/tylercrawford/drift/cmd/drift@latest` installs the CLI correctly
- [x] **OSS-05**: Project includes an MIT `LICENSE` file
- [ ] **OSS-06**: `README.md` covers installation, CLI usage, library API, and rendering examples
- [ ] **OSS-07**: Benchmarks exist for diffing a 10,000-line file (unified and split views)
- [x] **OSS-08**: Property-based tests verify algorithm correctness invariants (round-trip: `apply(diff(a,b), a) == b`)
- [x] **OSS-09**: Project includes a `justfile` (or `Makefile`) for common repository maintenance tasks (test, lint, build, benchmark)

## v2 Requirements

### Rendering

- **REND-V2-01**: Intra-line word-level diff highlighting (character-level changes within a modified line)
- **REND-V2-02**: HTML render output (for web embedding)

### CLI

- **CLI-V2-01**: Binary file detection — print "Binary files X and Y differ" instead of diff output

### Developer Experience

- **DX-V2-01**: GitHub Actions CI workflow (lint, test, benchmark)
- **DX-V2-02**: Changelog (`CHANGELOG.md`) and `CONTRIBUTING.md`
- **DX-V2-03**: Interactive scrollable TUI mode (Bubble Tea, vimdiff-style pager)

## Out of Scope

| Feature | Reason |
|---------|--------|
| Real-time / file-watching diff | Not a diff library concern; callers can poll and call Diff() themselves |
| Interactive TUI (vimdiff-style) | High complexity; static output covers the primary use case for v1 |
| HTML render output | Terminal-first; HTML is additive and deferred to v2 |
| Intra-line word-level diff | `Line` type must stabilize at v1.0 to avoid breaking changes; word diff ships as v1.x additive |
| GitHub Actions CI | Not required for a functional v1; can be added without breaking the library |

## Traceability

Which phases cover which requirements. Updated during roadmap creation.

| Requirement | Phase | Status |
|-------------|-------|--------|
| CORE-01 | Phase 1 — Foundation | Complete |
| CORE-02 | Phase 1 — Foundation | Complete |
| CORE-03 | Phase 1 — Foundation | Complete |
| CORE-04 | Phase 1 — Foundation | Complete |
| CORE-06 | Phase 1 — Foundation | Complete |
| CORE-07 | Phase 1 — Foundation | Complete |
| OSS-01 | Phase 1 — Foundation | Complete |
| OSS-05 | Phase 1 — Foundation | Complete |
| OSS-08 | Phase 1 — Foundation | Complete |
| OSS-09 | Phase 1 — Foundation | Complete |
| ALGO-01 | Phase 2 — Algorithms | Pending |
| ALGO-02 | Phase 2 — Algorithms | Pending |
| ALGO-03 | Phase 2 — Algorithms | Pending |
| ALGO-04 | Phase 2 — Algorithms | Pending |
| REND-01 | Phase 3 — Unified Rendering | Pending |
| REND-03 | Phase 3 — Unified Rendering | Pending |
| REND-04 | Phase 3 — Unified Rendering | Pending |
| REND-05 | Phase 3 — Unified Rendering | Pending |
| REND-06 | Phase 3 — Unified Rendering | Pending |
| REND-07 | Phase 3 — Unified Rendering | Pending |
| REND-08 | Phase 3 — Unified Rendering | Pending |
| REND-09 | Phase 3 — Unified Rendering | Pending |
| REND-02 | Phase 4 — Split Rendering | Pending |
| CLI-01 | Phase 5 — CLI | Pending |
| CLI-02 | Phase 5 — CLI | Pending |
| CLI-03 | Phase 5 — CLI | Pending |
| CLI-04 | Phase 5 — CLI | Pending |
| CLI-05 | Phase 5 — CLI | Pending |
| CLI-06 | Phase 5 — CLI | Pending |
| CLI-07 | Phase 5 — CLI | Pending |
| OSS-04 | Phase 5 — CLI | Pending |
| CORE-05 | Phase 6 — API Hardening & OSS Packaging | Pending |
| OSS-02 | Phase 6 — API Hardening & OSS Packaging | Pending |
| OSS-03 | Phase 6 — API Hardening & OSS Packaging | Pending |
| OSS-06 | Phase 6 — API Hardening & OSS Packaging | Pending |
| OSS-07 | Phase 6 — API Hardening & OSS Packaging | Pending |

**Coverage:**
- v1 requirements: 35 total
- Mapped to phases: 35
- Unmapped: 0 ✓

---
*Requirements defined: 2026-03-25*
*Last updated: 2026-03-25 after initial definition*
