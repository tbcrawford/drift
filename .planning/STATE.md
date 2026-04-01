---
gsd_state_version: 1.0
milestone: v1.0.0
milestone_name: "blockers: Hirschberg Myers, WithContext validation, goreleaser"
status: Milestone complete
stopped_at: Completed 17-05-PLAN.md
last_updated: "2026-03-31T17:19:55.483Z"
progress:
  total_phases: 17
  completed_phases: 6
  total_plans: 16
  completed_plans: 16
---

# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-03-25)

**Core value:** A Go developer can `import "github.com/tbcrawford/drift"` and get a production-quality, richly-rendered diff with one function call — the same quality they see in GitHub's PR review UI but in the terminal.
**Current focus:** Phase 17 — address-medium-priority-council-review-issues

## Current Position

Phase: 17
Plan: Not started

## Performance Metrics

**Velocity:**

- Total plans completed: 0
- Average duration: —
- Total execution time: 0 hours

**By Phase:**

| Phase | Plans | Total | Avg/Plan |
|-------|-------|-------|----------|
| - | - | - | - |

**Recent Trend:**

- Last 5 plans: —
- Trend: —

*Updated after each plan completion*
| Phase 01-foundation P01 | 2 | 2 tasks | 5 files |
| Phase 01-foundation P02 | 83 | 2 tasks | 5 files |
| Phase 01-foundation P03 | 231 | 3 tasks | 2 files |
| Phase 01-foundation P04 | 6 | 2 tasks | 8 files |
| Phase 01-foundation P05 | 432 | 3 tasks | 5 files |
| Phase 02-algorithms P02-01 | 7min | 2 tasks | 2 files |
| Phase 12-restructure-project-layout-for-idiomatic-go-library-and-cli P12-01 | 220 | 2 tasks | 7 files |
| Phase 12-restructure-project-layout-for-idiomatic-go-library-and-cli P12-02 | 100 | 2 tasks | 2 files |
| Phase 13-refactor-project-layout-no-go-files-in-root-directory P01 | 144 | 3 tasks | 19 files |
| Phase 13-refactor-project-layout-no-go-files-in-root-directory P02 | 105 | 2 tasks | 2 files |
| Phase 14-deep-cruft-removal-clean-code-comments-and-commit-uncommitted-changes P01 | 3 | 2 tasks | 6 files |
| Phase 14-deep-cruft-removal-clean-code-comments-and-commit-uncommitted-changes P02 | 1 | 2 tasks | 2 files |
| Phase 15-architecture-driven-refactor P15-01 | 69 | 2 tasks | 2 files |
| Phase 15-architecture-driven-refactor P15-02 | 4 | 1 tasks | 2 files |
| Phase 16 P03 | 147 | 4 tasks | 2 files |
| Phase 16-fix-v1-blockers-hirschberg-myers-context-validation-goreleaser P16-02 | 406 | 4 tasks | 4 files |
| Phase 17-address-medium-priority-council-review-issues P17-01 | 12 | 3 tasks | 21 files |
| Phase 17-address-medium-priority-council-review-issues P17-04 | 4 | 2 tasks | 2 files |
| Phase 17-address-medium-priority-council-review-issues P02 | 4 | 2 tasks | 2 files |
| Phase 17-address-medium-priority-council-review-issues P03 | 8 | 2 tasks | 8 files |
| Phase 17-address-medium-priority-council-review-issues P05 | 5 | 2 tasks | 3 files |

## Accumulated Context

### Decisions

Decisions are logged in PROJECT.md Key Decisions table.
Recent decisions affecting current work:

