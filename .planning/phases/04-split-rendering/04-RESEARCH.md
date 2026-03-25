# Phase 4: Split Rendering — Research

**Phase Goal:** Callers can request side-by-side split diff output with correct two-panel layout at any terminal width
**Requirements:** REND-02
**Research Date:** 2026-03-25

---

## Executive Summary

**Key decisions:**

1. **`lipgloss.JoinHorizontal` is the correct primitive** — it aligns multi-line string blocks side-by-side along a vertical axis and handles differing heights between panels (pads the shorter panel). Import path is `charm.land/lipgloss/v2` (NOT `github.com/charmbracelet/lipgloss`).

2. **`lipgloss.Width()` is the only correct width measurement for ANSI-highlighted content** — `len(string)` and `len([]rune(string))` both over-count because they include invisible ANSI escape sequence bytes. Lip Gloss `Width()` strips ANSI and measures Unicode cell width correctly. Always use `lipgloss.Width()` for panel content.

3. **Terminal width detection uses `github.com/charmbracelet/x/term.GetSize(fd)`** — already an indirect dependency via lipgloss v2. Falls back to `COLUMNS` env var when output is not a TTY (piped). If neither is available, defaults to 80 columns (safe conservative fallback, never panics).

4. **`runewidth.EastAsianWidth = false`** (the default) is the correct setting for developer tooling — we do not change the global default. Lip Gloss internally uses `rivo/uniseg` (also already in go.mod) for Unicode width, which is more accurate than `mattn/go-runewidth` for emoji and ZWJ sequences.

5. **Split rendering does not need a new public API function** — add a `WithSplit()` option that routes `drift.Render()` to `render.Split()` instead of `render.Unified()`. The caller-facing API stays identical.

6. **Panel alignment strategy:** Each panel is `(termWidth - 1) / 2` columns wide (the -1 accounts for the separator). Lines are truncated at `panelWidth` using `lipgloss.NewStyle().MaxWidth(panelWidth)` before being passed to `JoinHorizontal`. This prevents overflow at any terminal width.

7. **Hunk pairing strategy:** The split renderer must pair old (Delete) and context lines in the left panel against new (Insert) and context lines in the right panel. Equal lines appear in both panels. Deleted lines appear only in the left; inserted lines appear only in the right. A "blank" placeholder fills the opposite panel for unmatched lines.

---

## Research Area 1: Lip Gloss v2 `JoinHorizontal` — API and Two-Panel Layout

### Import Path (Critical)

```go
import "charm.land/lipgloss/v2"
```

**NOT** `github.com/charmbracelet/lipgloss` (that is v1 — has I/O race bugs). The v2 canonical import path changed in February 2026.

### `JoinHorizontal` Signature

```go
func JoinHorizontal(pos Position, strs ...string) string
```

- `pos`: vertical alignment for blocks of different heights. Use `lipgloss.Top` (0.0) for diff panels — both panels should start at the top edge.
- `strs`: one string per column. Each string may contain embedded `\n` newlines to form a multi-line block.
- Returns a single string with the blocks placed side-by-side, padded to the same height.

### Usage Pattern for Two-Panel Diff

```go
// Build each panel as a multi-line string block
var leftLines, rightLines []string
// ... populate left (old) and right (new) panels ...

leftBlock := strings.Join(leftLines, "\n")
rightBlock := strings.Join(rightLines, "\n")

// Join horizontally at the top
row := lipgloss.JoinHorizontal(lipgloss.Top, leftBlock, sepBlock, rightBlock)
```

### `lipgloss.Width()` — ANSI-Aware Width Measurement

```go
func Width(str string) (width int)
```

From godoc: "ANSI sequences are ignored and characters wider than one cell (such as Chinese characters and emojis) are appropriately measured."

**Critical for diff rendering:** Chroma-highlighted lines contain ANSI escape sequences like `\033[38;2;255;183;0m` (TrueColor foreground). These are zero-width for display purposes but contribute bytes/runes to the string. `len(highlighted)` will be much larger than the actual terminal display width. Always use `lipgloss.Width(highlighted)` when measuring content for panel width budgets.

### Style-Based Width Enforcement

```go
panelStyle := lipgloss.NewStyle().MaxWidth(panelWidth)
truncated := panelStyle.Render(line)
```

