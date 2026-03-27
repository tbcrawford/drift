# Phase 3: Unified Rendering — Research

**Phase Goal:** `drift.Diff()` produces richly highlighted unified diff output renderable to any `io.Writer`
**Requirements:** REND-01, REND-03, REND-04, REND-05, REND-06, REND-07, REND-08, REND-09
**Research Date:** 2026-03-25

---

## Executive Summary

**Key decisions:**

1. **`HasDarkBackground` has a real 2-second timeout** — verified from `terminal.go` source. It sends OSC 11 + DA1 query sequences to stdin/stdout in raw mode and uses a cancel reader with `time.After(2s)`. In non-TTY environments `term.MakeRaw` fails, the function returns `err != nil`, and it defaults to `true` (dark). **Safe to call without hanging as long as non-TTY is guarded with `colorprofile.Detect` first.**

2. **Chroma formatter selection maps directly to `colorprofile`** — `TrueColor` → `formatters.TTY16m` (`terminal16m`), `ANSI256` → `formatters.TTY256` (`terminal256`), `ANSI` → `formatters.TTY16` (`terminal16`), `NoTTY`/`Ascii` → `formatters.NoOp` (`noop`). This is a one-to-one map, no guessing required.

3. **`colorprofile.NewWriter` handles all ANSI downsampling automatically** — writing TrueColor ANSI through a `colorprofile.Writer` with the detected profile downgrades to the correct depth transparently. Drift should generate TrueColor output and wrap the output writer in `colorprofile.NewWriter` for automatic degradation.

4. **`NO_COLOR` is detected by `colorprofile.Detect`** — if `NO_COLOR` is set in the environment, `colorprofile.Detect(os.Stdout, os.Environ())` returns `colorprofile.Ascii`, which maps to `formatters.NoOp`. No separate `NO_COLOR` check needed beyond using `colorprofile`.

5. **Unified diff format is fully covered by the existing `Hunk` type** — `Hunk.OldStart`, `Hunk.OldLines`, `Hunk.NewStart`, `Hunk.NewLines` directly produce `@@ -OldStart,OldLines +NewStart,NewLines @@`. The renderer just needs to format this header and prefix each `Line` with `+`, `-`, or ` `.

6. **Phase 3 does NOT need to change `drift.Diff()` signature** — rendering is additive. Add a `drift.Render(result DiffResult, w io.Writer, opts ...Option) error` function (or `DiffResult.WriteTo(w io.Writer, opts ...Option) error`). The `config` struct already has `noColor`, `lang`, and `theme` fields from Phase 1.

7. **Theme selection strategy** — dark terminal defaults: `monokai`. Light terminal defaults: `github`. Both are well-tested Chroma styles with complete token coverage and ANSI terminal output support.

---

## Research Area 1: Lip Gloss v2 `HasDarkBackground` — Timeout and Non-TTY Behavior

### API Signature (verified from source)

```go
// charm.land/lipgloss/v2
func HasDarkBackground(in term.File, out term.File) bool
func BackgroundColor(in term.File, out term.File) (bg color.Color, err error)
```

`term.File` is `interface { Fd() uintptr; Read([]byte) (int, error) }` — satisfied by `*os.File`.

### Mechanism (verified from `terminal.go` in `charm.land/lipgloss/v2@v2.0.1`)

1. Calls `term.MakeRaw(in.Fd())` to put the terminal in raw mode.
2. Sends `\033]11;?\007` (OSC 11 — query background color) + `\033[c` (DA1 — primary device attributes) to `out`.
3. Spawns a goroutine with `time.After(2 * time.Second)` that cancels the reader if no response arrives.
4. Reads responses in a loop; OSC 11 response sets the color; DA1 response (`?...c`) is the sentinel that signals "done reading".
5. Returns the parsed color, or `nil, err` on timeout/failure.

**`HasDarkBackground` defaults to `true` on error.** The godoc comment states: "By default, this function will return true if it encounters an error."

### Non-TTY / Piped Behavior

