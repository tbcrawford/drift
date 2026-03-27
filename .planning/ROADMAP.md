# Roadmap: drift

## Overview

`drift` is built in six focused phases, each delivering a verifiable capability that the next phase depends on. The ordering is dictated by the dependency graph: Myers algorithm must exist before Patience/Histogram can fall back to it; the stable `DiffResult` type must exist before any renderer consumes it; unified rendering must exist before split view is additive on top; and the CLI is a thin last-mile consumer assembled from the fully-formed library. The final phase hardens the public API surface and packages the library for open-source distribution as `v1.0.0`.

## Phases

**Phase Numbering:**
- Integer phases (1, 2, 3): Planned milestone work
- Decimal phases (2.1, 2.2): Urgent insertions (marked with INSERTED)

Decimal phases appear between their surrounding integers in numeric order.

- [x] **Phase 1: Foundation** - Module scaffold, exported data model, Myers algorithm, functional API, and OSS baseline (completed 2026-03-25)
- [x] **Phase 2: Algorithms** - Patience and Histogram diff algorithms with Myers fallback and algorithm selection option (completed 2026-03-25)
- [x] **Phase 3: Unified Rendering** - Chroma syntax highlighting, terminal theme detection, language detection, and unified diff output (completed 2026-03-25)
- [x] **Phase 4: Split Rendering** - Side-by-side split diff view via Lip Gloss two-panel layout (completed 2026-03-25)
- [x] **Phase 5: CLI** - Cobra command wrapping the library for file, stdin, and raw-string input with all flags (completed 2026-03-25)
- [x] **Phase 6: API Hardening & OSS Packaging** - Builder API, godoc, examples, benchmarks, and v1.0.0 stabilization (completed 2026-03-25)
- [x] **Phase 7: support diffs from git** - single-file path in a repo shows working tree diff (completed 2026-03-26)
- [x] **Phase 8: Terminal palette best-match Chroma theme** - OSC 4 palette + BestMatchTheme (completed 2026-03-26)
- [x] **Phase 9: Dual line-number gutters** - old/new gutters, unified and split (completed 2026-03-26)
- [x] **Phase 10: Theme-aware full-line diff styling** - depends on Phase 9 (completed 2026-03-26)
- [x] **Phase 11: GitHub PR-style intra-line highlights** - depends on Phase 10 (completed 2026-03-26)

## Phase Details

### Phase 1: Foundation
**Goal**: The core diff engine is working — a caller can diff two strings using Myers algorithm and receive a structured `DiffResult` via the functional API
**Depends on**: Nothing (first phase)
**Requirements**: CORE-01, CORE-02, CORE-03, CORE-04, CORE-06, CORE-07, OSS-01, OSS-05, OSS-08, OSS-09
**Success Criteria** (what must be TRUE):
  1. `drift.Diff(a, b)` returns a `DiffResult` with correct `[]Hunk` for any two multi-line string inputs
  2. Identical inputs return an empty `DiffResult` immediately (verifiable via benchmark: zero allocations)
  3. Files with Windows `\r\n` line endings produce the same diff output as Unix `\n` files
  4. Property-based tests pass: `apply(diff(a, b), a) == b` holds for all generated inputs
  5. `go.mod` exists at module root with path `github.com/tylercrawford/drift`, MIT LICENSE file present, and `just test` / `just build` / `just lint` run successfully
**Plans**: 5 plans

Plans:
- [x] 01-01-PLAN.md — Module scaffold: go.mod (github.com/tylercrawford/drift), MIT LICENSE, justfile, golangci-lint config
- [x] 01-02-PLAN.md — Exported data model: Op, Edit, Hunk, Line, DiffResult, Option/config, internal algo.Differ interface
- [x] 01-03-PLAN.md — Myers algorithm (TDD): implement internal/algo/myers/ with table-driven + cross-validated tests
- [x] 01-04-PLAN.md — Hunk builder + drift.Diff() API: internal/hunk/, drift.go, \r\n normalization
- [x] 01-05-PLAN.md — Property-based + fuzz testing: rapid round-trip invariant, Go fuzz for Myers

