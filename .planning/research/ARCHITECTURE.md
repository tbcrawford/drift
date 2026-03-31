# Architecture Research

**Domain:** Go diff library + CLI (terminal text diffing with syntax highlighting)
**Researched:** 2026-03-25
**Confidence:** HIGH

---

## Standard Architecture

### System Overview

```
┌──────────────────────────────────────────────────────────────────┐
│                       Public API Layer                           │
│  ┌─────────────────────┐   ┌──────────────────────────────────┐  │
│  │  Functional API     │   │  Builder API                     │  │
│  │  drift.Diff(a, b,   │   │  drift.New().                    │  │
│  │    ...Option)       │   │    Algorithm(Myers).             │  │
│  └──────────┬──────────┘   │    Lang("go").Unified()          │  │
│             │              └──────────────┬───────────────────┘  │
└─────────────┼────────────────────────────┼──────────────────────┘
              │ Options struct              │ Options struct
              ▼                            ▼
┌──────────────────────────────────────────────────────────────────┐
│                        Core Engine Layer                         │
│  ┌────────────────────────────────────────────────────────────┐  │
│  │               Algorithm Dispatcher                         │  │
│  │  selects Myers | Patience | Histogram via options.Algo     │  │
│  └──────────────────────┬─────────────────────────────────────┘  │
│                         │                                        │
│  ┌──────────┐  ┌─────────┴──┐  ┌───────────┐                    │
│  │  Myers   │  │  Patience  │  │ Histogram │  (algorithm impls) │
│  └──────────┘  └────────────┘  └───────────┘                    │
│                         │ []Edit (op, lineA, lineB)              │
│  ┌──────────────────────▼─────────────────────────────────────┐  │
│  │                  Hunk Builder                              │  │
│  │  groups edits into Hunks with context window               │  │
│  └──────────────────────┬─────────────────────────────────────┘  │
└─────────────────────────┼────────────────────────────────────────┘
                          │ DiffResult{Hunks, Meta}
                          ▼
┌──────────────────────────────────────────────────────────────────┐
│                     Highlighting Layer                           │
│  ┌────────────────────────────────────────────────────────────┐  │
│  │                  Highlighter                               │  │
│  │  Chroma lexer → token iterator → terminal256/true-color   │  │
│  │  formatter → ANSI string per line                         │  │
│  └────────────────────────────────────────────────────────────┘  │
│  ┌────────────────────────────────────────────────────────────┐  │
│  │                  Theme Detector                            │  │
│  │  $COLORFGBG / OSC 11 query → light|dark → Chroma style    │  │
│  └────────────────────────────────────────────────────────────┘  │
└─────────────────────────┬────────────────────────────────────────┘
                          │ []HighlightedLine (ANSI strings + op)
                          ▼
┌──────────────────────────────────────────────────────────────────┐
│                      Renderer Layer                              │
│  ┌─────────────────────────┐  ┌─────────────────────────────┐    │
│  │   UnifiedRenderer       │  │   SplitRenderer             │    │
│  │   writes unified diff   │  │   lipgloss.JoinHorizontal   │    │
│  │   hunk headers + lines  │  │   left | right panels       │    │
│  └──────────────┬──────────┘  └──────────────┬──────────────┘    │
└─────────────────┼──────────────────────────────┼──────────────────┘
                  │                              │
                  └──────────────┬───────────────┘
                                 ▼ io.Writer (os.Stdout / strings.Builder)
┌──────────────────────────────────────────────────────────────────┐
│                          CLI Layer                               │
│  cmd/drift/main.go  →  cobra commands  →  flag parsing          │
│  input: two files | stdin | two strings                         │
│  output: ANSI terminal text                                     │
└──────────────────────────────────────────────────────────────────┘
```

### Component Responsibilities