When `stdout` is piped (e.g., `drift a.go b.go | cat`):
- `term.MakeRaw(in.Fd())` returns an error immediately on non-TTY file descriptors.
- `BackgroundColor` returns `nil, err`.
- `HasDarkBackground` returns `true` (dark default).

**Critical implication:** In piped environments, `HasDarkBackground` would incorrectly return `true` AND colors would be written to the pipe. The correct guard is: **check `colorprofile.Detect` first**. If the profile is `NoTTY`, skip theme detection entirely and use no-color output.

### Safe Usage Pattern

```go
// internal/theme/theme.go
import (
    "os"
    "github.com/charmbracelet/colorprofile"
    "charm.land/lipgloss/v2"
)

func DetectDarkBackground(profile colorprofile.Profile) bool {
    // If not a TTY or no color, don't attempt OSC 11 query.
    // HasDarkBackground would return true anyway, but we avoid
    // the 2-second timeout path entirely.
    if profile == colorprofile.NoTTY || profile == colorprofile.Ascii {
        return true // default; won't be used since colors are off
    }
    return lipgloss.HasDarkBackground(os.Stdin, os.Stdout)
}
```

This pattern guarantees:
- No 2-second hang in piped/non-TTY environments (TTY check short-circuits).
- Correct dark/light detection in interactive terminals.
- Correct `true` default for environments where detection is impossible.

### Timeout Risk Mitigation

The 2-second timeout is only reached if stdin is a TTY but the terminal does not respond to OSC 11 (rare: dumb terminals, some CI systems with pseudo-TTYs). `colorprofile.Detect` uses environment variables (`TERM`, `COLORTERM`, `NO_COLOR`) before checking TTY, so CI environments with `TERM=dumb` or `NO_COLOR=1` will return `Ascii` or `NoTTY` and skip the OSC 11 path entirely.

---

## Research Area 2: Chroma v2 API — Terminal Highlighting

### Import Path

```go
import (
    "github.com/alecthomas/chroma/v2"
    "github.com/alecthomas/chroma/v2/lexers"
    "github.com/alecthomas/chroma/v2/formatters"
    "github.com/alecthomas/chroma/v2/styles"
)
```

### Complete Highlighting Pipeline

```go
// 1. Identify the lexer
lexer := lexers.Get("go")          // by explicit language name
// OR
lexer := lexers.Match("foo.go")    // by filename/extension
// OR
lexer := lexers.Analyse(content)   // by content analysis

if lexer == nil {
    lexer = lexers.Fallback        // plaintext fallback
}
lexer = chroma.Coalesce(lexer)     // merge adjacent same-type tokens (recommended)

// 2. Select style (theme)
style := styles.Get("monokai")
if style == nil {
    style = styles.Fallback        // swapoff
}

// 3. Select formatter based on color depth
formatter := formatters.Get("terminal16m") // TrueColor
// or "terminal256" / "terminal16" / "terminal" / "noop"
if formatter == nil {
    formatter = formatters.Fallback        // NoOp
}

// 4. Tokenize and format
iterator, err := lexer.Tokenise(nil, content)
if err != nil { ... }

var buf bytes.Buffer
err = formatter.Format(&buf, style, iterator)
```

### Formatter Registry Names (verified from `formatters/` source)

| Variable | Registered Name | Color Depth | Use When |
|----------|-----------------|-------------|----------|
| `formatters.TTY16m` | `"terminal16m"` | 24-bit TrueColor | `colorprofile.TrueColor` |
| `formatters.TTY256` | `"terminal256"` | 8-bit 256-color | `colorprofile.ANSI256` |
| `formatters.TTY16` | `"terminal16"` | 16 named ANSI colors | `colorprofile.ANSI` |
| `formatters.TTY8` | `"terminal8"` | 8 basic ANSI colors | (rarely needed) |
| `formatters.TTY` | `"terminal"` | 8 basic ANSI colors | Fallback for ANSI |
| `formatters.NoOp` | `"noop"` | No color (plain text) | `colorprofile.NoTTY` / `colorprofile.Ascii` / `NO_COLOR` |

**Important:** `formatters.Get(name)` returns `formatters.Fallback` (which is `NoOp`) for unknown names — so an invalid theme name silently produces plain text rather than panicking.