### Phase 2: Algorithms
**Goal**: Callers can select Patience or Histogram algorithms, with correct Myers fallback for edge cases
**Depends on**: Phase 1
**Requirements**: ALGO-01, ALGO-02, ALGO-03, ALGO-04
**Success Criteria** (what must be TRUE):
  1. `drift.WithAlgorithm(drift.Patience)` produces a diff that correctly identifies moved blocks a pure Myers diff misses on representative test inputs
  2. `drift.WithAlgorithm(drift.Histogram)` produces a diff with correct hunk boundaries on repetitive-line inputs
  3. Patience and Histogram both automatically fall back to Myers on high-repetition inputs without panicking or producing incorrect output
  4. Algorithm selection can be round-tripped through property-based tests: all three algorithms satisfy `apply(diff(a, b), a) == b`
**Plans**: TBD

Plans:
- [x] 02-01: Research Histogram frequency cutoff and fallback trigger (raygard source); document decision
- [x] 02-02: Implement Histogram diff in `internal/algo/histogram/` with Myers fallback for high-repetition inputs
- [x] 02-03: Implement Patience diff in `internal/algo/patience/` with inter-anchor Myers fallback
- [x] 02-04: Wire `WithAlgorithm()` option and integrate all three algorithms into `drift.Diff()`

### Phase 3: Unified Rendering
**Goal**: `drift.Diff()` produces richly highlighted unified diff output renderable to any `io.Writer`
**Depends on**: Phase 2
**Requirements**: REND-01, REND-03, REND-04, REND-05, REND-06, REND-07, REND-08, REND-09
**Success Criteria** (what must be TRUE):
  1. Unified diff output contains correct `@@ -a,b +c,d @@` hunk headers with `+`/`-` prefixed lines matching Git's format
  2. Syntax highlighting is applied per-line using Chroma v2; Go source tokens visually distinguish keywords, strings, and identifiers in terminal output
  3. Terminal dark/light theme is auto-detected; running in a dark terminal selects a dark Chroma theme and a light terminal selects a light theme without any flags
  4. `drift.WithTheme("monokai")` overrides auto-detection; `drift.WithLang("go")` overrides language detection from file extension
  5. `NO_COLOR` env var or `drift.WithNoColor()` completely strips all ANSI codes from output; 16-color and 256-color terminals receive appropriately degraded output
**Plans**: TBD

Plans:
- [x] 03-01: Research Lip Gloss v2 `HasDarkBackground` timeout behavior; implement `internal/theme/` with safe OSC 11 detection
- [x] 03-02: Implement `internal/highlight/` — Chroma v2 lexer → formatter pipeline with color profile detection
- [x] 03-03: Implement language auto-detection (`lexers.Match()` + `lexers.Analyse()` fallback) and `WithLang()` / `WithTheme()` options
- [x] 03-04: Implement `internal/render/unified.go` — UnifiedRenderer with hunk headers and `+`/`-` prefixes
- [x] 03-05: Wire `WithNoColor()` and terminal color depth detection (TrueColor / 256 / 16 / none)

### Phase 4: Split Rendering
**Goal**: Callers can request side-by-side split diff output with correct two-panel layout at any terminal width
**Depends on**: Phase 3
**Requirements**: REND-02
**Success Criteria** (what must be TRUE):
  1. Split diff output shows left (old) and right (new) panels side-by-side, each occupying half the terminal width, with syntax highlighting preserved in both panels
  2. Output renders correctly at narrow (80 col), standard (120 col), and wide (200 col) terminal widths without panel overflow or misalignment
  3. ANSI escape sequences in highlighted lines do not inflate measured panel width (verified by comparing `lipgloss.Width()` vs `len()` on highlighted output)
  4. Split view works correctly when output is piped (no TTY): falls back to a safe default width rather than panicking
**Plans**: TBD

Plans:
- [x] 04-01: Implement `internal/render/split.go` — Lip Gloss `JoinHorizontal` two-panel layout with ANSI-aware width measurement
- [x] 04-02: Add terminal width detection with pipe fallback; handle Unicode wide characters (`runewidth` `EastAsian = false`)
- [x] 04-03: Wire `WithSplit()` option into `drift.Render()` and validate split vs unified output