| Component | Responsibility | Implementation Notes |
|-----------|---------------|----------------------|
| Public API (`drift.go`) | Entry points: `Diff()`, `DiffFiles()`, `New()` | Root package; exports types + functional options |
| Options / Config | Carry algorithm, language, theme, context-lines, output mode | Unexported `config` struct; `Option` = `func(*config)` |
| Algorithm Dispatcher | Select and invoke the right diff algorithm | Interface `algo.Differ`; switch on `config.Algorithm` |
| Myers implementation | Greedy SES algorithm; default | `internal/algo/myers/` |
| Patience implementation | Unique-anchor LCS; better for code refactors | `internal/algo/patience/` |
| Histogram implementation | Frequency-aware; Git's preferred | `internal/algo/histogram/` |
| Hunk Builder | Group raw `[]Edit` into `[]Hunk` with context | `internal/hunk/` |
| Data Model (`DiffResult`) | Canonical diff representation (hunks, lines, ops) | Exported types in root package |
| Highlighter | Run Chroma lexer+formatter per line; return ANSI strings | `internal/highlight/` |
| Theme Detector | Query terminal background; map to Chroma style name | `internal/theme/` |
| UnifiedRenderer | Emit `@@ -a,b +c,d @@` headers + colored lines | `internal/render/unified.go` |
| SplitRenderer | Two-panel layout via `lipgloss.JoinHorizontal` | `internal/render/split.go` |
| CLI (`cmd/drift/`) | Parse flags, read files/stdin, call library, write to stdout | `cmd/drift/main.go` + cobra |

---

## Recommended Project Structure

```
drift/
├── go.mod                    # module: github.com/tbcrawford/drift
├── go.sum
├── drift.go                  # package drift — public API surface
├── options.go                # Option type + WithAlgorithm, WithLang, WithTheme, etc.
├── types.go                  # DiffResult, Hunk, Line, Op (exported types)
├── errors.go                 # sentinel errors
├── doc.go                    # package-level godoc
├── examples/
│   ├── basic/main.go         # functional API example
│   └── builder/main.go       # builder API example
├── internal/
│   ├── algo/
│   │   ├── algo.go           # Differ interface + Algorithm enum
│   │   ├── myers/
│   │   │   └── myers.go
│   │   ├── patience/
│   │   │   └── patience.go
│   │   └── histogram/
│   │       └── histogram.go
│   ├── hunk/
│   │   └── hunk.go           # Edit → Hunk grouping with context window
│   ├── highlight/
│   │   ├── highlighter.go    # Chroma integration: lexer+formatter pipeline
│   │   └── highlight_test.go
│   ├── render/
│   │   ├── renderer.go       # Renderer interface
│   │   ├── unified.go        # UnifiedRenderer
│   │   └── split.go          # SplitRenderer (uses lipgloss)
│   └── theme/
│       └── detect.go         # terminal theme detection
└── cmd/
    └── drift/
        └── main.go           # CLI entry point (thin: parse → call library)
```

### Structure Rationale

- **Root package (`package drift`):** Library importers get a clean `import "github.com/tbcrawford/drift"` with no sub-path required. Public types (`DiffResult`, `Hunk`, `Line`, `Op`) and the two API styles live here. Mirrors Go community best practice for library+CLI repos (Cobra, Viper, goldmark all use this pattern).
- **`internal/`:** All implementation details. Compiler-enforced; third-party consumers can't import internals. Algorithms, hunk building, highlighting, rendering are all internal.
- **`cmd/drift/`:** Thin CLI wrapper. `main.go` should stay under ~50 lines — parse flags, call library, handle errors, exit. No business logic here.
- **`examples/`:** Runnable examples for godoc and onboarding. Shows both functional + builder styles.
- **No `pkg/` directory:** Only needed when you have two different binaries sharing internal library code. Single-module single-library doesn't need the extra indirection.

---

## Architectural Patterns

### Pattern 1: Functional Options (Primary API Style)

**What:** `Option = func(*config)` functions accepted as variadic args by top-level functions. Rob Pike pattern, widely used in Go stdlib and major libraries.
**When to use:** All configuration that has sensible defaults — algorithm, language, theme, context lines, output mode.
**Trade-offs:** Slightly more boilerplate for authors; excellent DX for users. IDE autocomplete works naturally with `With`-prefixed option funcs.