### Key Formatter Behavior

- **`TTY16m` (TrueColor):** Emits `\033[38;2;R;G;Bm` (foreground) and `\033[48;2;R;G;Bm` (background) ANSI sequences. Resets at line boundaries so each line is self-contained (safe for piped use).
- **`TTY256`:** Maps TrueColor Chroma colors to nearest ANSI 256 palette using Lab color distance. Emits `\033[38;5;Nm` sequences.
- **`TTY16`:** Maps to 16 ANSI named colors. Uses `\033[30m`–`\033[37m` and `\033[90m`–`\033[97m`.
- **`NoOp`:** Strips all formatting, outputs token text only — pure plain text.

### Style (Theme) Behavior

`styles.Get(name)` returns `nil` for unknown names; always check and fall back to `styles.Fallback`.

**Style names are case-insensitive** (from Chroma README): `"monokai"` and `"Monokai"` both work.

**Available styles** (verified from source — `styles/*.xml`):

Dark themes: `monokai`, `dracula`, `github-dark`, `gruvbox`, `hrdark`, `onedark`, `tokyonight-night`, `tokyonight-storm`, `tokyonight-moon`, `nord`, `catppuccin-frappe`, `catppuccin-macchiato`, `catppuccin-mocha`, `solarized-dark`, `solarized-dark256`, `rose-pine`, `rose-pine-moon`, `vulcan`, `witchhazel`, `doom-one`

Light themes: `github`, `xcode`, `gruvbox-light`, `solarized-light`, `catppuccin-latte`, `rose-pine-dawn`, `tokyonight-day`, `modus-operandi`, `manni`, `autumn`, `friendly`, `pastie`, `pygments`, `tango`, `emacs`, `colorful`, `borland`

**Recommended defaults:**
- Dark: `"monokai"` — widely recognized, good terminal ANSI coverage, comprehensive token types
- Light: `"github"` — matches GitHub's diff UI which is the reference for drift's value prop

### `chroma.Coalesce` — Why It Matters

Without `Coalesce`, a single `func` keyword might produce multiple `Keyword` tokens from the lexer's state machine. `Coalesce` merges adjacent same-type tokens. This is important for per-line syntax highlighting because it reduces the number of ANSI escape sequences per line (cleaner output, faster rendering).

### Per-Line Highlighting Pattern

For unified diff rendering, we highlight each line independently (not the whole file):

```go
func HighlightLine(line string, lexer chroma.Lexer, style *chroma.Style, formatter chroma.Formatter) (string, error) {
    iterator, err := lexer.Tokenise(nil, line)
    if err != nil {
        return line, err // fallback to plain
    }
    var buf bytes.Buffer
    if err := formatter.Format(&buf, style, iterator); err != nil {
        return line, err
    }
    return buf.String(), nil
}
```

**Caveat:** Per-line tokenization loses multi-line context (e.g., a string spanning lines). For v1, this is acceptable — unified diff output shows individual lines anyway. The tradeoff is documented.

---

## Research Area 3: `colorprofile` Detection

### Core API

```go
import "github.com/charmbracelet/colorprofile"

// Detect from TTY status + environment variables
p := colorprofile.Detect(os.Stdout, os.Environ())

// Detect from environment only (no TTY check)
p := colorprofile.Env(os.Environ())
```

### Profile Values

```go
colorprofile.TrueColor // 24-bit; COLORTERM=truecolor or COLORTERM=24bit
colorprofile.ANSI256   // 8-bit; TERM=xterm-256color (and no truecolor env)
colorprofile.ANSI      // 4-bit; TERM=xterm or TERM=screen (no 256 support)
colorprofile.Ascii     // 1-bit bold/italic only, no color; TERM=dumb or NO_COLOR=1
colorprofile.NoTTY     // Not a terminal; stdout is piped/redirected
```

### `NO_COLOR` Compliance

