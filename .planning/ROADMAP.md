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
- [x] **Phase 12: Restructure project layout** - idiomatic Go library layout (completed 2026-03-26)
- [x] **Phase 13: Refactor: no Go files in root** - move library files to drift/ subdir (completed 2026-03-26)
- [x] **Phase 14: Deep cruft removal** - commit pending changes, remove dead exports (completed 2026-03-27)
- [x] **Phase 15: Architecture-driven refactor** - IOStreams, Flags→Options→run() lifecycle (completed 2026-03-27)
- [x] **Phase 16: Fix v1.0.0 blockers** - Hirschberg Myers, WithContext validation, goreleaser (completed 2026-03-31)
- [x] **Phase 17: Address medium-priority council review issues** - Import path docs, Line.Spans removal, golden tests, bottom-aligned split pairing, term dep cleanup (completed 2026-03-31)
- [x] **Phase 18: Auto algorithm mode** - Add Auto as 4th Algorithm constant; O(N) heuristic selects Myers or Histogram based on file size and line frequency; make Auto the new default (completed 2026-04-01)

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
  5. `go.mod` exists at module root with path `github.com/tbcrawford/drift`, MIT LICENSE file present, and `just test` / `just build` / `just lint` run successfully
**Plans**: 5 plans

Plans:
- [x] 01-01-PLAN.md — Module scaffold: go.mod (github.com/tbcrawford/drift), MIT LICENSE, justfile, golangci-lint config
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
  5. `go install github.com/tbcrawford/drift/cmd/drift@latest` installs the binary successfully and `drift --help` runs
**Plans**: TBD

Plans:
- [x] 05-01: Implement `cmd/drift/main.go` with Cobra root command and flag definitions
- [x] 05-02: Implement file path, stdin pipe, and `--from`/`--to` raw string input handling
- [x] 05-03: Wire all flags through to library options; implement exit code logic
- [x] 05-04: Verify `go install github.com/tbcrawford/drift/cmd/drift@latest` and test all input modes end-to-end

### Phase 6: API Hardening & OSS Packaging
**Goal**: The library is ready for `v1.0.0` — public API is stable, documented, exemplified, and benchmarked
**Depends on**: Phase 5
**Requirements**: CORE-05, OSS-02, OSS-03, OSS-06, OSS-07
**Success Criteria** (what must be TRUE):
  1. `drift.New().Algorithm(drift.Histogram).WithTheme("github").Diff(a, b)` compiles and returns correct output (builder API works)
  2. Every exported type, function, and option has a godoc comment; `go doc github.com/tbcrawford/drift` renders clean documentation with no missing entries
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
| 12. Restructure project layout | 2/2 | Complete | 2026-03-26 |
| 13. Refactor: no Go files in root | 2/2 | Complete | 2026-03-26 |
| 14. Deep cruft removal | 2/2 | Complete | 2026-03-27 |
| 15. Architecture-driven refactor | 2/2 | Complete | 2026-03-27 |
| 16. Fix v1.0.0 blockers | 3/3 | Complete | 2026-03-31 |
| 17. Address medium-priority council review issues | 5/5 | Complete    | 2026-03-31 |
| 18. Auto algorithm mode | 1/1 | Complete    | 2026-04-01 |

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

**Goal:** Move all 13 root-level library `.go` files into a `drift/` subdirectory so the module root contains only metadata files (go.mod, go.sum, README.md, LICENSE, justfile). Update all import paths from `github.com/tbcrawford/drift` to `github.com/tbcrawford/drift/drift`, add `.gitignore`, and update documentation.
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
**Plans:** 2/2 plans complete

Plans:
- [x] 14-01-PLAN.md — Commit 6 pending modified files (.golangci.yml, justfile, main_test.go, diff_line.go, gutter.go, split_test.go)
- [x] 14-02-PLAN.md — Remove dead DiffLineMutedBackgroundColour export from diffcolors.go

### Phase 15: Architecture-driven refactor: apply ARCHITECTURE.md principles to library and CLI

**Goal:** Eliminate global state, init() flag registration, and direct os.Stderr writes from cmd/drift by introducing IOStreams injection and a Flags → Options → run() lifecycle that matches the ARCHITECTURE.md canonical pattern.
**Requirements**: ARCH-01, ARCH-02, ARCH-03, ARCH-04
**Depends on:** Phase 14
**Plans:** 2/2 plans complete

