# Stack Research

**Domain:** Go text diff library + CLI tool
**Researched:** 2026-03-25
**Confidence:** HIGH — all versions verified against official releases and pkg.go.dev as of research date

---

## Recommended Stack

### Core Technologies

| Technology | Import Path | Version | Purpose | Why Recommended |
|------------|-------------|---------|---------|-----------------|
| Go stdlib | — | 1.21+ | Language, `testing`, `bufio`, `strings`, `io`, `os` | Generics available; `slices`/`maps` packages reduce boilerplate; minimum viable for target users |
| Chroma v2 | `github.com/alecthomas/chroma/v2` | **v2.23.1** | Syntax highlighting → ANSI token stream | Only production-ready Go syntax highlighter; 300+ language lexers; MIT license; powers Hugo, Glamour, and the Go playground |
| Lip Gloss v2 | `charm.land/lipgloss/v2` | **v2.0.2** | Terminal layout — side-by-side panels, column widths, adaptive colors | Released stable Feb 24 2026; v2 fixes I/O races that plagued v1; `JoinHorizontal`, `Width`, and `HasDarkBackground` are exactly the APIs needed for split-view diff rendering |
| Bubble Tea v2 | `charm.land/bubbletea/v2` | **v2.0.2** | Interactive TUI (optional, future) | Same org as Lip Gloss; v2 designed to work lockstep with Lip Gloss v2; defer for post-v1 interactive scrollable mode |
| Cobra | `github.com/spf13/cobra` | **v1.9.1** | CLI command/flag parsing | Industry standard for Go CLIs; used by kubectl, Hugo, GitHub CLI; subcommands, persistent flags, auto-complete, `RunE` pattern |

### Diff Algorithm Layer

**Decision: Implement all three algorithms yourself using `znkr.io/diff` and `github.com/sergi/go-diff` as reference, not as dependencies.**

Rationale: `drift` *is* the diff library. Importing another diff library as a runtime dependency would create a layering problem — users expect to depend on `drift`, not a transitive diff dep they can't control. Study the reference implementations, implement cleanly.

| Reference (study only) | Version | Algorithm | Why Study It |
|------------------------|---------|-----------|--------------|
| `znkr.io/diff` | v1.0.0 (Mar 15 2026) | Myers + heuristics | Best-in-class 2026 Go Myers; includes readability heuristics; Apache-2 licensed; benchmarks show 2–5× faster than sergi/go-diff |
| `github.com/sergi/go-diff` | v1.4.0 | Myers (Google Diff-Match-Patch) | Most widely used Go diff today; character/word/line granularity; study the `DiffMainLines` pattern |
| `github.com/peter-evans/patience` | v0.3.0 | Patience | Clean Go patience impl; only one that ships unified diff output natively; small codebase (~300 LOC) |
| `golang.org/x/tools/internal/diff/myers` | x/tools | Myers (deprecated) | Internal Go tools reference; formally deprecated, but canonical test cases are worth studying |

**Algorithm summary for implementation:**

- **Myers** (default): O(ND), minimal edit distance, fastest; use with readability heuristics (boundary shifting to blank lines/logical breaks)
- **Patience**: anchor on unique common lines → better for refactors; falls back to Myers when no unique lines exist
- **Histogram**: extends patience by frequency-bucketing (rare lines preferred as anchors); Git's preferred algorithm since 2011; better than patience on large files with repeated structure; falls back to Myers when all lines appear > 64 times

### Supporting Libraries

| Library | Import Path | Version | Purpose | When to Use |
|---------|-------------|---------|---------|-------------|
| `github.com/charmbracelet/colorprofile` | charmbracelet/colorprofile | v0.4.2 | Terminal color capability detection (TrueColor / ANSI256 / ANSI / NoTTY) | Always — needed by Lip Gloss v2 writer setup; also useful for graceful no-color fallback when stdout is piped |
| `github.com/mattn/go-isatty` | go-isatty | v0.0.20 | Is stdout a TTY? | Use in CLI to suppress color when piped; `isatty.IsTerminal(os.Stdout.Fd())` |
| `pgregory.net/rapid` | rapid | v1.2.0 | Property-based testing | Phase 1 testing: verify `Apply(Diff(a,b), a) == b` holds for all inputs; verify round-trip invariants |
| `github.com/sebdah/goldie/v2` | goldie | v2.8.0 | Golden file testing for rendered diff output | Snapshot the ANSI output of known diffs; catch rendering regressions; `go test -update` regenerates |
| `github.com/rogpeppe/go-internal` | go-internal | v1.14.x | `testscript` — CLI integration testing via txtar scripts | Test the `drift` CLI binary end-to-end; captures stdout/stderr; handles exit codes; the way Go tools team tests CLI tools |

### Development Tools