- [Pre-Phase 1]: Lip Gloss v2 (`charm.land/lipgloss/v2`) — NOT `github.com/charmbracelet/lipgloss`; import path changed in v2
- [Pre-Phase 1]: Myers trace must be saved at END of `d` loop, not top — off-by-one invisible on short inputs
- [Pre-Phase 1]: Bubble Tea dependency should NOT be added until v2 (interactive TUI is out of scope for v1)
- [Pre-Phase 1]: Phase 2 needs dedicated research into Histogram frequency cutoff (65 vs 512?) before implementing
- [Phase 01-foundation]: go 1.21 pinned as minimum version per STACK.md despite local Go 1.26.1
- [Phase 01-foundation]: justfile established as canonical task runner; all dev workflows via just
- [Phase 01-foundation]: drift.go root package stub added so go test ./... exits 0 with no test files
- [Phase 01-foundation]: Op enum uses iota starting at Equal=0 — zero value means 'no change', safe default
- [Phase 01-foundation]: Line.Spans []Span kept nil in v1.0 — reserved for v1.x intra-line word diff, no breaking change
- [Phase 01-foundation]: Algorithm enum defined in options.go (public API) not internal/algo — it's a caller-facing choice
- [Phase 01-foundation]: Myers trace saved at END of d-loop — off-by-one invisible on short inputs, catastrophic at 100+ lines
- [Phase 01-foundation]: V array pre-allocated to 2*(N+M)+1 for O(1) forward pass allocations
- [Phase 01-foundation]: Myers edge cases (both-empty, old-empty, new-empty) handled before algorithm entry point
- [Phase 01-foundation]: internal/edittype package introduced to break import cycle (drift root → internal/algo/myers → drift root); all public types re-exported as aliases
- [Phase 01-foundation]: hunk.Build() uses edit-sequence indices (not line numbers) for context expansion — handles Insert/Delete uniformly
- [Phase 01-foundation]: Property tests compare canonical (normalized) text to handle drift's trailing-newline stripping; SliceOfN(elem, -1, 50) used for rapid v1.2.0 (no options-style MaxLen); Apply() uses 0-indexed cursor through hunks
- [Phase 02-algorithms]: Iterative stack with tagged union for patience diff, O(N*M) LCS anchors, Myers fallback for no-unique-lines sub-ranges — Avoids recursion depth issues and post-sort step; matches research section 5.1 design recommendations
- [Phase 02-algorithms]: Histogram.Diff uses tagged stackItem union (isEmit bool) — same pattern as Patience; OldLine-proxy post-sort is fundamentally uncomputable without full traversal, so sort approach was dropped entirely
- [Phase 02-algorithms]: Property tests use rapid.SliceOfN(elem, -1, 50) with drift.WithAlgorithm() option for all three algorithms; 1000 iterations each
- [Phase 03-unified-rendering/03-02]: styles.Get() returns Fallback (not nil) — use Registry[name] map lookup in SelectTheme to distinguish known vs unknown theme names
- [Phase 03-unified-rendering/03-02]: chroma.FormatterFunc is not comparable with == (panics) — test formatter selection behaviorally via output comparison, not pointer identity
- [Phase 03-unified-rendering/03-04]: RenderConfig holds pre-resolved Chroma Lexer/Style/Formatter — avoids redundant lookups per render call; callers populate cfg once, renderer uses it
- [Phase 03-unified-rendering/03-04]: Unified() fail-open on HighlightLine error — plain content used rather than returning error; diff output always appears
- [Phase 03-unified-rendering/03-05]: resolveProfile checks cfg.noColor and NO_COLOR env before *os.File detection — explicit user intent wins
- [Phase 03-unified-rendering/03-05]: colorprofile.NewWriter used for ANSI downsampling; non-file writers default to NoTTY → NoOp formatter → plain text
- [Phase 03-unified-rendering/03-05]: Public drift.Render() follows same options application pattern as drift.Diff()
- [Phase 12-restructure-project-layout-for-idiomatic-go-library-and-cli]: Apply() moved to internal/testhelpers — test-only helper should not be in testdata/ public package
- [Phase 12-restructure-project-layout-for-idiomatic-go-library-and-cli]: config split into diffConfig+renderConfig sub-structs for self-documenting separation of diff vs render options
- [Phase Phase 12]: doc.go uses # section headers (Go 1.19+ godoc style) for Functional API, Builder API, Diff Options, Render Options, Git Integration
- [Phase 13-refactor-project-layout-no-go-files-in-root-directory]: Used git mv to preserve file history when moving 13 library files from root to drift/ subdirectory; internal/ packages unchanged (visibility rule allows drift/ to import them)
- [Phase 13-refactor-project-layout-no-go-files-in-root-directory]: drift/doc.go required no import path changes — it uses only Go identifiers in examples, not import strings
- [Phase 14]: golangci-lint v2 config: version key required at top level; exclusions restructured under linters: and formatters: blocks
- [Phase 14]: gutterColumnSeparator uses U+2502 BOX DRAWINGS LIGHT VERTICAL (│), not ASCII pipe — matches Unicode box-drawing TUI convention
- [Phase 14]: DiffLineStyle returns chroma.Colour not lipgloss.Style — leaner API, callers apply directly to rendering pipeline
- [Phase 14]: DiffLineMutedBackgroundColour removed — had zero callers since phase 11 terrasort parity refactor; .gitignore root-anchored to /drift to protect drift/ library subdirectory
- [Phase Phase 15]: IOStreams defined in its own file for single-responsibility I/O abstraction contract
- [Phase Phase 15]: resolveRootOptions centralizes all I/O decisions; show-theme callback uses streams.Err not os.Stderr
- [Phase Phase 15]: runCLI signature changed from (stdout, stderr, stdin, args) to (streams IOStreams, args) — all tests updated for cleaner IOStreams-first architecture
- [Phase Phase 15]: newRootCmd(streams IOStreams) constructor produces fresh cobra.Command per invocation — eliminates need for flag-reset hacks between test runs
- [Phase 16]: goreleaser v2 config: formats plural array, snapshot.version_template, CGO_ENABLED=0 static binaries, linux/arm64+windows/arm64 excluded pending CI runners
- [Phase 16]: validate() called at Diff() time (not WithContext() time) — standard Go functional-options pattern: validate on use, not on set
- [Phase 17-01]: go.work replace directive used for local CLI module dev: replace github.com/tbcrawford/drift => ../.. in cmd/drift/go.mod enables go mod tidy without published module
- [Phase 17-01]: git mv used for all 13 library files to preserve file history when moving from drift/ to module root
- [Phase 17]: Bottom-aligned pairHunkLines: for D>I top (D-I) deletes get blank right; for I>D top (I-D) inserts get blank left — matches git split-view convention (REVIEW-08)
- [Phase 17]: Span struct retained in internal/edittype as internal-only for future word-diff; Line.Spans []Span field removed from public API at v1.0 (zero-value stub is API smell)
- [Phase 17-03]: WithTermWidth(w int) added to public API to allow deterministic split-view width in tests; wired through buildRenderPipeline
- [Phase 17-03]: Golden fixtures use WithNoColor() — plain text, no ANSI, CI-portable; goldie.WithFixtureDir(testdata/golden) for explicit fixture location
- [Phase 17-05]: Migrated palette_unix.go from golang.org/x/term to charmbracelet/x/term; charmbracelet/x/term takes uintptr fd (tty.Fd() returns uintptr directly — no int() cast needed)

