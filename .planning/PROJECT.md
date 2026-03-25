# drift

## What This Is

`drift` is a production-ready Go library and CLI for comprehensive text diffing between two multi-line strings or files. It is inspired by Git's diff engine and GitHub's unified diff rendering, providing structured unified and side-by-side (split) diff output with rich syntax highlighting via Chroma. The library is designed for open-source distribution and idiomatic use in other Go projects.

## Core Value

A Go developer can `import "github.com/tylercrawford/drift"` and get a production-quality, richly-rendered diff with one function call — the same quality they see in GitHub's PR review UI but in the terminal.

## Requirements

### Validated

- [x] Single go.mod monorepo: library at root, CLI at cmd/drift/ — Validated in Phase 01: Foundation
- [x] Standard Go OSS packaging: MIT license, go.mod, godoc comments — Validated in Phase 01: Foundation
- [x] Supports Myers diff algorithm — Validated in Phase 01: Foundation
- [x] Library exposes functional Option API style — Validated in Phase 01: Foundation
- [x] Property-based round-trip: `Apply(Diff(a,b), a) == b` holds for all inputs — Validated in Phase 01: Foundation
- [x] Supports three diff algorithms: Myers, Patience, Histogram (selectable per call) — Validated in Phase 02: algorithms

### Validated

- [x] Produces unified diff output (Git-style context hunks) — Validated in Phase 03: unified-rendering
- [x] Chroma syntax highlighting applied per-line via HighlightLine pipeline — Validated in Phase 03: unified-rendering
- [x] Auto-detects terminal color theme (OSC 11 guard, NoTTY short-circuit) — Validated in Phase 03: unified-rendering
- [x] User can override Chroma theme via WithTheme() option — Validated in Phase 03: unified-rendering
- [x] Language auto-detected from file extension; overridable via WithLang() — Validated in Phase 03: unified-rendering
- [x] Color depth detection + graceful degradation (TrueColor→ANSI256→ANSI→NoTTY) — Validated in Phase 03: unified-rendering
- [x] NO_COLOR env var and WithNoColor() suppresses all ANSI sequences — Validated in Phase 03: unified-rendering
- [x] Produces side-by-side split diff output (left/right panels) — Validated in Phase 04: split-rendering
- [x] CLI at `cmd/drift` (Cobra): file paths, stdin `-`, `--from`/`--to`, flags for algorithm/context/theme/lang/no-color/split — Validated in Phase 05: cli
- [x] CLI exit codes 0 (no diff), 1 (diff), 2 (errors); testscript integration tests — Validated in Phase 05: cli
- [x] CLI single path inside a Git worktree: diff working tree file vs `HEAD` via `git` subprocess (`GIT_TERMINAL_PROMPT=0`); errors mention `git` and two-path fallback — Validated in Phase 07: support-diffs-from-git
- [x] Library exposes builder/fluent API (`drift.New()` chain delegating to functional options) — Validated in Phase 06: api-hardening-oss-packaging
- [x] Runnable `examples/basic` and `examples/builder`; root README (install, CLI, library, rendering); godoc package overview — Validated in Phase 06: api-hardening-oss-packaging
- [x] Benchmarks for large (~10k line) diff and unified/split render — Validated in Phase 06: api-hardening-oss-packaging

### Active

- [ ] Lip Gloss used for terminal layout and styling; Bubble Tea available for interactive scenarios

### Out of Scope

- HTML render target — terminal ANSI only for v1; web rendering deferred
- Interactive scrollable TUI (vimdiff-style) — static output only for v1
- Real-time / streaming diff (watching files) — not a v1 goal
- Separate library and CLI modules — single module simplifies distribution

## Context

- Inspired by Git's diff algorithm and GitHub's PR diff rendering (unified + split views)
- Chroma (github.com/alecthomas/chroma) is the established Go syntax highlighting library
- Lip Gloss (github.com/charmbracelet/lipgloss) is the dominant Go terminal layout/styling library
- Bubble Tea (github.com/charmbracelet/bubbletea) is the established Go TUI framework (same org)
- Myers algorithm is the standard in most diff tools; Patience/Histogram handle code-specific edge cases (moved blocks, refactors) better
- Go module layout: single go.mod at root, `package drift` importable at `github.com/tylercrawford/drift`, CLI at `cmd/drift/main.go`
- Terminal theme auto-detection should inspect $COLORFGBG, $TERM, or terminal background color queries to pick a light vs dark Chroma theme

## Constraints

- **Tech Stack**: Go stdlib + Chroma + Lip Gloss — no unnecessary dependencies
- **API Stability**: Public API should be stable enough to version at v1.0.0
- **Compatibility**: Go 1.21+ (generics available, modern stdlib)
- **Distribution**: `go install github.com/tylercrawford/drift/cmd/drift@latest` must work
- **Importability**: Library usable with `go get github.com/tylercrawford/drift` by third-party projects

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Library can use deps (Chroma, Lip Gloss) | Rendering is a core library capability, not just CLI concern | — Pending |
| Single go.mod monorepo | Simpler distribution; `go install` and `go get` both work from one module | — Pending |
| Terminal-only output for v1 | Keeps scope tight; HTML rendering is additive and can come later | — Pending |
| Both functional + builder API | Convenience for simple use cases, flexibility for advanced ones | `New()` + chain methods in Phase 06 |
| Three algorithms (Myers + Patience + Histogram) | Different algorithms shine on different input types; user should choose | — Pending |
| Lip Gloss for split view layout | Industry standard for Go terminal layouts; pairs naturally with Chroma ANSI output | — Pending |

## Evolution

This document evolves at phase transitions and milestone boundaries.

**After each phase transition** (via `/gsd-transition`):
1. Requirements invalidated? → Move to Out of Scope with reason
2. Requirements validated? → Move to Validated with phase reference
3. New requirements emerged? → Add to Active
4. Decisions to log? → Add to Key Decisions
5. "What This Is" still accurate? → Update if drifted

**After each milestone** (via `/gsd-complete-milestone`):
1. Full review of all sections
2. Core Value check — still the right priority?
3. Audit Out of Scope — reasons still valid?
4. Update Context with current state

---
*Last updated: 2026-03-25 after Phase 07 complete — CLI single-arg git mode (working tree vs HEAD), README example; milestone roadmap phase 7 marked complete*