| Tool | Purpose | Notes |
|------|---------|-------|
| `golangci-lint` | Linter aggregator | Run `staticcheck`, `govet`, `unused`, `errcheck`; config at `.golangci.yml`; used by Charm's own repos |
| `goreleaser` | Multi-platform binary releases | Produces `darwin/amd64`, `darwin/arm64`, `linux/amd64`, `windows/amd64` archives; works with `go install` |
| `go test -fuzz` | Built-in fuzzing | Go 1.21+ fuzzing for diff algorithm edge cases; complement to `rapid`; no extra dep |

---

## Installation

```bash
# Initialize module (single module, library at root)
go mod init github.com/tbcrawford/drift

# Core rendering dependencies
go get github.com/alecthomas/chroma/v2
go get charm.land/lipgloss/v2
go get github.com/charmbracelet/colorprofile

# CLI framework
go get github.com/spf13/cobra@v1.9.1

# Dev / test only
go get pgregory.net/rapid
go get github.com/sebdah/goldie/v2
go get github.com/rogpeppe/go-internal/testscript

# Optional: Bubble Tea (add when interactive scrolling mode is tackled)
# go get charm.land/bubbletea/v2
```

---

## Alternatives Considered

| Category | Recommended | Alternative | Why Not Alternative |
|----------|-------------|-------------|---------------------|
| Diff algorithm impl | Custom (study znkr.io/diff) | Import `znkr.io/diff` as dep | `drift` *is* the diff library; opaque dep breaks the value prop; licensing complication (Apache-2 vs MIT) |
| Diff algorithm impl | Custom | Import `sergi/go-diff` | Same problem; also `sergi/go-diff` produces fragmented word-level output unsuitable for line-level unified diff |
| Syntax highlighting | Chroma v2 | `muesli/termenv` + manual coloring | Chroma is the *only* Go solution with full language lexer support; `termenv` only handles color output, not tokenization |
| Terminal layout | Lip Gloss v2 | Manual ANSI escape strings | Lip Gloss handles Unicode width, color downsampling, and TTY detection that manual ANSI doesn't; Lip Gloss v2 is now stable |
| Terminal layout | Lip Gloss v2 | Lip Gloss v1 (`github.com/charmbracelet/lipgloss`) | v1 has I/O race bugs when both Lip Gloss and Bubble Tea are active; v2 (stable Feb 24 2026) fixes this; import path changed to `charm.land/lipgloss/v2` |
| CLI framework | Cobra v1.9.1 | `urfave/cli` | Cobra is the community default for Go CLIs; better shell completion; used by the tools `drift` is most similar to (delta, difftastic) |
| CLI framework | Cobra v1.9.1 | `flag` (stdlib) | `flag` lacks subcommands, persistent flags, and completion; acceptable for trivial tools, not a production CLI |
| CLI framework | Cobra v1.9.1 | `urfave/cli/v3` | `urfave/cli` is solid but Cobra has broader adoption and better godoc integration; either is fine, but Cobra is the default |
| Property testing | `pgregory.net/rapid` | `testing/quick` (stdlib) | `testing/quick` has no automatic shrinking and poor generator ergonomics; `rapid` is the community standard since v1.0 (2023) |
| Golden files | `goldie/v2` | Manual `testdata/*.golden` | `goldie` adds `go test -update` flag, colored diff on failure, and JSON/file fixtures; worth the dep for a rendering-heavy library |
| Integration testing | `testscript` (go-internal) | Hand-rolled subprocess tests | `testscript` is how the Go team itself tests `go` CLI; txtar format is readable, version-controllable, and produces good diffs on failure |

---

## What NOT to Use

| Avoid | Why | Use Instead |
|-------|-----|-------------|
| `github.com/alecthomas/chroma` (v1, no `/v2`) | v1 is unmaintained; import path is deprecated; v2 API has breaking changes but is the only supported version | `github.com/alecthomas/chroma/v2` |
| `github.com/charmbracelet/lipgloss` (v1) | v1 has stdin/stdout race conditions when combined with Bubble Tea; `HasDarkBackground` API differs; deprecated by Charm | `charm.land/lipgloss/v2` |
| `github.com/charmbracelet/bubbletea` (v1) | Same I/O race issues as Lip Gloss v1; v2 is stable and works lockstep with Lip Gloss v2 | `charm.land/bubbletea/v2` (when needed) |
| `sergi/go-diff` as a runtime dep | Produces fragmented character-level diffs; not designed for line-level unified output; Google Diff-Match-Patch focus is text sync/patching, not human-readable diffs | Implement Myers/Patience/Histogram directly |
| `golang.org/x/tools/internal/diff/myers` | Package is `internal` — explicitly not importable outside Go tools; `ComputeEdits` is marked `deprecated` | Study as reference only; implement your own |
| `testing/quick` | Deprecated ergonomics; no shrinking; poor generator composition | `pgregory.net/rapid` |
| Multiple `go.mod` files (workspace mode) | Splits `go get` and `go install` into separate operations; confuses importers; Go docs explicitly recommend single module for lib+CLI | Single `go.mod` at repo root with `cmd/drift/main.go` |

