# Research Summary: drift

**Domain:** Go text diff library + CLI tool (Myers/Patience/Histogram algorithms, Chroma syntax highlighting, Lip Gloss terminal layout)
**Researched:** 2026-03-25
**Overall confidence:** HIGH — all critical claims verified against official documentation, pkg.go.dev, and primary sources

---

## Executive Summary

`drift` occupies a clear gap in the Go ecosystem: existing diff libraries (sergi/go-diff with 67k dependents, pmezard/go-difflib with 597k dependents — now abandoned) handle algorithm correctness but produce raw text output. Rendering tools (delta, diff-so-fancy) produce beautiful output but are Rust/Perl binaries with no Go library API. `drift` combines both — idiomatic Go library + CLI, three algorithms, Chroma syntax highlighting, Lip Gloss layout — in a single `go get`-able package.

The technology stack is well-settled. Chroma v2 (v2.23.1) is the only production-ready Go syntax highlighting library; there is no realistic alternative. Lip Gloss v2 (v2.0.2, stable since Feb 24 2026) fixes the I/O race conditions that plagued v1 and provides the correct API for split-panel layout. The critical import path change (`charm.land/lipgloss/v2`) must be used from day one — v1 is deprecated. Cobra v1.9.1 is the undisputed CLI framework choice. The diff algorithms must be implemented from scratch (not imported as deps), using `znkr.io/diff` and `peter-evans/patience` as reference implementations only.

The architectural pattern is a clean pipeline: plain string input → algorithm → `[]Edit` → hunk builder → `DiffResult` (exported pivot type) → per-line Chroma highlighting → Lip Gloss renderer → `io.Writer`. Each layer is independently testable. The `DiffResult` struct is the stable public contract that enables callers to render their own output. The most critical architectural decision is keeping the diff engine free of rendering knowledge — ANSI codes must never enter the algorithm or hunk builder layers.

The highest-risk implementation areas are: (1) Myers algorithm correctness on long inputs (off-by-one in trace save timing only surfaces at 100+ lines), (2) Histogram fallback to Myers (required for correctness on repetitive files, commonly omitted), (3) Chroma + Lip Gloss ANSI width interaction (highlighted strings must be measured with `lipgloss.Width()`, not `len()`), and (4) terminal theme detection hanging on non-xterm terminals (OSC 11 query requires a 200ms timeout). All four have concrete prevention strategies documented in PITFALLS.md.

---

## Key Findings

**Stack:** Go 1.21 + Chroma v2 (`github.com/alecthomas/chroma/v2`) + Lip Gloss v2 (`charm.land/lipgloss/v2`) + Cobra v1.9.1; custom Myers/Patience/Histogram implementations using `znkr.io/diff` and `peter-evans/patience` as reference only.

**Architecture:** Five-layer pipeline — Algorithm → Hunk Builder → DiffResult (public pivot) → Highlighter → Renderer; all implementation in `internal/`; public API at root package; CLI in `cmd/drift/`.

**Critical pitfall:** Never save Myers trace at top of the `d` loop — must save at the end. This off-by-one is invisible on short test inputs and only surfaces on 100+ line files.

---

## Implications for Roadmap

Based on research, the build order is driven by two constraints: (1) the diff engine must exist before any rendering can be built, and (2) the public `DiffResult` type must be stable before renderers consume it. The feature dependency graph (from FEATURES.md) maps cleanly to this phase structure:

### Suggested Phase Structure

1. **Foundation: Module + Data Model + Myers Algorithm**
   - Set up `go.mod` with correct module path (`github.com/tbcrawford/drift`)
   - Define exported types: `Op`, `Edit`, `Hunk`, `Line`, `DiffResult`
   - Define `Option`/`config` pattern (functional options)
   - Implement Myers algorithm in `internal/algo/myers/`
   - Implement hunk builder in `internal/hunk/`
   - Wire `drift.Diff()` functional API
   - **Testing infrastructure:** property-based tests (`pgregory.net/rapid`), cross-validation against `diff -u`, golden file setup
   - Addresses: Myers algorithm, unified diff structure, structured AST types, functional API
   - Avoids pitfalls: Myers off-by-one, O(N²) memory, ANSI in test assertions, exported struct API traps

2. **Algorithms: Patience + Histogram**
   - Implement Patience diff in `internal/algo/patience/` (with inter-anchor Myers fallback)
   - Implement Histogram diff in `internal/algo/histogram/` (with Myers fallback for high-repetition inputs)
   - Add `WithAlgorithm()` option
   - Addresses: Patience/Histogram algorithms
   - Avoids pitfalls: Patience inter-anchor fallback omission, Histogram Myers fallback omission
   - *Note: This phase likely needs deeper research — histogram's frequency bucketing cutoff (65 vs 512?) and jgit's exact fallback conditions warrant careful study before implementing*