When `NO_COLOR` is set to any non-empty value, `colorprofile.Detect` / `colorprofile.Env` returns `colorprofile.Ascii`. Mapping `Ascii` → `formatters.NoOp` automatically satisfies the NO_COLOR standard (no-color.org). **No separate `os.Getenv("NO_COLOR")` check is needed.**

### `colorprofile.NewWriter` — Automatic Downsampling

```go
// Wrap any io.Writer for automatic ANSI downsampling
w := colorprofile.NewWriter(os.Stdout, os.Environ())
// Now write TrueColor ANSI to w — it gets downsampled to terminal capability
fmt.Fprintln(w, chromaHighlightedLine)
```

`colorprofile.Writer` is an `io.Writer` that transparently rewrites ANSI escape sequences to the target profile. Writing TrueColor sequences (`\033[38;2;R;G;Bm`) to a 256-color terminal gets rewritten to `\033[38;5;Nm` automatically.

**Decision for drift:** Always generate output using the `terminal16m` (TrueColor) formatter, then wrap the caller's `io.Writer` in a `colorprofile.NewWriter`. This way drift always produces the highest-quality output and downsampling happens transparently at the output layer. This simplifies the highlight pipeline — one formatter, one style, automatic adaptation.

### Profile → Formatter Mapping (simplified)

```go
func formatterForProfile(p colorprofile.Profile) chroma.Formatter {
    switch p {
    case colorprofile.TrueColor:
        return formatters.TTY16m
    case colorprofile.ANSI256:
        return formatters.TTY256
    case colorprofile.ANSI:
        return formatters.TTY16
    default: // Ascii, NoTTY
        return formatters.NoOp
    }
}
```

---

## Research Area 4: Unified Diff Format

### Exact Format (Git-compatible)

```
--- a/path/to/old.go
+++ b/path/to/new.go
@@ -OldStart,OldLines +NewStart,NewLines @@ [optional context label]
 context line (unchanged)
-deleted line
+added line
 context line (unchanged)
```

### Hunk Header Rules

- `@@ -a,b +c,d @@` where:
  - `a` = `Hunk.OldStart` (1-indexed line number in old file where hunk begins)
  - `b` = `Hunk.OldLines` (total old file lines in hunk: Equal + Delete)
  - `c` = `Hunk.NewStart` (1-indexed line number in new file where hunk begins)
  - `d` = `Hunk.NewLines` (total new file lines in hunk: Equal + Insert)
- **Special case:** When `b == 0` (all-insert hunk), git writes `@@ -a,0 +c,d @@` with `a` being the line *before* the insertion. Our hunk builder already handles this with `oldStart = 1` edge case.
- **Special case:** When `b == 1`, git omits the `,1` (writes `@@ -a +c,d @@`). However, for simplicity in v1, always emit `,b` and `,d` — tools like `patch` accept both forms.

### Line Prefix Rules

| `Line.Op` | Prefix | Color (terminal) |
|-----------|--------|-----------------|
| `Equal`   | ` ` (space) | No color / dim gray |
| `Insert`  | `+` | Green (bright) |
| `Delete`  | `-` | Red (bright) |

### Existing `Hunk` Type Already Encodes This

The `Hunk` struct from Phase 1 has exactly the fields needed:

```go
type Hunk struct {
    OldStart int    // → a in @@ -a,b +c,d @@
    OldLines int    // → b
    NewStart int    // → c
    NewLines int    // → d
    Lines    []Line // each Line has Op + Content
}
```

The renderer only needs to format this data — no algorithm changes needed.

### File Header Convention

Git's unified diff includes `--- a/file` / `+++ b/file` headers. For the library API where the caller passes raw strings (no file paths), these headers are optional or can be omitted. The renderer should accept optional `oldName`/`newName` parameters, defaulting to `"a/input"` / `"b/input"` when not provided.

---

## Research Area 5: `diff.Edit` → Unified Diff Rendering Pipeline

**Phase 3 does NOT need to rebuild the `Edit` → `Hunk` pipeline.** The existing `hunk.Build()` function already:
1. Consumes `[]edittype.Edit` from any algorithm.
2. Groups edits into context windows.
3. Computes `OldStart`, `OldLines`, `NewStart`, `NewLines`.
4. Returns `[]Hunk` with `[]Line` for each hunk.