Plans:
- [x] 15-01-PLAN.md — Define IOStreams struct + rootFlags/rootOptions lifecycle contracts (iostreams.go, flags.go)
- [x] 15-02-PLAN.md — Rewrite cmd/drift/main.go: newRootCmd(), runRoot(opts), runCLI(IOStreams) — no globals, no init()

### Phase 17: Address medium-priority council review issues

**Goal:** Resolve all five medium-priority issues identified in `.reviews/drift-library/REVIEW.md`:
(1) migrate the library package from `drift/` to the module root so the canonical import is `import "github.com/tbcrawford/drift"` (CLI gets its own `go.mod` + `go.work` workspace);
(2) remove the stub `Spans []Span` field from the public `Line` struct;
(3) add goldie v2 golden file tests for the rendering pipeline;
(4) improve `pairHunkLines` to use bottom-aligned split pairing matching git's behavior;
(5) audit and clean up the dual `golang.org/x/term` + `charmbracelet/x/term` direct dependencies.
**Requirements**: REVIEW-05, REVIEW-06, REVIEW-07, REVIEW-08, REVIEW-09
**Depends on:** Phase 16
**Plans:** 5/5 plans complete

Plans:
- [x] 17-01-PLAN.md — Library-to-root migration: git mv 13 library files from drift/ to root, create cmd/drift/go.mod + go.work, update all import sites
- [x] 17-02-PLAN.md — Remove Line.Spans stub: remove Spans field from Line struct; keep Span as internal-only type
- [x] 17-03-PLAN.md — Golden file tests: add goldie v2 snapshot tests for unified, split, and no-color rendering at testdata/golden/
- [x] 17-04-PLAN.md — pairHunkLines bottom-aligned pairing: surplus deletes/inserts align to bottom of asymmetric block
- [x] 17-05-PLAN.md — Dual term dep cleanup: audit import sites, consolidate to one direct term dependency

### Phase 18: Auto algorithm mode

**Goal:** Add `Auto` as a fourth `Algorithm` constant that intelligently selects between Myers and Histogram at diff-time using an O(N) heuristic: use Histogram for files ≤ 2000 total lines where no old-side line appears > 32 times; use Myers otherwise. Make `Auto` the new default in `defaultConfig()`. Patience is excluded from Auto selection.
**Research:** `.planning/research/AUTO-ALGORITHM.md`
**Depends on:** Phase 17
**Plans:** 1/1 plans complete

Plans:
- [x] 18-01-PLAN.md — Auto constant, selectAuto() heuristic, default change, CLI + tests

### Phase 16: Fix v1.0.0 blockers: Hirschberg Myers, WithContext validation, goreleaser

**Goal:** Resolve three v1.0.0 blockers: (1) replace the O((N+M)²) trace-snapshot Myers implementation with a Hirschberg linear-space divide-and-conquer variant; (2) add validation so `WithContext(-1)` returns an error from `drift.Diff()` instead of silently expanding context to the entire file; (3) create `.goreleaser.yaml` for multi-platform binary releases (darwin/amd64, darwin/arm64, linux/amd64, windows/amd64).
**Requirements**: PERF-01, API-01, OSS-04
**Depends on:** Phase 15
**Plans:** 3/3 plans complete

Plans:
- [x] 16-01-PLAN.md — Hirschberg linear-space Myers: replace trace-snapshot O((N+M)²) with divide-and-conquer O(N+M) space
- [x] 16-02-PLAN.md — WithContext validation: drift.Diff() returns error for negative contextLines
- [x] 16-03-PLAN.md — goreleaser config: .goreleaser.yaml for multi-platform CLI binary releases

### Phase 19: add pager support for large diffs that automatically gets invoked in tty terminal instances

**Goal:** When stdout is a TTY and the diff output is taller than the terminal height, automatically invoke a pager (`$PAGER`, `less -R`, or `more`) so users can scroll large diffs. Add `--no-pager` flag to bypass. Piped output is never paged.
**Requirements**: PAGER-01, PAGER-02, PAGER-03
**Depends on:** Phase 18
**Plans:** 2/2 plans complete