**Example:**
```go
// Public API surface in drift.go
type Option func(*config)

func WithAlgorithm(a Algorithm) Option {
    return func(c *config) { c.algorithm = a }
}

func WithLang(lang string) Option {
    return func(c *config) { c.lang = lang }
}

// One-liner functional API
func Diff(a, b string, opts ...Option) (DiffResult, error) {
    cfg := defaultConfig()
    for _, o := range opts {
        o(cfg)
    }
    return run(a, b, cfg)
}
```

### Pattern 2: Builder / Fluent API (Secondary API Style)

**What:** A `Differ` struct with chainable setter methods, terminated by `Unified()` or `Split()`.
**When to use:** When callers want to configure once and call multiple times, or when the call site benefits from reading like a sentence.
**Trade-offs:** More code to maintain; solves a real use case (reuse configured differ). Both styles should delegate to the same internal `config`.

**Example:**
```go
// Builder in drift.go
type Differ struct { cfg *config }

func New(opts ...Option) *Differ {
    cfg := defaultConfig()
    for _, o := range opts {
        o(cfg)
    }
    return &Differ{cfg: cfg}
}

func (d *Differ) Algorithm(a Algorithm) *Differ {
    d.cfg.algorithm = a
    return d
}

func (d *Differ) Unified(a, b string) (string, error) { ... }
func (d *Differ) Split(a, b string) (string, error)   { ... }
```

### Pattern 3: Strategy Interface for Algorithms

**What:** A narrow `internal/algo.Differ` interface. Each algorithm (Myers, Patience, Histogram) implements it. Dispatcher selects based on `config.Algorithm`.
**When to use:** Whenever behavior needs to be swappable without changing callers.
**Trade-offs:** Slight abstraction overhead; buys clean unit-testability per algorithm and easy addition of new algorithms later.

**Example:**
```go
// internal/algo/algo.go
type Algorithm int
const (
    Myers Algorithm = iota
    Patience
    Histogram
)

type Differ interface {
    Diff(a, b []string) ([]Edit, error)
}

type Edit struct {
    Op   Op
    Line string
    // LineA, LineB int (original line numbers)
}

type Op int
const (
    OpMatch Op = iota
    OpInsert
    OpDelete
)

func New(a Algorithm) Differ {
    switch a {
    case Patience:  return patience.New()
    case Histogram: return histogram.New()
    default:        return myers.New()
    }
}
```

### Pattern 4: Chroma as a Direct Dependency, Not a Wrapper

**What:** Call Chroma's lexer → tokenize → format pipeline directly inside `internal/highlight/`. Do NOT wrap Chroma behind a custom interface unless you have a concrete reason to swap it.
**When to use:** When a library is an established, stable dependency (Chroma v2.23+ is mature).
**Trade-offs:** Tight coupling to Chroma API, but that's acceptable — Chroma IS the highlighting. A wrapper interface adds complexity with no benefit for v1.

**Example:**
```go
// internal/highlight/highlighter.go
func HighlightLine(line string, lang string, style *chroma.Style) (string, error) {
    lexer := lexers.Get(lang)
    if lexer == nil {
        lexer = lexers.Analyse(line)
    }
    if lexer == nil {
        lexer = lexers.Fallback
    }
    lexer = chroma.Coalesce(lexer)

    formatter := formatters.Get("terminal256")
    // returns ANSI-escaped string for this single line
    var buf strings.Builder
    iter, err := lexer.Tokenise(nil, line)
    if err != nil {
        return line, err
    }
    return buf.String(), formatter.Format(&buf, style, iter)
}
```

---

## Data Flow

### Complete Pipeline: String Input → Terminal Output