---

## Stack Patterns by Variant

**Rendering Chroma ANSI output (the critical path):**

```go
import (
    "github.com/alecthomas/chroma/v2"
    "github.com/alecthomas/chroma/v2/formatters"
    "github.com/alecthomas/chroma/v2/lexers"
    "github.com/alecthomas/chroma/v2/styles"
)

// 1. Get lexer (by filename extension or explicit language)
lexer := lexers.Match("main.go")
if lexer == nil {
    lexer = lexers.Fallback
}
lexer = chroma.Coalesce(lexer)

// 2. Get style (theme — "monokai" for dark, "monokailight" for light)
style := styles.Get("monokai")
if style == nil {
    style = styles.Fallback
}

// 3. Get ANSI formatter ("terminal16m" = TrueColor, "terminal256" = 256-color, "terminal" = 8-color)
formatter := formatters.Get("terminal16m")

// 4. Tokenize and format to an io.Writer
iterator, err := lexer.Tokenise(nil, sourceCode)
err = formatter.Format(w, style, iterator)
```

Formatter names: `"terminal16m"` (TrueColor), `"terminal256"` (256-color), `"terminal"` (8-color ANSI). Use `colorprofile.Detect()` to pick the right one at runtime.

**Terminal theme auto-detection (Lip Gloss v2 pattern):**

```go
import (
    "charm.land/lipgloss/v2"
    "os"
)

// Query terminal background — blocks briefly, should be done once at startup
hasDarkBG := lipgloss.HasDarkBackground(os.Stdin, os.Stdout)
lightDark := lipgloss.LightDark(hasDarkBG)

// Pick Chroma style based on background
chromaTheme := lightDark("monokailight", "monokai")

// Pick diff colors based on background
addedColor   := lightDark(lipgloss.Color("#006600"), lipgloss.Color("#00AA00"))
removedColor := lightDark(lipgloss.Color("#880000"), lipgloss.Color("#CC3333"))
```

Note: `lipgloss.HasDarkBackground` sends an OSC escape sequence to query the terminal. This requires a real TTY. In tests or piped output, default to dark theme.

**Side-by-side split view layout (Lip Gloss v2):**

```go
import "charm.land/lipgloss/v2"

termWidth := 120 // get from terminal: golang.org/x/term or os.Stdout
colWidth   := (termWidth - 3) / 2  // 3 chars for divider

leftStyle := lipgloss.NewStyle().Width(colWidth)
rightStyle := lipgloss.NewStyle().Width(colWidth)

left  := leftStyle.Render(leftContent)
right := rightStyle.Render(rightContent)

output := lipgloss.JoinHorizontal(lipgloss.Top, left, " │ ", right)
```

**go.mod structure for single-module library + CLI:**

```
drift/                          # module: github.com/tbcrawford/drift
├── go.mod                      # module github.com/tbcrawford/drift; go 1.21
├── go.sum
├── diff.go                     # package drift — public API
├── options.go                  # package drift — functional options
├── unified.go                  # package drift — unified format renderer
├── split.go                    # package drift — side-by-side renderer
├── highlight.go                # package drift — Chroma integration
├── internal/
│   ├── myers/                  # package myers — algorithm implementation
│   ├── patience/               # package patience — algorithm implementation
│   └── histogram/              # package histogram — algorithm implementation
├── cmd/
│   └── drift/
│       └── main.go             # package main — thin CLI wrapper
└── examples/
    └── basic/
        └── main.go
```

Import the library: `go get github.com/tbcrawford/drift`
Install the CLI: `go install github.com/tbcrawford/drift/cmd/drift@latest`

**Cobra CLI root command pattern:**

```go
// cmd/drift/main.go
package main

import (
    "os"
    "github.com/spf13/cobra"
    "github.com/tbcrawford/drift"
)

func main() {
    root := &cobra.Command{
        Use:           "drift <file1> <file2>",
        Short:         "Rich terminal diff with syntax highlighting",
        SilenceUsage:  true,
        SilenceErrors: true,
        RunE:          runDiff,
    }
    root.Flags().StringP("algorithm", "a", "myers", "diff algorithm: myers, patience, histogram")
    root.Flags().StringP("style", "s", "", "Chroma theme (default: auto)")
    root.Flags().StringP("lang", "l", "", "language for syntax highlighting (default: auto from file extension)")
    root.Flags().BoolP("split", "", false, "side-by-side split view")
    
    if err := root.Execute(); err != nil {
        os.Exit(1)
    }
}
```

