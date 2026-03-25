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

### Active

- [ ] Library exposes both functional and builder/fluent API styles
- [ ] Produces unified diff output (Git-style context hunks)
- [ ] Produces side-by-side split diff output (left/right panels)
- [ ] Chroma syntax highlighting applied per-line with diff colors (red/green) layered on top
- [ ] Auto-detects terminal color theme to select best-matching Chroma theme
- [ ] User can override Chroma theme via option/flag
- [ ] Language auto-detected from file extension; overridable via --lang flag
- [ ] CLI accepts two file paths, stdin piping, or two raw string arguments
- [ ] Single go.mod monorepo: library at root, CLI at cmd/drift/
- [ ] Standard Go OSS packaging: MIT license, go.mod, godoc comments, examples/ directory
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
| Both functional + builder API | Convenience for simple use cases, flexibility for advanced ones | — Pending |
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
*Last updated: 2026-03-25 after Phase 02: algorithms complete — Patience and Histogram algorithms implemented, wired into drift.Diff(), all property-based tests green*