### Roadmap Evolution

- Phase 7 added: support diffs from git that is, if a single file is provided and the file is in a git repo drift will show the current changes
- Phase 9 added: dual line-number gutters (old | new), centered numbers, brighter gutter vs muted code — unified and split
- Phase 10 added: theme-aware full-line add/delete diff styling (hybrid Chroma + Lip Gloss/ANSI) — depends on 9
- Phase 11 added: GitHub PR-style intra-line word-level highlights — depends on 10
- Phase 12 added: Restructure project layout for idiomatic Go library and CLI
- Phase 13 added: Refactor project layout: no Go files in root directory
- Phase 14 added: Deep cruft removal: clean code, comments, and commit uncommitted changes
- Phase 15 added: Architecture-driven refactor: apply ARCHITECTURE.md principles to library and CLI

### Quick Tasks Completed

| Date | Summary |
|------|---------|
| 2026-03-26 | Word diff: full-line `DiffLineStyle` on paired rows (unified + split); intra-line changed spans use neutral `GutterBackgroundHex` only |
| 2026-03-26 | `/gsd-plan-phase 11` — added `11-03-PLAN.md` (terrasort chroma parity, gutter-neutral words, verification refresh) |
| 2026-03-26 | `/gsd-plan-phase 11 --research` — `11-RESEARCH.md` (drift vs terrasort pipeline diff), `11-VALIDATION.md` (Nyquist map) |
| 2026-03-26 | Fix full-line background highlighting: prefix char now carries line bg; trailing whitespace extended with bg-colored spaces (terrasort parity); monokai invisible-bg bug fixed; default dark theme → github-dark |
| 2026-03-31 | Revamp README.md: centered hero image, badges, full TOC, CLI flags table, render options table, algorithms table, builder API, mobile-friendly HTML centering |
| 2026-03-31 | Add GitHub Actions CI, release, and security workflows; fix goreleaser owner typo; track dist/config.yaml in git |
| 2026-04-01 | [260401-k11] Collapse two-module workspace layout into single go.mod; delete go.work + cmd/drift/go.mod; add cobra+go-internal as direct deps; remove duplicate CI test step |

### Pending Todos

None yet.

### Blockers/Concerns

- [Phase 02-algorithms]: histogramMaxOccurrences=64 confirmed (matches jgit/Git implementation)
- [Phase 02-algorithms]: sortEdits() uses NewLine as primary key to fix out-of-order Equal emission from iterative stack
- [Phase 02-algorithms]: Myers fallback fires when findBestMatch returns found=false (no match within 64-occurrence threshold)
- [Phase 03-unified-rendering]: DetectDarkBackground short-circuits for NoTTY/Ascii profiles — avoids 2s OSC 11 timeout on piped outputs
- [Phase 03-unified-rendering]: Dark background is default (true) for non-TTY environments; most developer terminals are dark-themed

## Session Continuity

Last session: 2026-04-01T18:41:48Z
Stopped at: Completed quick task 260401-k11 (switch-to-single-module-build)
Resume file: None