The `DiffResult.Hunks []Hunk` coming out of `drift.Diff()` is already in the exact shape the renderer needs. **Phase 3's rendering pipeline begins at `DiffResult`, not at `[]Edit`.**

```
drift.Diff(old, new, opts...) → DiffResult
    ↓
DiffResult.Hunks → []Hunk
    ↓
drift.Render(result, w, opts...) → io.Writer  ← Phase 3 adds this
    ↓
For each Hunk:
  Write "@@ -OldStart,OldLines +NewStart,NewLines @@\n"
  For each Line:
    highlighted = HighlightLine(Line.Content, lexer, style, formatter)
    Write prefix + highlighted + "\n"
```

---

## Research Area 6: Color Theme Selection

### Theme Recommendation Table

| Terminal | Default Theme | Rationale |
|----------|---------------|-----------|
| Dark background | `"monokai"` | Iconic dark theme; comprehensive coverage; all major token types; well-tested in terminal formatters |
| Light background | `"github"` | Matches GitHub's diff UI (drift's stated reference); designed for white backgrounds; good contrast |

### Fallback Chain

```go
func selectTheme(requested string, isDark bool) *chroma.Style {
    if requested != "" {
        if s := styles.Get(requested); s != nil {
            return s
        }
        // Unknown theme name: fall through to auto-detect
    }
    name := "monokai"
    if !isDark {
        name = "github"
    }
    if s := styles.Get(name); s != nil {
        return s
    }
    return styles.Fallback // swapoff — always present
}
```

### `styles.Fallback`

`styles.Fallback` is `styles.Get("swapoff")`, which is always registered. It provides a neutral dark theme suitable for any terminal background. It is the safe last-resort fallback.

### Style Name Is Case-Insensitive

`styles.Get("Monokai")` and `styles.Get("monokai")` both work. The registry normalizes names to lowercase internally.

---

## Research Area 7: Existing Codebase Integration Points

### What Exists (verified by reading source)

**`config` struct (options.go):** Already has all needed fields:
```go
type config struct {
    algorithm    Algorithm
    contextLines int
    noColor      bool   // WithNoColor() sets this
    lang         string // WithLang("go") sets this
    theme        string // WithTheme("monokai") sets this
}
```

**`DiffResult` / `Hunk` / `Line` types (edittype/edittype.go):** Fully defined:
- `Hunk.OldStart`, `Hunk.OldLines`, `Hunk.NewStart`, `Hunk.NewLines` — ready for `@@` header
- `Line.Op` (Equal/Insert/Delete) + `Line.Content` (line text without newline) — ready for `+`/`-`/` ` prefix

**`drift.Diff()` signature (drift.go):**
```go
func Diff(old, new string, opts ...Option) (DiffResult, error)
```
Phase 3 adds a new function, not a change to `Diff`.

### New Functions/Packages Phase 3 Adds

```
internal/theme/
  theme.go         ← DetectDarkBackground() with safe TTY guard
  theme_test.go

internal/highlight/
  highlight.go     ← HighlightLine(), formatterForProfile(), selectTheme()
  highlight_test.go

internal/render/
  unified.go       ← UnifiedRenderer.Render(result DiffResult, w io.Writer, cfg *config)
  unified_test.go

render.go          ← drift.Render(result DiffResult, w io.Writer, opts ...Option) error (public API)
```

### New Public API

```go
// render.go — new top-level function
func Render(result DiffResult, w io.Writer, opts ...Option) error {
    cfg := defaultConfig()
    for _, o := range opts {
        o(cfg)
    }
    // Detect color profile from writer (if *os.File) or default to NoTTY
    // Detect dark background (guarded by profile)
    // Select lexer, style, formatter
    // Render via internal/render/unified
    return render.Unified(result, w, cfg)
}
```

**Design consideration:** `Render` accepts a `DiffResult` (not raw strings) — the caller must call `Diff` first. This is intentional: separates diffing from rendering, allows callers to inspect the structured result before rendering.