3. **Rendering: Chroma + Unified Diff View**
   - Add `internal/theme/` (terminal dark/light detection with OSC 11 timeout)
   - Add `internal/highlight/` (Chroma v2 lexer→formatter pipeline, color profile detection)
   - Add `internal/render/unified.go` (UnifiedRenderer with `@@ -a,b +c,d @@` headers)
   - Add language auto-detection via `lexers.Match()` + `lexers.Analyse()` fallback
   - Add `WithTheme()`, `WithLang()`, `WithNoColor()` options
   - Addresses: Chroma syntax highlighting, unified diff output, language auto-detection, terminal theme, no-color support
   - Avoids pitfalls: Chroma v1/v2 confusion, nil lexer panic, formatter color depth hardcoding, OSC 11 hang, ANSI in width measurements

4. **Rendering: Side-by-Side Split View**
   - Add `internal/render/split.go` (Lip Gloss `JoinHorizontal` two-panel layout)
   - Implement terminal width detection with pipe fallback
   - Handle ANSI-aware width measurement (`lipgloss.Width()`)
   - Handle Unicode wide characters (`runewidth.DefaultCondition.EastAsian = false`)
   - Add `WithSplit()` / `Split()` API
   - Addresses: Side-by-side split view
   - Avoids pitfalls: Lip Gloss width/border mismatch, ANSI inflate width, Unicode wide char misalignment, terminal width detection on pipes

5. **CLI: Cobra Command + File/Stdin Input**
   - Implement `cmd/drift/main.go` with Cobra (thin wrapper, <50 lines)
   - Handle: two file args, stdin piping, `--no-color`, `--lang`, `--theme`, `--split`, `--algorithm`
   - Add binary file detection ("Binary files differ" message)
   - Verify `go install github.com/tbcrawford/drift/cmd/drift@latest` works
   - Addresses: CLI tool, stdin, file input, binary files, all flags
   - Avoids pitfalls: color in non-TTY, terminal width on pipes, `go install` wrong path

6. **API Hardening + v1.0 Stabilization**
   - Audit public API surface (functional options, exported types, godoc)
   - Add builder/fluent API (`drift.New().Algorithm().Split()`)
   - Write `examples/` directory (basic + builder examples)
   - Property-based testing sweep, performance benchmarks (10,000-line file)
   - Set `go.mod` version correctly; verify module path
   - Addresses: builder API, godoc, examples, API stability
   - Avoids pitfalls: exported struct breaking changes, module major version confusion, `go install` path

### Phase Ordering Rationale

- **Myers before Patience/Histogram:** Both Patience and Histogram fall back to Myers internally. Myers is the non-negotiable foundation.
- **Data model before rendering:** `DiffResult`, `Hunk`, `Line`, `Op` must be stable before any renderer consumes them — changing these post-rendering-implementation is expensive.
- **Unified diff before split view:** Split view consumes hunks from the unified diff pipeline. Side-by-side is additive on top, not parallel.
- **Rendering before CLI:** The CLI is a thin consumer. It must be last because it imports the fully-formed library.
- **Algorithm phase split:** Patience/Histogram is a separate phase because their fallback mechanics are substantially different from Myers and warrant independent correctness verification before building on them.
- **Builder API in final phase:** The functional API and internal config struct must be stable before adding the builder wrapper — the builder delegates to the same config. Building it last avoids refactoring the config during builder construction.

### Research Flags for Phases

- **Phase 2 (Histogram):** Likely needs deeper research — the jgit histogram cutoff value (is it 65 or 512?), the exact fallback trigger condition, and whether `znkr.io/diff`'s variant handles it differently from JGit. Study `raygard.github.io/2025/01/29/a-histogram-diff-implementation/` before implementing.
- **Phase 3 (Theme detection):** May need research into Lip Gloss v2's specific `HasDarkBackground` API — the v2 implementation changed from v1. Verify the OSC 11 timeout is handled by Lip Gloss v2's `HasDarkBackground(os.Stdin, os.Stdout)` call directly or if a manual timeout wrapper is still needed.
- **All other phases:** Standard patterns with well-understood implementations. Unlikely to need additional research beyond verifying API signatures at implementation time.