```
User Input (string a, string b)
    │
    ▼
[Options resolved into config struct]
    │
    ▼
[Algorithm Dispatcher] → selects Myers | Patience | Histogram
    │
    ▼
[Differ.Diff(lines_a, lines_b)] → []Edit{Op, Line}
    │
    ▼
[Hunk Builder] → groups contiguous edits + context lines → []Hunk
    │
    ▼
[DiffResult{Hunks []Hunk, OldFile, NewFile, Meta}]   ← STABLE TYPE
    │
    ├──────────────────────────────┐
    ▼                              ▼
[Highlighter]              (DiffResult returned to caller as-is
  per-line: Chroma tokenize      if they call Diff() directly
  → ANSI string)                  and want to render themselves)
    │
    ▼
[[]HighlightedLine{ANSI string, Op}]
    │
    ├─────────────────────┬──────────────────────┐
    ▼                     ▼                      ▼
[UnifiedRenderer]  [SplitRenderer]        [NoColor fallback]
  io.Writer          lipgloss panels        plain text
    │                     │
    └──────────┬───────────┘
               ▼
         io.Writer (os.Stdout in CLI, any writer in library)
```

### Key Data Flows

1. **Algo → Hunk:** `[]Edit` is a flat list. Hunk Builder groups contiguous non-match ops with N context lines on either side, producing `[]Hunk`. This separation lets the algorithm stay pure (no rendering knowledge) and the hunk builder be tested independently.

2. **Highlighting layered over diff, not before:** Chroma runs on already-split lines *after* the diff is computed. It highlights each line's content independently using `strings.Builder` → ANSI string. Diff color (red/green background via Lip Gloss) is applied *after* Chroma syntax color, layered on top as a background tint. This avoids Chroma fighting with diff colors.

3. **DiffResult as the pivot type:** The exported `DiffResult` is what the library returns. Renderers consume it. If a user wants to render custom output, they call `Diff()` and process `DiffResult` themselves. This makes the library useful as a pure diff engine independent of our renderers.

---

## Component Boundaries

| Boundary | Direction | Communication | Rule |
|----------|-----------|---------------|------|
| Public API → internal/algo | outbound | `[]Edit` return value | algo never imports anything from root |
| internal/algo → internal/hunk | outbound | `[]Edit` | hunk has no knowledge of algorithms |
| internal/hunk → root types | inbound | returns `[]Hunk` using exported types | hunks use exported `Hunk`, `Line`, `Op` |
| internal/highlight → root types | inbound | receives lines from `DiffResult` | no highlight knowledge in diff engine |
| internal/render → internal/highlight | inbound | renderer calls highlighter, not vice versa | renderers own the pipeline |
| internal/render → lipgloss | outbound | direct import; lipgloss is a layout primitive | no abstraction needed |
| cmd/drift/ → root package | inbound | thin CLI wrapper; imports `drift` like any user | CLI cannot import `internal/` |

---

## Build Order (Component Dependencies)

Build components in this order — each depends only on things already built:

```
1. types.go + errors.go          ← no dependencies; define Op, Edit, Hunk, Line, DiffResult
2. options.go (config struct)    ← depends on types
3. internal/algo/algo.go         ← depends on types (Edit, Op)
4. internal/algo/myers/          ← implements algo.Differ
5. internal/algo/patience/       ← implements algo.Differ
6. internal/algo/histogram/      ← implements algo.Differ
7. internal/hunk/                ← depends on algo (Edit) + types (Hunk)
8. internal/theme/               ← standalone; no internal deps
9. internal/highlight/           ← depends on types; imports Chroma directly
10. internal/render/unified.go   ← depends on highlight, types; imports lipgloss
11. internal/render/split.go     ← depends on unified + lipgloss JoinHorizontal
12. drift.go (Diff, DiffFiles)   ← wires all internal layers together
13. drift.go (New / Builder)     ← depends on drift.go (functional API)
14. cmd/drift/main.go            ← imports root package; adds cobra CLI
```

**Why this order matters:**
- Types must exist before anything references them.
- Algorithms are pure computation — no I/O, no rendering deps.
- Hunk builder depends on the Edit type from algo but nothing rendering-specific.
- Highlighting and rendering are the outermost layers; they can import freely from inner layers but inner layers never import back.
- CLI is always last. It is a consumer, not a producer.