`MaxWidth` truncates (not wraps) the rendered string at `panelWidth` terminal cells, correctly handling ANSI sequences within the content. This is simpler than manual truncation and handles Unicode wide characters.

Alternatively, `lipgloss.NewStyle().Width(panelWidth)` pads the line to exactly `panelWidth` cells — useful for enforcing fixed-width columns.

**Recommendation:** Use `.Width(panelWidth)` to force equal-width panels so `JoinHorizontal` doesn't need to align unequal-length lines; this gives consistent separator placement regardless of content length.

---

## Research Area 2: ANSI-Aware Width — `lipgloss.Width()` vs `len()`

### The Problem

```go
line := "\033[38;2;255;183;0mfunc\033[0m main() {}"

len(line)          // e.g., 42  — includes ANSI bytes; WRONG for display width
len([]rune(line))  // e.g., 38  — runes include escape chars; still WRONG
lipgloss.Width(line) // e.g., 15  — strips ANSI, counts display cells; CORRECT
```

### Why It Matters for Split View

If panel width is set to 40 columns but content measurement uses `len()`, you would compute that a line with 15 visible characters "fills" 42 columns (due to ANSI bytes), incorrectly conclude it overflows, and truncate the visible content. The output would appear garbled.

Using `lipgloss.Width()` correctly reports 15 display cells, leaving room for 25 more characters before truncation is needed.

### Verification Test

The Phase 4 test suite should include:

```go
// Verify ANSI-aware width measurement
func TestWidth_ANSIStrip(t *testing.T) {
    plain := "func main() {}"
    highlighted, _ := highlight.HighlightLine(plain, goLexer, monokaiStyle, formatters.TTY16m)
    
    if len(highlighted) == len(plain) {
        t.Error("expected highlighted string to be longer than plain due to ANSI codes")
    }
    if lipgloss.Width(highlighted) != lipgloss.Width(plain) {
        t.Errorf("lipgloss.Width(highlighted)=%d; want %d (same as plain)", 
            lipgloss.Width(highlighted), lipgloss.Width(plain))
    }
}
```

---

## Research Area 3: Terminal Width Detection and Pipe Fallback

### Available API — Already in `go.mod`

`github.com/charmbracelet/x/term` is already an indirect dependency (pulled in by lipgloss v2):

```go
import "github.com/charmbracelet/x/term"

width, height, err := term.GetSize(os.Stdout.Fd())
```

Returns `(0, 0, error)` when stdout is not a TTY (piped). Does NOT panic.

### Resolution Order

```go
// internal/render/termwidth.go
func TerminalWidth(w io.Writer) int {
    // 1. If w is a TTY *os.File, query the terminal
    if f, ok := w.(*os.File); ok {
        if width, _, err := term.GetSize(f.Fd()); err == nil && width > 0 {
            return width
        }
    }
    // 2. COLUMNS env var (set by shell, respected by diff tools like delta)
    if cols := os.Getenv("COLUMNS"); cols != "" {
        if n, err := strconv.Atoi(cols); err == nil && n > 0 {
            return n
        }
    }
    // 3. Safe default for piped output
    return 80
}
```

**Default of 80**: Industry standard for "safe" terminal width. When output is piped, 80 columns ensures the diff is readable without overflow even on narrow displays.

### Pipe Detection

When `drift render --split ... | less`, `os.Stdout` is not a TTY. `term.GetSize` fails, so `COLUMNS` is checked, then the 80-column default applies. The output is still valid; it just uses the narrower fallback width.

### Existing Precedents

- `git diff --stat` uses 80 columns when piped
- `delta` (a popular diff pager) uses `COLUMNS` env var
- `bat` falls back to 80 when no TTY width is available

---

## Research Area 4: Unicode Wide Characters — `runewidth` / `uniseg`

### The Global Setting

```go
// go-runewidth global
var runewidth.EastAsianWidth bool // default false
```

When `EastAsianWidth = true`, ambiguous-width Unicode characters (e.g., certain box-drawing chars, some punctuation) are counted as 2 cells instead of 1. This is for CJK locale users. **For a developer diff tool targeting international source code, the default `false` is correct** — we do not want to treat common ASCII-adjacent characters as double-width.