Plans:
- [x] 19-01-PLAN.md — Pager primitives: resolvePager, shouldPage, startPager in cmd/drift/pager.go with unit tests
- [x] 19-02-PLAN.md — Wire pager into runRoot: render-to-buffer, line count, --no-pager flag, integration tests

### Phase 20: add directory diff support with automatic pager and file name headers

**Goal:** When two directory paths are given as positional arguments, drift walks both trees, diffs every file that differs (or is exclusive to one side), and prints a `=== <filename> ===` header before each per-file diff. Added files diff against empty; removed files diff empty against the old content. Identical directories exit 0 silently. The existing pager from Phase 19 handles large multi-file output automatically.
**Requirements**: DIR-01, DIR-02, DIR-03
**Depends on:** Phase 19
**Plans:** 2/2 plans complete

Plans:
- [x] 20-01-PLAN.md — Directory diff primitives: filePair type + diffDirectories walker in cmd/drift/dirwalk.go with unit tests
- [x] 20-02-PLAN.md — Wire directory diff into runRoot: dir detection, file name headers, runDirectoryDiff orchestrator, integration tests

### Phase 21: respect gitignore rules where necessary

**Goal:** When walking directories (both `diffDirectories` and `gitDirectoryVsHEAD`), skip files that git marks as ignored so that build artifacts, vendor directories, and other gitignored files never appear in diff output. Uses `git check-ignore -z --stdin` for per-directory filtering with fail-open behavior when git is unavailable.
**Requirements**: GITIGNORE-01, GITIGNORE-02
**Depends on:** Phase 20
**Plans:** 1/1 plans complete

Plans:
- [x] 21-01-PLAN.md — filterGitIgnored helper + gitDirectoryVsHEAD + diffDirectories gitignore filtering

### Phase 22: replace all git actions by the library or cli in this repository with https://github.com/go-git/go-git

**Goal:** Replace all `exec.Command("git", ...)` subprocess calls in `cmd/drift/gitworktree.go` and `cmd/drift/dirwalk.go` with pure-Go equivalents using `github.com/go-git/go-git/v5`, eliminating the runtime dependency on the `git` binary. Update tests to use real in-memory/on-disk go-git repos instead of fake shell scripts.
**Requirements**: GIT-01
**Depends on:** Phase 21
**Plans:** 2/2 plans complete

Plans:
- [x] 22-01-PLAN.md — Add go-git dependency + rewrite gitworktree.go with go-git API (no more exec.Command)
- [x] 22-02-PLAN.md — Rewrite gitworktree_test.go with go-git test repos + full module test verification

Plans:
- [x] TBD (run /gsd:plan-phase 22 to break down) (completed 2026-04-02)

### Phase 23: performance analysis and optimization

**Goal:** Identify and eliminate the dominant rendering performance bottlenecks via CPU profiling: cache gutter lipgloss.Style objects per render call (eliminating per-line `lipgloss.NewStyle()` in the gutter hot path) and replace per-token `lipgloss.Style` allocations in `HighlightLineWithLineBackground` with a direct ANSI SGR escape builder. Measure improvement against a baseline benchmark suite that covers the color rendering path.
**Requirements**: PERF-01, PERF-02, PERF-03, PERF-04
**Depends on:** Phase 22
**Plans:** 2 plans

Plans:
- [ ] 23-01-PLAN.md — Baseline benchmarks (color path) + GutterStyleCache to eliminate per-line lipgloss.NewStyle()
- [ ] 23-02-PLAN.md — Replace per-token lipgloss.Style in HighlightLineWithLineBackground with direct ANSI SGR builder

### Phase 24: performance optimization phase 2: target sub-500ms response time for drift --split on real repos, benchmark against delta

**Goal:** Replace `wt.Status()` (go-git's full filesystem scan — 4,500ms) with a fast index+mtime based change detection (23ms), then eliminate redundant repo opens and stream diff output to the pager incrementally. Target: < 500ms wall-clock on auth0-tenant-config (down from 4.5s baseline).
**Requirements**: PERF-01
**Depends on:** Phase 23
**Plans:** 2 plans

Plans:
- [ ] 24-01-PLAN.md — Replace wt.Status() with changedFilesViaIndex (index+mtime, 200x faster) + wall-clock benchmark
- [ ] 24-02-PLAN.md — Eliminate redundant repo opens + streaming render to pager