### Phase 5: CLI
**Goal**: The `drift` CLI binary is installable and correctly wraps the library for all input modes and flags
**Depends on**: Phase 4
**Requirements**: CLI-01, CLI-02, CLI-03, CLI-04, CLI-05, CLI-06, CLI-07, OSS-04
**Success Criteria** (what must be TRUE):
  1. `drift file1.go file2.go` diffs two files and prints unified output; `drift --split file1.go file2.go` prints split output
  2. `cat a.txt | drift - b.txt` and `drift --from 'text a' --to 'text b'` both produce correct diff output
  3. All flags work: `--algorithm`, `--lang`, `--theme`, `--no-color`, `--context N`, `--split`
  4. CLI exits with code `1` when inputs differ and `0` when identical (verified with `echo $?`)
  5. `go install github.com/tylercrawford/drift/cmd/drift@latest` installs the binary successfully and `drift --help` runs
**Plans**: TBD

Plans:
- [x] 05-01: Implement `cmd/drift/main.go` with Cobra root command and flag definitions
- [x] 05-02: Implement file path, stdin pipe, and `--from`/`--to` raw string input handling
- [x] 05-03: Wire all flags through to library options; implement exit code logic
- [x] 05-04: Verify `go install github.com/tylercrawford/drift/cmd/drift@latest` and test all input modes end-to-end

### Phase 6: API Hardening & OSS Packaging
**Goal**: The library is ready for `v1.0.0` — public API is stable, documented, exemplified, and benchmarked
**Depends on**: Phase 5
**Requirements**: CORE-05, OSS-02, OSS-03, OSS-06, OSS-07
**Success Criteria** (what must be TRUE):
  1. `drift.New().Algorithm(drift.Histogram).WithTheme("github").Diff(a, b)` compiles and returns correct output (builder API works)
  2. Every exported type, function, and option has a godoc comment; `go doc github.com/tylercrawford/drift` renders clean documentation with no missing entries
  3. `examples/basic/` and `examples/builder/` directories contain runnable programs; `go run examples/basic/main.go` produces visible diff output
  4. Benchmark for 10,000-line file diff completes in under 1 second for both unified and split renderers (verifiable with `go test -bench=.`)
  5. `README.md` covers installation, CLI usage, library functional API, builder API, and rendering examples with at least one code snippet each
**Plans**: TBD

Plans:
- [x] 06-01: Implement builder/fluent API (`drift.New()` with method chaining delegating to functional options)
- [x] 06-02: Audit all exported symbols; write or improve godoc comments to complete coverage
- [x] 06-03: Write `examples/basic/main.go` and `examples/builder/main.go` runnable examples
- [x] 06-04: Write performance benchmarks for 10,000-line unified and split diffs
- [x] 06-05: Write `README.md` with installation, usage, API reference, and rendering screenshots/snippets

## Progress

**Execution Order:**
Phases execute in numeric order: 1 → 2 → 3 → 4 → 5 → 6

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. Foundation | 5/5 | Complete   | 2026-03-25 |
| 2. Algorithms | 4/4 | Complete    | 2026-03-25 |
| 3. Unified Rendering | 5/5 | Complete    | 2026-03-25 |
| 4. Split Rendering | 3/3 | Complete | 2026-03-25 |
| 5. CLI | 4/4 | Complete | 2026-03-25 |
| 6. API Hardening & OSS Packaging | 5/5 | Complete | 2026-03-25 |
| 7. Git working-tree diff | 2/2 | Complete | 2026-03-26 |
| 8. Terminal palette / BestMatchTheme | 2/2 | Complete | 2026-03-26 |
| 9. Dual line-number gutters | 2/2 | Complete | 2026-03-26 |
| 10. Theme-aware full-line diff styling | 1/1 | Complete | 2026-03-26 |
| 11. Intra-line word highlights | 2/2 | Complete    | 2026-03-26 |

### Phase 7: support diffs from git that is, if a single file is provided and the file is in a git repo drift will show the current changes

**Goal:** [To be planned]
**Requirements**: TBD
**Depends on:** Phase 6
**Plans:** 0 plans

Plans:
- [ ] TBD (run /gsd-plan-phase 7 to break down)

### Phase 8: Terminal palette best-match Chroma theme

**Goal:** When the user does not pass an explicit Chroma theme, drift may query the terminal’s ANSI palette (OSC 4) and pick the registered Chroma style whose sampled syntax-token colors are closest to that palette (Terrasort-style `BestMatchTheme`), then fall back to the existing light/dark `SelectTheme` path when OSC 4 is unavailable, fails, or stdout is not a TTY.

**Depends on:** Phase 3 (rendering) — extends theme resolution; library and CLI remain stdlib-first (no new deps unless already allowed).

**Requirements:** REND-04 (best matching Chroma theme — currently binary dark/light), REND-08