**If building for Go 1.21+ only (recommended):**
- Use `slices.Equal`, `slices.Contains` from stdlib instead of adding `golang.org/x/exp`
- Use `min()`/`max()` builtins (added Go 1.21)
- `go.mod` should declare `go 1.21`

**If output is piped (no TTY):**
- `isatty.IsTerminal(os.Stdout.Fd())` returns false
- Emit plain unified diff format, no color
- `colorprofile.Detect()` returns `colorprofile.NoTTY` → use `"noop"` Chroma formatter

---

## Version Compatibility

| Package | Version | Compatible With | Notes |
|---------|---------|-----------------|-------|
| `charm.land/lipgloss/v2` | v2.0.2 | `charm.land/bubbletea/v2` v2.0.2 | Must use matching v2s together; v1 Lip Gloss + v2 Bubble Tea will have I/O conflicts |
| `charm.land/lipgloss/v2` | v2.0.2 | `github.com/alecthomas/chroma/v2` v2.23.1 | No direct version coupling; Chroma outputs `io.Writer`-compatible ANSI strings that Lip Gloss renders correctly |
| `github.com/alecthomas/chroma/v2` | v2.23.1 | Go 1.18+ | v2 requires Go modules; no minimum version beyond what's reasonable |
| `github.com/spf13/cobra` | v1.9.1 | Go 1.20+ | Cobra docs state Go 1.20+ for latest; Go 1.21 comfortably supported |
| `pgregory.net/rapid` | v1.2.0 | Go 1.18+ (generics) | Rapid v1 uses generics; requires Go 1.18+; no conflicts with other test deps |

---

## Confidence Assessment

| Area | Level | Reason |
|------|-------|--------|
| Chroma v2 version (v2.23.1) | HIGH | Verified on pkg.go.dev and GitHub releases page |
| Lip Gloss v2 version (v2.0.2) | HIGH | Stable released Feb 24 2026; patch v2.0.2 released Mar 11 2026; confirmed on pkg.go.dev |
| Bubble Tea v2 version (v2.0.2) | HIGH | Confirmed stable release Feb 24 2026; patch v2.0.2 Mar 6 2026; v2 powers Charm's own products in production |
| Charm vanity import paths (`charm.land/...`) | HIGH | Official Charm blog post + pkg.go.dev confirms new canonical import paths for v2 |
| Cobra v1.9.1 | HIGH | Confirmed via Context7 + cobra.dev docs |
| Algorithm implementation approach | MEDIUM | Industry consensus (study reference impls, don't depend on them) is well-supported by evidence; specific implementation choices (e.g., histogram cutoff thresholds) will need tuning |
| `znkr.io/diff` as reference | HIGH | v1.0.0 released Mar 15 2026; Apache-2.0; active repo; includes benchmarks vs all major Go diff libs |
| rapid v1.2.0 | HIGH | Confirmed on pkg.go.dev; stable since v1.0 (2023) |
| goldie v2.8.0 | HIGH | Confirmed on pkg.go.dev |
| testscript (go-internal) | HIGH | Confirmed; used by Go team itself |

---

## Sources

- `github.com/alecthomas/chroma` releases page → v2.23.1 confirmed (Jan 23 2026)
- `pkg.go.dev/github.com/alecthomas/chroma/v2` → import path, formatter names, API patterns
- `pkg.go.dev/charm.land/lipgloss/v2` + Context7 `/charmbracelet/lipgloss` → v2.0.2, `JoinHorizontal`, `HasDarkBackground`, `LightDark` API
- `charm.land/blog/v2` (Feb 23 2026) → Charm official stable v2 announcement; Bubble Tea + Lip Gloss v2 released together
- `pkg.go.dev/charm.land/bubbletea/v2` → v2.0.2, `tea.View` return type, `charm.land` import path
- `pkg.go.dev/github.com/spf13/cobra` + Context7 `/spf13/cobra` → v1.9.1, `RunE` pattern
- `pkg.go.dev/znkr.io/diff` → v1.0.0, Myers + heuristics, benchmark comparison table
- `pkg.go.dev/pgregory.net/rapid` → v1.2.0, property-based testing
- `pkg.go.dev/github.com/sebdah/goldie/v2` → v2.8.0, golden file testing
- `pkg.go.dev/github.com/rogpeppe/go-internal` → testscript, txtar
- `pkg.go.dev/github.com/charmbracelet/colorprofile` → v0.4.2, terminal color profile detection
- https://go.dev/doc/modules/layout — official Go library + cmd/ module layout guidance
- `github.com/peter-evans/patience` → Patience diff Go reference implementation

---

*Stack research for: drift — Go text diff library and CLI*
*Researched: 2026-03-25*