**Alternative:** Add rendering to `Diff` directly via options. Rejected — it conflates diffing (pure function) with rendering (side-effectful, I/O). The two-step pattern matches how users want to use the library (inspect hunks, then optionally render).

### Import Graph (no new cycles)

```
drift (root) → internal/render/unified → internal/highlight → internal/theme
                                       → internal/edittype (Hunk/Line types)
```

No new import cycles — `internal/theme` and `internal/highlight` are leaf packages.

---

## Research Area 8: `NO_COLOR` Compliance

### Standard (no-color.org)

> When set, callers should not add ANSI color escape codes to output. The `NO_COLOR` variable should have no effect on formatting that is explicitly requested by the user (e.g., `--color`).

### Detection Chain

In drift's case:

1. `colorprofile.Detect(w, os.Environ())` — returns `colorprofile.Ascii` when `NO_COLOR` is set.
2. `Ascii` profile maps to `formatters.NoOp`.
3. `formatters.NoOp` writes token text only — no ANSI codes.
4. Result: plain unified diff with `+`/`-`/` ` prefixes, no color.

**Explicit `WithNoColor()` option:** When `cfg.noColor == true`, override the detected profile to `colorprofile.Ascii` regardless of environment. This ensures `drift.Render(result, os.Stdout, drift.WithNoColor())` always produces plain output even on a TrueColor TTY.

```go
func resolveProfile(w io.Writer, cfg *config) colorprofile.Profile {
    if cfg.noColor || os.Getenv("NO_COLOR") != "" {
        return colorprofile.Ascii
    }
    if f, ok := w.(*os.File); ok {
        return colorprofile.Detect(f, os.Environ())
    }
    return colorprofile.NoTTY // non-file writers treated as piped
}
```

Note: `colorprofile.Detect` already checks `NO_COLOR` in `os.Environ()`, so the `os.Getenv("NO_COLOR")` check in `resolveProfile` is belt-and-suspenders but documents the intent clearly.

---

## Research Area 9: `go-isatty` Integration

### API

```go
import "github.com/mattn/go-isatty"

isatty.IsTerminal(os.Stdout.Fd())       // true if stdout is a real TTY
isatty.IsCygwinTerminal(os.Stdout.Fd()) // true for Windows Cygwin/MSYS2
```

### Role in Phase 3

**`go-isatty` is NOT needed as a direct dependency in Phase 3.** The `colorprofile.Detect` function performs TTY detection internally (it calls the same syscalls). `colorprofile.NoTTY` is returned when stdout is not a terminal.

If Phase 3 needs to decide whether to call `HasDarkBackground`, the guard is:

```go
p := colorprofile.Detect(os.Stdout, os.Environ())
if p == colorprofile.NoTTY || p == colorprofile.Ascii {
    // skip dark background detection
}
```

This is equivalent to `!isatty.IsTerminal(os.Stdout.Fd())` but uses the already-needed `colorprofile` package, avoiding an extra dependency.

**Decision: Do NOT add `go-isatty` as a dependency in Phase 3.** `colorprofile` subsumes it.

---

## Risk Areas and Mitigations

| Risk | Likelihood | Mitigation |
|------|-----------|------------|
| `HasDarkBackground` 2s hang in CI / test environments | HIGH | Guard with `colorprofile.Detect` first; skip OSC 11 query for `NoTTY`/`Ascii` profiles. Never call in tests without mocking. |
| Per-line tokenization loses multi-line string/comment context | MEDIUM | Accepted for v1; document as known limitation. Chroma's fallback tokens handle unparseable lines gracefully (they emit as `Text`). |
| ANSI color bleeding between diff lines | LOW | `TTY16m` formatter resets ANSI at line boundaries (`\033[0m` after each token). Each line is self-contained. |
| Unknown theme name passed by caller causes silent `NoOp` | MEDIUM | `selectTheme` falls back to auto-detected theme when `styles.Get(requested)` returns nil. Consider logging a warning. |
| `colorprofile.NewWriter` wrapping a non-`*os.File` writer | LOW | `colorprofile.NewWriter(w, env)` works with any `io.Writer`; for non-file writers it uses environment-only profile. Handle gracefully. |
| Module path for `charm.land/lipgloss/v2` vs `github.com/charmbracelet/lipgloss` | HIGH | MUST use `charm.land/lipgloss/v2` — this is the v2 canonical import path. The old GitHub path is v1 only. |
| go.mod `go 1.21` declaration but lipgloss v2 requires Go 1.25 (per pkg.go.dev) | MEDIUM | Verify with `go get` — module system enforces the minimum; may need to update `go 1.21` to `go 1.21` plus toolchain directive, or accept that users need a newer toolchain. Check if lipgloss v2 actually requires 1.25 at import (the `go` directive in go.mod may be advisory). |