**Decision: Do NOT modify `runewidth.EastAsianWidth`.** Leave it at the default `false`. This matches how GitHub's diff view, `git diff`, and most terminals measure column widths in developer contexts.

### How Lip Gloss Handles It

Lip Gloss v2 uses `rivo/uniseg` (also in go.mod) for grapheme cluster segmentation, which correctly handles:
- Combining characters (accents, diacritics)
- Zero-width joiners (ZWJ sequences for emoji families)
- Regional indicators (flag emoji)
- Variation selectors

`lipgloss.Width()` delegates to this grapheme-cluster-aware measurement. You do NOT need to call `runewidth.StringWidth()` directly — always use `lipgloss.Width()` for display-width measurement in the split renderer.

### Line Truncation for Wide Characters

When a source line contains wide Unicode characters (e.g., a comment in Chinese):

```go
// Line "// 这是注释" takes 12 display columns, not 6 rune positions
panelStyle := lipgloss.NewStyle().MaxWidth(panelWidth)
truncated := panelStyle.Render(line) // Lip Gloss truncates at cell boundary
```

Lip Gloss correctly handles partial wide-character truncation — it will not split a 2-cell character in half, instead rounding down.

---

## Research Area 5: Panel Layout and Line Pairing Strategy

### Panel Width Calculation

```go
totalWidth := TerminalWidth(w)
// Reserve 1 column for separator (" │ " or " | ")
sepWidth    := 3  // " │ " is 3 display cells
panelWidth  := (totalWidth - sepWidth) / 2
// Left panel: panelWidth, separator: sepWidth, right panel: totalWidth - sepWidth - panelWidth
rightPanelWidth := totalWidth - sepWidth - panelWidth
```

For odd total widths, this gives left panel 1 column narrower than right — acceptable.

### Hunk Line Pairing

The core challenge in split view: a hunk may have N delete lines and M insert lines where N ≠ M. The split view must pair them row-by-row.

**Pairing algorithm:**

1. Walk the `hunk.Lines` slice.
2. Equal lines → emit to both panels on the same row.
3. When encountering a run of Delete/Insert lines, consume all adjacent deletes and inserts:
   - Pair them 1:1 (left=delete, right=insert) while both lists have entries.
   - When deletes run out, emit (blank, insert) rows.
   - When inserts run out, emit (delete, blank) rows.

```
Hunk lines: [Equal, Delete, Delete, Insert, Equal, Delete, Insert, Insert]

Row 1: left=" equal"    right=" equal"      (Equal, Equal)
Row 2: left="-delete1"  right="+insert1"    (Delete, Insert)  — paired
Row 3: left="-delete2"  right="           " (Delete, blank)   — unpaired delete
Row 4: left=" equal"    right=" equal"      (Equal, Equal)
Row 5: left="-delete"   right="+insert1"    (Delete, Insert)  — paired
Row 6: left="          " right="+insert2"   (blank, Insert)   — unpaired insert
```

### Separator Style

A vertical separator between panels aids readability:

```go
const separator = " │ "  // Unicode box-drawing char, 3 display cells
```

For `NoColor` mode, use `" | "` (ASCII pipe) to avoid non-ASCII in plain output.

The separator height must match the number of rows in the rendered block. Since `JoinHorizontal` is called on the complete left and right column strings, the separator must also be a multi-line string with the same height:

```go
sepLines := strings.Repeat(separator+"\n", rowCount)
sepBlock  := strings.TrimRight(sepLines, "\n")
row := lipgloss.JoinHorizontal(lipgloss.Top, leftBlock, sepBlock, rightBlock)
```

### Hunk Headers in Split View

Hunk headers (`@@ -a,b +c,d @@`) span both panels:

```
@@ -1,5 +1,6 @@
```

Two options:
1. Print the header full-width above both panels (simpler, matches standard diff tools).
2. Print the header only in the left panel, blank in the right.

**Recommendation:** Print full-width hunk header (spanning the full terminal width) above each hunk block. This avoids ambiguity about which hunk the header belongs to and matches how GitHub's side-by-side diff renders hunk markers.

---

## Research Area 6: Phase 3 Unified Renderer Patterns to Follow

Reading `internal/render/unified.go`, the established patterns are:

### 1. `RenderConfig` struct — pre-resolved dependencies