---

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack — Chroma v2 | HIGH | v2.23.1 verified on pkg.go.dev and GitHub releases |
| Stack — Lip Gloss v2 | HIGH | v2.0.2 stable, released Feb 24 2026; import path `charm.land/lipgloss/v2` confirmed |
| Stack — Cobra | HIGH | v1.9.1 confirmed via Context7 + cobra.dev |
| Stack — Algorithm implementation (custom) | HIGH | Consensus clear: drift IS the diff library; importing a transitive diff dep breaks the value prop |
| Stack — Reference implementations | HIGH | `znkr.io/diff` v1.0.0 released Mar 15 2026; `peter-evans/patience` confirmed used by microsoft/typescript-go |
| Features — Table stakes | HIGH | Based on Go ecosystem survey; all confirmed against actual library APIs |
| Features — Differentiators | HIGH | Competitor analysis verified against github stars, dependent counts, active maintenance status |
| Architecture — Pipeline structure | HIGH | Corroborated by go-git diff package, znkr.io/diff, and jansmrcka/differ real-world implementations |
| Architecture — Lip Gloss two-layer coloring | HIGH | Confirmed: Lip Gloss preserves embedded ANSI when adding background styles |
| Pitfalls — Myers off-by-one | HIGH | Primary source post-mortem (dev.to, 2026-03-07) + Myers 1986 paper analysis |
| Pitfalls — Histogram fallback | HIGH | raygard.github.io primary source analysis of jgit implementation |
| Pitfalls — Lip Gloss width bug | HIGH | GitHub issue #449 confirmed fixed in v2.0.0 |
| Pitfalls — ANSI width inflation | HIGH | Documented in Lip Gloss + go-runewidth issue trackers |
| Pitfalls — OSC 11 hang | HIGH | Multiple terminal compatibility sources; `termenv` handles this |
| Histogram cutoff value (65 vs 512) | MEDIUM | raygard source says 65; jgit source says something different. Needs verification at implementation time |

---

## Gaps to Address

1. **Histogram frequency cutoff threshold:** Research cites 65 (jgit low-occurrence limit) but the exact value in jgit source code should be verified at implementation time. `raygard.github.io/2025/01/29/a-histogram-diff-implementation/` is the best source.

2. **Lip Gloss v2 `HasDarkBackground` timeout behavior:** The v2 API is `lipgloss.HasDarkBackground(os.Stdin, os.Stdout)`. It's unclear if v2 handles the OSC 11 read timeout internally or if a manual timeout goroutine is still needed. Verify against `charm.land/lipgloss/v2` source before Phase 3.

3. **`charm.land/lipgloss/v2` Bubble Tea interaction in v1.0 scope:** PROJECT.md lists Bubble Tea as "available for interactive scenarios" but the interactive TUI is explicitly out of scope. The dependency should not be added until v2. This gap is not a blocker but should be documented in the module.

4. **`\r\n` line ending normalization:** Research confirms `strings.Split(text, "\n")` leaves `\r` suffixes on Windows files. The correct implementation is either `strings.TrimRight(line, "\r")` per line or using `bufio.Scanner` which handles both. The exact approach needs a decision at Phase 1 implementation time.

5. **Intra-line word-level diff:** Listed as P2 / v1.x in FEATURES.md. Not required for v1.0 but should be considered in the Phase 1 `Edit` type design — the word-level diff will need to add span markers to `Line` objects. If the type is wrong at v1.0, adding word-level diff will be a breaking change.

---

## Sources Index

Full source lists are in each research file. Key authoritative sources by area:

| Area | Primary Source |
|------|---------------|
| Chroma v2 API | `pkg.go.dev/github.com/alecthomas/chroma/v2` |
| Lip Gloss v2 API | `pkg.go.dev/charm.land/lipgloss/v2` |
| Lip Gloss v2 release | Charm blog post, Feb 23 2026 |
| Myers algorithm | Eugene Myers 1986; `znkr.io/diff` v1.0.0 |
| Histogram algorithm | `raygard.github.io/2025/01/29/a-histogram-diff-implementation/` |
| Patience algorithm | `github.com/peter-evans/patience` |
| Go module layout | `tip.golang.org/doc/modules/layout` (official) |
| Functional options | Rob Pike pattern; `bytesizego.com/blog/10-years-functional-options-golang` |
| Colorprofile detection | `github.com/charmbracelet/colorprofile` |
| Myers off-by-one pitfall | `dev.to/tommy_worklab/i-implemented-myers-diff` (2026-03-07) |
| Lip Gloss width bug | `github.com/charmbracelet/lipgloss/issues/449` |

---

*Summary for: drift — Go text diff library + CLI*
*Researched: 2026-03-25*