**Plans:** 2 plans

Plans:
- [x] 08-01-PLAN.md — Pure `BestMatchTheme` + `ParseOSC4Responses` in `internal/highlight` (completed 2026-03-26)
- [x] 08-02-PLAN.md — Unix OSC 4 query, `Render` auto-theme wiring, `--show-theme`, README (completed 2026-03-26)

### Phase 9: Dual line-number gutters (old | new), centered numbers with space padding; brighter gutter background vs muted code — unified and split

**Goal:** Shared gutter formatting and unified/split integration with `WithLineNumbers` / `WithoutLineNumbers` (default on).
**Requirements**: TBD
**Depends on:** Phase 8
**Plans:** 2 plans

Plans:
- [x] 09-01-PLAN.md — Gutter primitives + unified (completed 2026-03-26)
- [x] 09-02-PLAN.md — Split gutters + README + builder (completed 2026-03-26)

### Phase 10: Theme-aware full-line diff styling for additions and deletions: entire line highlighted; hybrid Chroma + ANSI/Lip Gloss; green/red from selected theme — unified and split

**Goal:** Theme-derived full-line backgrounds on +/- lines from Chroma `GenericDeleted` / `GenericInserted`, ANSI-safe token boundaries, `WithLineDiffStyle` (default on).
**Requirements**: TBD
**Depends on:** Phase 9
**Plans:** 1 plan

Plans:
- [x] 10-01-PLAN.md — highlight + unified + split + options + tests (completed 2026-03-26)

### Phase 11: GitHub PR-style intra-line change highlights: word-level chunks; changed spans use gutter background; rest of line muted; Chroma tokens on tinted backgrounds — unified and split

**Goal:** Word-level alignment + split/unified segmented rendering (muted vs gutter-tinted changed spans); `WithWordDiff` (default on).
**Requirements**: TBD
**Depends on:** Phase 10
**Plans:** 4/4 plans complete

Plans:
- [x] 11-01-PLAN.md — `internal/worddiff` foundation (completed 2026-03-26)
- [x] 11-02-PLAN.md — render integration + options + tests (completed 2026-03-26)

### Phase 12: Restructure project layout for idiomatic Go library and CLI

**Goal:** Clean up internal project layout for idiomatic Go library distribution: move the exported `testdata.Apply` test helper to `internal/testhelpers`, split the flat `config` struct into named `diffConfig` + `renderConfig` sub-structs for self-documenting separation of concerns, and update `doc.go` + README to reflect the full API surface added in phases 7–11.
**Requirements**: LAYOUT-01, LAYOUT-02, LAYOUT-03
**Depends on:** Phase 11
**Plans:** 2/2 plans complete

Plans:
- [x] 12-01-PLAN.md — Move testdata/apply.go → internal/testhelpers; split config struct into diffConfig + renderConfig
- [x] 12-02-PLAN.md — Update doc.go package overview + README.md with complete API surface (phases 7–11 additions)

### Phase 13: Refactor project layout: no Go files in root directory

**Goal:** Move all 13 root-level library `.go` files into a `drift/` subdirectory so the module root contains only metadata files (go.mod, go.sum, README.md, LICENSE, justfile). Update all import paths from `github.com/tylercrawford/drift` to `github.com/tylercrawford/drift/drift`, add `.gitignore`, and update documentation.
**Requirements**: LAYOUT-04
**Depends on:** Phase 12
**Plans:** 2/2 plans complete

Plans:
- [x] 13-01-PLAN.md — Move 13 root .go files to drift/ subdir; update all import paths; verify go test ./... passes
- [x] 13-02-PLAN.md — Add .gitignore; update README.md and doc.go with new import path

### Phase 14: Deep cruft removal: clean code, comments, and commit uncommitted changes

**Goal:** Commit 6 pending working-tree changes accumulated during phases 11–13, and remove the dead exported function `DiffLineMutedBackgroundColour` from the internal highlight package.
**Requirements**: CRUFT-01, CRUFT-02
**Depends on:** Phase 13
**Plans:** 1/2 plans executed

Plans:
- [x] 14-01-PLAN.md — Commit 6 pending modified files (.golangci.yml, justfile, main_test.go, diff_line.go, gutter.go, split_test.go)
- [ ] 14-02-PLAN.md — Remove dead DiffLineMutedBackgroundColour export from diffcolors.go