```go
type RenderConfig struct {
    OldName   string
    NewName   string
    Lang      string
    Lexer     chroma.Lexer        // pre-resolved; nil triggers auto-detect
    Style     *chroma.Style       // pre-resolved
    Formatter chroma.Formatter    // pre-resolved
    Profile   colorprofile.Profile
    NoColor   bool
}
```

The split renderer **reuses the same `RenderConfig`** — no new config struct needed. Add `TermWidth int` to `RenderConfig` to carry the resolved terminal width into the renderer (resolved by the public `drift.Render()` or `drift.RenderSplit()` function, not inside the renderer itself).

### 2. Function signature pattern

```go
// internal/render/split.go
func Split(result edittype.DiffResult, w io.Writer, cfg *RenderConfig) error
```

Mirrors `Unified(result edittype.DiffResult, w io.Writer, cfg *RenderConfig) error` exactly.

### 3. Fail-open on highlight errors

```go
highlighted, err := highlight.HighlightLine(line.Content, lexer, style, formatter)
if err != nil {
    highlighted = line.Content // plain text fallback
}
```

Same fail-open pattern used in unified renderer — never block output on a highlight error.

### 4. Empty result fast path

```go
if len(result.Hunks) == 0 {
    return nil
}
```

### 5. Public API routing

The public `drift.Render()` in `render.go` currently calls `render.Unified(...)`. With a `WithSplit()` option, the routing becomes:

```go
// render.go
func Render(result DiffResult, w io.Writer, opts ...Option) error {
    cfg := defaultConfig()
    for _, o := range opts { o(cfg) }
    // ... resolve profile, lexer, style, formatter, termWidth ...
    rcfg := &render.RenderConfig{...}
    if cfg.split {
        return render.Split(result, wrapped, rcfg)
    }
    return render.Unified(result, wrapped, rcfg)
}
```

---

## Research Area 7: Existing Codebase Integration Points

### What Already Exists (verified by reading source)

**`config` struct (`options.go`):**
```go
type config struct {
    algorithm    Algorithm
    contextLines int
    noColor      bool
    lang         string
    theme        string
    // split bool  ← ADD THIS in plan 04-03
}
```

**`render.RenderConfig` (`internal/render/unified.go`):**
```go
type RenderConfig struct {
    OldName   string
    NewName   string
    Lang      string
    Lexer     chroma.Lexer
    Style     *chroma.Style
    Formatter chroma.Formatter
    Profile   colorprofile.Profile
    NoColor   bool
    // TermWidth int  ← ADD THIS in plan 04-01
}
```

**`drift.Render()` and `drift.RenderWithNames()` (`render.go`):**
Already resolve `profile`, `isDark`, `lexer`, `style`, `formatter` and call `render.Unified`. Phase 4 adds `termWidth` resolution and split routing.

**Terminal width packages already in go.mod:**
- `github.com/charmbracelet/x/term v0.2.2` — `term.GetSize(fd)` for TTY width
- `github.com/mattn/go-runewidth v0.0.19` — Unicode width (used internally by lipgloss)
- `charm.land/lipgloss/v2 v2.0.2` — `lipgloss.Width()`, `JoinHorizontal`, `Style.Width()`

No new dependencies need to be added to `go.mod`.

### Import Graph (no new cycles)

```
drift (root)
  → internal/render/split     (new)
  → internal/render/unified   (existing)
  → internal/highlight        (existing)
  → internal/edittype         (existing)

internal/render/split
  → charm.land/lipgloss/v2    (existing indirect dep → make direct)
  → github.com/charmbracelet/x/term (existing indirect dep → make direct)
  → internal/highlight
  → internal/edittype
```

Both `charm.land/lipgloss/v2` and `github.com/charmbracelet/x/term` are already present as indirect deps. Phase 4 makes them direct deps by importing them explicitly in non-test `.go` files.

---

## Research Area 8: Go Module Setup

### Current `go.mod` (verified)

```
module github.com/tylercrawford/drift

go 1.25.0

require pgregory.net/rapid v1.2.0

require (
    charm.land/lipgloss/v2 v2.0.2 // indirect
    github.com/alecthomas/chroma/v2 v2.23.1 // indirect
    github.com/charmbracelet/colorprofile v0.4.2 // indirect
    github.com/charmbracelet/x/term v0.2.2 // indirect
    github.com/mattn/go-runewidth v0.0.19 // indirect
    github.com/rivo/uniseg v0.4.7 // indirect
    ...
)
```