---

## Recommended Implementation Sequence

### Plan 03-01: `internal/theme/` — Safe Dark Background Detection
- Add `charm.land/lipgloss/v2` and `github.com/charmbracelet/colorprofile` to `go.mod`
- Implement `theme.DetectDarkBackground(profile colorprofile.Profile) bool`
- Unit tests: mock TTY profile → verify returns correct bool without hanging

### Plan 03-02: `internal/highlight/` — Chroma Pipeline
- Add `github.com/alecthomas/chroma/v2` to `go.mod`
- Implement `highlight.HighlightLine(line, lexer, style, formatter)` 
- Implement `highlight.FormatterForProfile(p colorprofile.Profile) chroma.Formatter`
- Implement `highlight.SelectTheme(requested string, isDark bool) *chroma.Style`
- Unit tests: Go source line → verify ANSI output contains correct escape sequences for monokai; noop formatter returns plain text

### Plan 03-03: Language Detection + `WithLang`/`WithTheme`
- Implement `highlight.DetectLexer(filename, content string) chroma.Lexer` — tries `lexers.Get(lang)` (explicit), then `lexers.Match(filename)`, then `lexers.Analyse(content)`, then `lexers.Fallback`
- The `config.lang` and `config.theme` fields already exist; wire them through the highlight pipeline
- Unit tests: `.go` extension → Go lexer; explicit `"python"` override → Python lexer; unknown extension + Go content → fallback lexer

### Plan 03-04: `internal/render/unified.go` — UnifiedRenderer
- Implement `render.Unified(result DiffResult, w io.Writer, cfg *renderConfig)` 
- Writes `--- a\n+++ b\n` headers (if provided) and `@@ ... @@` hunk headers
- Applies highlighting per-line using `highlight` package
- Unit test: known Go diff → verify output contains correct hunk headers and `+`/`-` prefixes

### Plan 03-05: Public `drift.Render()` + `WithNoColor` + Color Depth Wire-up
- Add `render.go` to root package exposing `drift.Render(result DiffResult, w io.Writer, opts ...Option) error`
- `resolveProfile`: handles `WithNoColor`, `NO_COLOR` env var, non-file writers
- Wrap caller's `io.Writer` in `colorprofile.NewWriter` for automatic downsampling
- Integration test: call `drift.Diff` + `drift.Render` with `WithNoColor()` → verify no ANSI codes in output

---

## Package Structure to Create

```
internal/
  theme/
    theme.go          ← DetectDarkBackground(profile)
    theme_test.go
  highlight/
    highlight.go      ← HighlightLine, FormatterForProfile, SelectTheme, DetectLexer
    highlight_test.go
  render/
    unified.go        ← UnifiedRenderer (hunk header + line prefix + syntax color)
    unified_test.go

render.go             ← drift.Render() public API
render_test.go        ← integration tests (Diff + Render round-trip)
```

---

## New `go.mod` Dependencies to Add

```
charm.land/lipgloss/v2 v2.0.2
github.com/alecthomas/chroma/v2 v2.23.1
github.com/charmbracelet/colorprofile v0.4.2
```

Note: `colorprofile` is an indirect dependency of lipgloss v2 — `go get charm.land/lipgloss/v2` will add it. Add `github.com/alecthomas/chroma/v2` separately. `go-isatty` is NOT needed.

---

## Validation Architecture

### Per Success Criterion