---

## Anti-Patterns

### Anti-Pattern 1: Putting Rendering Logic in the Diff Engine

**What people do:** Mix ANSI escape codes or Lip Gloss styles directly into the algorithm or hunk builder.
**Why it's wrong:** Breaks testability of the algorithm (you can't assert on pure line ops anymore), couples diff correctness to render details, and prevents reuse of `DiffResult` by callers who want their own rendering.
**Do this instead:** Algorithms return `[]Edit`, hunk builder returns `[]Hunk` with plain strings. Renderers consume those downstream.

### Anti-Pattern 2: Wrapping Chroma Behind a Custom Interface "For Flexibility"

**What people do:** Define a `Highlighter` interface so you can swap Chroma for something else later.
**Why it's wrong:** Premature abstraction. Chroma is the dominant Go syntax highlighter with 290+ contributors and 2.23.x releases as of early 2026. There is no realistic alternative to swap in. The interface adds complexity with no concrete benefit for v1.
**Do this instead:** Import `github.com/alecthomas/chroma/v2` directly in `internal/highlight/`. If swap-ability is needed later, add the interface then.

### Anti-Pattern 3: Exposing `internal/` Types in the Public API

**What people do:** Return `*internal/algo.Edit` or `*internal/hunk.Hunk` from exported functions.
**Why it's wrong:** Forces internal package reorganizations to become public API breaks. Callers can't import `internal/` anyway, so you'd need awkward type aliases.
**Do this instead:** Define clean exported types in `types.go` at the root package. Internal packages use or convert to those exported types. The exported `Hunk`, `Line`, `Op` are the stable contract.

### Anti-Pattern 4: Highlighting the Entire File Before Diffing

**What people do:** Run Chroma on the whole file, then apply diff markers on top of already-highlighted strings.
**Why it's wrong:** ANSI sequences inside already-highlighted text interfere with diff boundary detection. You can't reliably split ANSI-escaped content at line boundaries.
**Do this instead:** Compute the diff on plain strings, then apply Chroma highlighting per-line on the plain string content. Apply diff background color (Lip Gloss) as a final wrapper. Pipeline: plain diff → plain lines → highlight lines → tint with diff color.

### Anti-Pattern 5: CLI Doing More Than Wire-Up

**What people do:** Put arg parsing, file reading, output logic, and error handling all in `main.go` with no separation.
**Why it's wrong:** Makes the CLI untestable and couples concerns.
**Do this instead:** Keep `cmd/drift/main.go` under 50 lines. It does: `cobra` setup → call `drift.DiffFiles()` or `drift.Diff()` → write result to `os.Stdout`. All real logic lives in the library.

---

## Integration Points

### External Libraries

| Library | Integration Pattern | Where Used | Notes |
|---------|---------------------|------------|-------|
| `github.com/alecthomas/chroma/v2` | Direct dependency in `internal/highlight/` | Per-line syntax tokenization + ANSI formatting | Use `terminal256` or `terminal16m` formatter; detect capability via `$TERM` |
| `github.com/charmbracelet/lipgloss` | Direct dependency in `internal/render/` | Split-view panel layout, diff background colors | Use `JoinHorizontal` for side-by-side; use `lipgloss.Color` for +/- line tints |
| `github.com/spf13/cobra` | CLI only (`cmd/drift/`) | Subcommand structure, flag parsing | Thin wrapper; do not leak cobra into library |

### Internal Boundaries

| Boundary | Communication | Notes |
|----------|---------------|-------|
| `internal/algo` ↔ `internal/hunk` | `[]Edit` slice | Edit is defined in root `types.go` or `internal/algo/algo.go` — decision to make at implementation time |
| `internal/hunk` ↔ `internal/render` | `DiffResult` (exported root type) | Renderer takes the full result |
| `internal/highlight` ↔ `internal/render` | Per-line ANSI strings | Renderer calls highlighter for each line; highlight is stateless |
| Root package ↔ `cmd/drift/` | Exported public API only | CLI is a pure consumer |