`go 1.25.0` is already set (was `1.21` in STACK.md but got updated during Phase 3 — go.mod shows `1.25.0`).

### Dependencies Phase 4 Promotes from `indirect` to `direct`

When `internal/render/split.go` imports `charm.land/lipgloss/v2` and `github.com/charmbracelet/x/term` directly, `go mod tidy` will remove the `// indirect` comment from those lines (promoting them to direct deps). No new packages to add.

---

## Research Area 9: Implementation Plan Breakdown

### Plan 04-01: `internal/render/split.go` — Core Split Renderer

**Files to create/modify:**
- `internal/render/split.go` — `Split()` function
- `internal/render/unified.go` — add `TermWidth int` to `RenderConfig`
- `internal/render/split_test.go` — unit tests

**Key work:**
1. Add `TermWidth int` to `RenderConfig` (default 0 → treat as 80 in Split renderer)
2. Implement `Split(result edittype.DiffResult, w io.Writer, cfg *RenderConfig) error`
3. Implement `pairHunkLines(lines []edittype.Line) []linePair` — the pairing algorithm
4. Implement `renderPanel(line string, op edittype.Op, width int, cfg *RenderConfig) string` — highlight + pad to width
5. Use `lipgloss.JoinHorizontal(lipgloss.Top, leftBlock, sepBlock, rightBlock)` per hunk
6. Write full-width hunk header before each hunk

### Plan 04-02: Terminal Width Detection + Unicode Handling

**Files to create/modify:**
- `internal/render/termwidth.go` — `TerminalWidth(w io.Writer) int`
- `internal/render/termwidth_test.go`

**Key work:**
1. Implement `TerminalWidth(w io.Writer) int` with TTY → COLUMNS → 80 fallback chain
2. Verify `lipgloss.Width()` vs `len()` with ANSI-highlighted test lines
3. Add test for Unicode wide characters (CJK) truncation at panel boundary
4. Document that `runewidth.EastAsianWidth` is left at default `false`

### Plan 04-03: Wire `WithSplit()` into `drift.Render()`

**Files to create/modify:**
- `options.go` — add `split bool` to `config`, add `WithSplit()` option
- `render.go` — add `termWidth` resolution; route to `render.Split` when `cfg.split == true`
- `render_test.go` — integration tests for split output

**Key work:**
1. Add `split bool` field to `config` struct
2. Add `WithSplit() Option` function
3. In `Render()` and `RenderWithNames()`, resolve `termWidth := render.TerminalWidth(w)` and set `rcfg.TermWidth = termWidth`
4. Route: `if cfg.split { return render.Split(result, wrapped, rcfg) }`
5. Integration test: call `drift.Diff` + `drift.Render(..., drift.WithSplit())` → verify two-panel output

---

## Validation Architecture

### Per Success Criterion

| Success Criterion | Test | How to Verify |
|-------------------|------|---------------|
| 1. Split diff shows left (old) and right (new) panels side-by-side with syntax highlighting | `TestSplit_TwoPanels` | Render a known diff; assert output contains the separator (`│`); assert left panel has `-`-prefixed old lines; assert right panel has `+`-prefixed new lines; assert ANSI codes present in both panels |
| 2. Correct rendering at 80, 120, 200 columns | `TestSplit_Width80`, `TestSplit_Width120`, `TestSplit_Width200` | Inject `TermWidth` directly into `RenderConfig`; verify each output line has `lipgloss.Width(line) <= totalWidth`; verify no line exceeds expected width |
| 3. ANSI sequences do not inflate measured panel width | `TestWidth_ANSIStrip` | Highlight a known Go line; assert `lipgloss.Width(highlighted) == lipgloss.Width(plainText)`; assert `len(highlighted) > len(plainText)` |
| 4. Pipe fallback (no TTY) uses safe default width | `TestTerminalWidth_PipeFallback` | Call `TerminalWidth(&bytes.Buffer{})` (non-`*os.File` writer); assert result is 80 |

### Additional Validation Tests