| Success Criterion | Test | How to Verify |
|-------------------|------|---------------|
| 1. Unified diff `@@ -a,b +c,d @@` with `+`/`-` prefixes matching Git format | `TestUnified_HunkHeaders` | Compare `drift.Render()` output against `git diff` on same inputs using `exec.Command("git", "diff", "--no-index", "--unified=3")` |
| 2. Syntax highlighting via Chroma; Go tokens visually distinct | `TestHighlight_GoTokens` | Render a known Go diff; assert output contains `\033[` ANSI codes; assert `func` keyword has a different color sequence than a string literal |
| 3. Dark terminal → dark theme auto-detected; light terminal → light theme | `TestTheme_DarkDetection`, `TestTheme_LightDetection` | Mock colorprofile as TrueColor; mock `HasDarkBackground` to return true/false; verify `SelectTheme` returns `monokai` / `github` |
| 4. `WithTheme("monokai")` overrides auto-detection; `WithLang("go")` overrides extension | `TestOption_ThemeOverride`, `TestOption_LangOverride` | Call `Render` with explicit options; verify output matches expected style/lexer |
| 5. `NO_COLOR` / `WithNoColor()` strips all ANSI codes; 16-color and 256-color receive degraded output | `TestNoColor_EnvVar`, `TestNoColor_WithOption`, `TestColorDepth_256`, `TestColorDepth_16` | Set `NO_COLOR=1` env, call `Render` to `bytes.Buffer`, strip `+`/`-`/` ` prefixes, assert no `\033[` in output; for 256/16 color, assert `\033[38;5;` / `\033[3Xm` patterns respectively |

### Verification Commands

```bash
# After Phase 3 implementation:
go test ./...                    # all tests pass
go test -race ./...              # race-clean (highlight pipeline is single-goroutine)
go vet ./...                     # no vet issues
go build ./...                   # no compile errors

# Manual validation:
echo 'package main\nfunc main() {}' > /tmp/a.go
echo 'package main\nfunc main() { println("hello") }' > /tmp/b.go
go run ./cmd/drift /tmp/a.go /tmp/b.go  # (Phase 5 CLI, but can test via example)
# Inspect output for @@ headers, + green, - red
```

### Golden File Tests (using `goldie/v2`)

For each success criterion involving visual output, create golden files under `testdata/render/`:
- `unified_go_diff.golden` — Go syntax highlighted unified diff
- `unified_nocolor.golden` — same diff with NO_COLOR, plain text
- `unified_256color.golden` — same diff with ANSI256 color codes

Run `go test -update` to regenerate goldens after intentional style changes.

---

## Key API Reference Summary

```go
// colorprofile — profile detection
p := colorprofile.Detect(os.Stdout, os.Environ())
// p ∈ {TrueColor, ANSI256, ANSI, Ascii, NoTTY}

// colorprofile — automatic ANSI downsampling writer
w := colorprofile.NewWriter(os.Stdout, os.Environ())

// lipgloss v2 — dark background detection (TTY only)
isDark := lipgloss.HasDarkBackground(os.Stdin, os.Stdout)
// Returns true on error (2s timeout, non-TTY)

// chroma v2 — lexer selection
lexer := lexers.Get("go")           // by name
lexer = lexers.Match("file.go")     // by filename
lexer = lexers.Analyse(content)     // by content
lexer = chroma.Coalesce(lexer)      // coalesce tokens (always apply)
if lexer == nil { lexer = lexers.Fallback }

// chroma v2 — style (theme)
style := styles.Get("monokai")      // case-insensitive
if style == nil { style = styles.Fallback }

// chroma v2 — formatter
formatter := formatters.TTY16m      // terminal16m — TrueColor
formatter = formatters.TTY256       // terminal256 — 256-color
formatter = formatters.TTY16        // terminal16 — 16-color
formatter = formatters.NoOp         // noop — plain text

// chroma v2 — render to writer
iterator, _ := lexer.Tokenise(nil, lineContent)
var buf bytes.Buffer
formatter.Format(&buf, style, iterator)
highlighted := buf.String()
```

---

## RESEARCH COMPLETE
