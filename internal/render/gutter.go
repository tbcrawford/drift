package render

import (
	"strconv"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/alecthomas/chroma/v2"
	"github.com/tbcrawford/drift/internal/edittype"
	"github.com/tbcrawford/drift/internal/highlight"
)

// gutterNumberPadEachSide is the number of ASCII spaces placed before and after the
// line number digits inside the gutter cell (see [GutterNumberRender]).
const gutterNumberPadEachSide = 1

// gutterColumnSeparator is space plus U+2502 (BOX DRAWINGS LIGHT VERTICAL, │). That is
// the conventional thin vertical for TUI / line-drawing borders (not ASCII U+007C |;
// the heavy box-drawing vertical is U+2503 ┃).
const gutterColumnSeparator = " │"

// gutterColWidth returns the minimum gutter column width needed to display any of the
// given line numbers: max(len(strconv.Itoa(n)) + 2*gutterNumberPadEachSide) for n > 0,
// or 1 when no positive number is present.
func gutterColWidth(nums []int) int {
	w := 1
	for _, n := range nums {
		if n > 0 {
			if dw := len(strconv.Itoa(n)) + 2*gutterNumberPadEachSide; dw > w {
				w = dw
			}
		}
	}
	return w
}

// gutterWidths returns minimum column widths for old and new line numbers in a hunk.
// Widths are based on the longest decimal string among lines with OldNum > 0 / NewNum > 0,
// plus one space on each side of the number.
// If no such line exists for a side, width is 1.
func gutterWidths(lines []edittype.Line) (oldW, newW int) {
	oldNums := make([]int, len(lines))
	newNums := make([]int, len(lines))
	for i, ln := range lines {
		oldNums[i] = ln.OldNum
		newNums[i] = ln.NewNum
	}
	return gutterColWidth(oldNums), gutterColWidth(newNums)
}

// gutterStyleForCell returns a Lip Gloss style for one gutter column on one logical line.
// Delete rows on the old column and insert rows on the new column use the same semantic
// background as [highlight.WordSpanBackgroundColour] (matches highlighted changed words).
// Context (unchanged) lines and blank gutter cells use **foreground only** — no gray fill,
// so the terminal default background shows through. Callers should use [GutterNumberRender]
// so Width + center alignment fill the column; padding spaces inherit the same style.
func gutterStyleForCell(style *chroma.Style, isDark, noColor bool, oldColumn bool, lineOp edittype.Op) lipgloss.Style {
	if noColor {
		return lipgloss.NewStyle()
	}
	dim := highlight.GutterDimForegroundHex(isDark)
	high := highlight.GutterHighlightForegroundHex(isDark)
	switch {
	case style != nil && oldColumn && lineOp == edittype.Delete:
		c := highlight.WordSpanBackgroundColour(style, isDark, true)
		if !c.IsSet() {
			return lipgloss.NewStyle().Background(lipgloss.Color(highlight.GutterBackgroundHex(isDark, true))).Foreground(lipgloss.Color(high))
		}
		return lipgloss.NewStyle().Background(lipgloss.Color(c.String())).Foreground(lipgloss.Color(high))
	case style != nil && !oldColumn && lineOp == edittype.Insert:
		c := highlight.WordSpanBackgroundColour(style, isDark, false)
		if !c.IsSet() {
			return lipgloss.NewStyle().Background(lipgloss.Color(highlight.GutterBackgroundHex(isDark, false))).Foreground(lipgloss.Color(high))
		}
		return lipgloss.NewStyle().Background(lipgloss.Color(c.String())).Foreground(lipgloss.Color(high))
	default:
		return lipgloss.NewStyle().Foreground(lipgloss.Color(dim))
	}
}

// gutterStyleKey uniquely identifies a gutter cell style variant.
// There are at most 6 distinct styles per render call (2 sides × 3 ops).
type gutterStyleKey struct {
	oldColumn bool
	op        edittype.Op
}

// GutterStyleCache pre-computes all gutter cell styles for a render call.
// Build it once at the start of a render; reuse for every line in every hunk.
// This eliminates per-line lipgloss.NewStyle() allocations in the hot render loop.
type GutterStyleCache struct {
	styles map[gutterStyleKey]lipgloss.Style
}

// NewGutterStyleCache builds a GutterStyleCache from the resolved render config.
// Call once per Render invocation (not per hunk, not per line).
// Pre-populates all 6 combinations of (oldColumn: bool) × (op: Equal|Delete|Insert).
func NewGutterStyleCache(style *chroma.Style, isDark, noColor bool) *GutterStyleCache {
	cache := &GutterStyleCache{
		styles: make(map[gutterStyleKey]lipgloss.Style, 6),
	}
	for _, oldCol := range []bool{true, false} {
		for _, op := range []edittype.Op{edittype.Equal, edittype.Delete, edittype.Insert} {
			k := gutterStyleKey{oldColumn: oldCol, op: op}
			cache.styles[k] = gutterStyleForCell(style, isDark, noColor, oldCol, op)
		}
	}
	return cache
}

// Get returns the pre-computed lipgloss.Style for the given gutter cell variant.
func (c *GutterStyleCache) Get(oldColumn bool, op edittype.Op) lipgloss.Style {
	return c.styles[gutterStyleKey{oldColumn: oldColumn, op: op}]
}

// styledGutterColumnSeparator returns the gutter column separator with dim foreground when color is on.
// Uses cfg.GutterMiddleSep when non-empty; falls back to gutterColumnSeparator (" │").
func styledGutterColumnSeparator(cfg *RenderConfig) string {
	sep := gutterColumnSeparator // default " │"
	if cfg != nil && cfg.GutterMiddleSep != "" {
		sep = cfg.GutterMiddleSep
	}
	if cfg == nil || cfg.NoColor {
		return sep
	}
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(highlight.GutterDimForegroundHex(cfg.IsDark))).
		Render(sep)
}

// GutterNumberRender renders a line number (or blank when n==0) in a fixed display width.
// Non-zero numbers are wrapped with [gutterNumberPadEachSide] spaces on each side, then
// centered in the column; padding uses the same Lip Gloss style so the background fills
// the entire gutter width.
func GutterNumberRender(st lipgloss.Style, width int, n int) string {
	if width < 1 {
		width = 1
	}
	if n == 0 {
		return st.Width(width).AlignHorizontal(lipgloss.Center).Render("")
	}
	s := strconv.Itoa(n)
	inner := strings.Repeat(" ", gutterNumberPadEachSide) + s + strings.Repeat(" ", gutterNumberPadEachSide)
	if len(inner) > width {
		if len(s) >= width {
			return st.Render(s)
		}
		return st.Width(width).AlignHorizontal(lipgloss.Center).Render(s)
	}
	return st.Width(width).AlignHorizontal(lipgloss.Center).Render(inner)
}

func gutterPairWidths(pairs []linePair) (oldW, newW int) {
	oldNums := make([]int, len(pairs))
	newNums := make([]int, len(pairs))
	for i, p := range pairs {
		oldNums[i] = p.leftOldNum
		newNums[i] = p.rightNewNum
	}
	return gutterColWidth(oldNums), gutterColWidth(newNums)
}