```go
// Verify panel width is exactly half of terminal width (minus separator)
func TestSplit_PanelWidthSymmetry(t *testing.T) {
    // Render with TermWidth=80; expect panels of (80-3)/2 = 38 columns each
}

// Verify COLUMNS env var is respected
func TestTerminalWidth_COLUMNSEnvVar(t *testing.T) {
    t.Setenv("COLUMNS", "120")
    w := &bytes.Buffer{} // non-TTY writer
    width := render.TerminalWidth(w)
    if width != 120 { t.Errorf("want 120, got %d", width) }
}

// Verify NoColor produces ASCII separator
func TestSplit_NoColorSeparator(t *testing.T) {
    // Render with NoColor=true; separator should be " | " not " │ "
}
```

### Verification Commands

```bash
# After Phase 4 implementation:
go test ./...               # all tests pass
go test -race ./...         # race-clean
go vet ./...                # no vet issues
go build ./...              # no compile errors
go mod tidy                 # no spurious indirect markers

# Manual validation (using a small test program):
cat > /tmp/splitdemo.go << 'EOF'
package main

import (
    "fmt"
    "os"
    "github.com/tylercrawford/drift"
)

func main() {
    a := "package main\n\nfunc hello() {\n\tfmt.Println(\"hello\")\n}\n"
    b := "package main\n\nfunc hello() {\n\tfmt.Println(\"hello, world\")\n\tfmt.Println(\"done\")\n}\n"
    result, _ := drift.Diff(a, b, drift.WithLang("go"))
    drift.Render(result, os.Stdout, drift.WithSplit(), drift.WithLang("go"))
    fmt.Println()
}
EOF
go run /tmp/splitdemo.go
# Inspect output: two columns with │ separator, highlighted left (old) and right (new)
```

---

## Tradeoffs and Edge Cases

### Edge Case: Hunk with Only Deletions

```
Left panel:   -line1 / -line2 / -line3
Right panel:  blank  / blank  / blank
```

The right panel should show empty styled cells, not be absent. Use `lipgloss.NewStyle().Width(panelWidth).Render("")` to produce a correctly-padded blank cell.

### Edge Case: Hunk with Only Insertions

```
Left panel:   blank  / blank
Right panel:  +line1 / +line2
```

Same as above but mirrored.

### Edge Case: Very Long Lines Exceeding Panel Width

A line like `// generated code: aaaaa...` with 200+ characters in a 40-column panel must be truncated. Use `lipgloss.NewStyle().MaxWidth(panelWidth)` — it truncates (with optional ellipsis) rather than wrapping. Wrapping would break the row-pairing invariant.

### Edge Case: Empty Input (No Hunks)

```go
if len(result.Hunks) == 0 {
    return nil // identical inputs — nothing to render
}
```

### Edge Case: TermWidth < 20

If `COLUMNS=10` is set by a pathological caller, `panelWidth` would be (10-3)/2 = 3 columns — unusably narrow. Consider a minimum `TermWidth` of 40 (2×20 columns). Clamp: `if termWidth < 40 { termWidth = 40 }`.

### Edge Case: NoColor in Split View

When `cfg.NoColor == true`:
- `cfg.Formatter` will be `formatters.NoOp` (no ANSI)
- Separator should be `" | "` (ASCII) instead of `" │ "` (Unicode) for clean piped output
- `lipgloss.Width()` still works correctly on plain text (trivially: equals `len(rune(s))` for ASCII)

---

## Key API Reference Summary

```go
// charm.land/lipgloss/v2
lipgloss.Width(str string) int                  // ANSI-aware cell width
lipgloss.JoinHorizontal(pos, strs...) string    // side-by-side block join
lipgloss.NewStyle().Width(n int) lipgloss.Style  // fixed-width cell (pads/truncates)
lipgloss.NewStyle().MaxWidth(n int) lipgloss.Style // truncate at n cells

// github.com/charmbracelet/x/term
term.GetSize(fd uintptr) (width, height int, err error) // TTY size; err on non-TTY

// github.com/mattn/go-runewidth (via lipgloss — do NOT call directly)
// Use lipgloss.Width() instead

// Standard library
os.Getenv("COLUMNS")          // pipe fallback: shell-set terminal width
strconv.Atoi(cols)            // parse COLUMNS value
```

---

## RESEARCH COMPLETE