---

## Chroma Integration: Two-Layer Coloring Strategy

This is the most subtle architectural decision. Chroma produces syntax-token colors (blue for keywords, yellow for strings, etc.). Diff adds line-level colors (green background for additions, red for deletions). These must coexist.

**The correct layering:**

```
Step 1: Chroma highlights plain source line → ANSI string with syntax colors
Step 2: Lip Gloss wraps the ANSI string with a background color tint
        (add: lipgloss.NewStyle().Background(lipgloss.Color("#1a3a1a")).Render(ansiLine))
        (del: lipgloss.NewStyle().Background(lipgloss.Color("#3a1a1a")).Render(ansiLine))
Step 3: Prefix character (+/-/ ) is prepended by the renderer
```

**Why this works:** Lip Gloss is ANSI-safe — it preserves embedded escape codes when adding background styles. The Chroma foreground colors remain visible against the diff background tint.

**Theme detection flow:**

```
$COLORFGBG set?  → parse "fg;bg", bg < 8 = dark, bg >= 8 = light
    │ no
    ▼
$TERM or $COLORTERM set to "truecolor"?  → query OSC 11 (terminal bg color)
    │ no / unsupported
    ▼
Default to "dark" (safer for most modern terminals)
    │
    ▼
dark  → use Chroma "github-dark", "dracula", or "monokai" style
light → use Chroma "github", "solarized-light", or "tango" style
```

---

## Scaling Considerations

This is a library; "scaling" means handling large inputs and many concurrent callers.

| Scale | Architecture Consideration |
|-------|---------------------------|
| Small files (< 1K lines) | Default Myers with context = 3; no special handling needed |
| Large files (> 10K lines) | Myers heuristic mode (not `--minimal`); hunk streaming via `io.Writer` avoids buffering entire result |
| Concurrent use | All exported functions should be stateless (take options, return results). No global mutable state. Builder's `*Differ` should document whether it's safe for concurrent calls. |
| Binary / non-text | Detect binary content early (null bytes check), return an error or a "binary files differ" placeholder rather than corrupting terminal output |

---

## Sources

- znkr.io/diff — Modern Go diff library architecture showing Edit/Hunk model and functional options pattern for algorithm selection: https://pkg.go.dev/znkr.io/diff (HIGH confidence — official docs)
- go-git diff package — Chunk/Operation/Patch/UnifiedEncoder type model showing clean separation of data model from encoder: https://pkg.go.dev/github.com/go-git/go-git/v5/plumbing/format/diff (HIGH confidence — official docs)
- Chroma v2 — Lexer/Formatter/Style pipeline, terminal formatters (terminal256, terminal16m): https://github.com/alecthomas/chroma (HIGH confidence — official repo)
- Lip Gloss v1.1 — JoinHorizontal for split panels, ANSI-safe background styling: https://pkg.go.dev/github.com/charmbracelet/lipgloss (HIGH confidence — official docs)
- Go project layout best practices 2026 — Library+CLI single-module layout (root package = library, cmd/ = binary): https://alnah.io/post/go-project-layout/ (MEDIUM confidence — community, corroborated by multiple sources)
- Functional options pattern — Rob Pike pattern, 10-year review: https://www.bytesizego.com/blog/10-years-functional-options-golang (HIGH confidence — well-established pattern)
- jansmrcka/differ — Real-world Go diff TUI using Chroma + Lipgloss + Bubbletea as reference architecture: https://pkg.go.dev/github.com/jansmrcka/differ (MEDIUM confidence — community project)
- Histogram diff algorithm explained — How jgit histogram works and when it falls back to Myers: https://raygard.github.io/2025/01/29/a-histogram-diff-implementation/ (HIGH confidence — primary source analysis)

---

*Architecture research for: drift — Go diff library + CLI*
*Researched: 2026-03-25*
