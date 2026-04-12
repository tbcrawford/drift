package render

import (
	"fmt"
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
// All rows use foreground-only styling — no background color — so the terminal default
// background shows through even on delete/insert rows. Changed lines use the diff-line
// foreground color: red for delete rows, green for insert rows (derived from the Chroma
// style's semantic diff colors, matching the line highlight color family). Context lines
// use the dim gray foreground.
// Callers should use [GutterNumberRender] so Width + alignment fill the column.
func gutterStyleForCell(style *chroma.Style, isDark, noColor bool, oldColumn bool, lineOp edittype.Op) lipgloss.Style {
	if noColor {
		return lipgloss.NewStyle()
	}
	dim := highlight.GutterDimForegroundHex(isDark)
	isDelete := lineOp == edittype.Delete
	switch {
	case style != nil && oldColumn && lineOp == edittype.Delete:
		// Delete row old-column: use the word-span red (brighter than the line bg).
		c := highlight.WordSpanBackgroundColour(style, isDark, isDelete)
		if c.IsSet() {
			hex := fmt.Sprintf("#%02x%02x%02x", c.Red(), c.Green(), c.Blue())
			return lipgloss.NewStyle().Foreground(lipgloss.Color(hex))
		}
		return lipgloss.NewStyle().Foreground(lipgloss.Color(highlight.GutterHighlightForegroundHex(isDark)))
	case style != nil && !oldColumn && lineOp == edittype.Insert:
		// Insert row new-column: use the word-span green (brighter than the line bg).
		c := highlight.WordSpanBackgroundColour(style, isDark, false)
		if c.IsSet() {
			hex := fmt.Sprintf("#%02x%02x%02x", c.Red(), c.Green(), c.Blue())
			return lipgloss.NewStyle().Foreground(lipgloss.Color(hex))
		}
		return lipgloss.NewStyle().Foreground(lipgloss.Color(highlight.GutterHighlightForegroundHex(isDark)))
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

// gutterBorderFg returns the hex color string to use for │ border characters.
// Uses cfg.GutterBorderColor when set; falls back to dim gray.
func gutterBorderFg(cfg *RenderConfig) string {
	if cfg != nil && cfg.GutterBorderColor != "" {
		return cfg.GutterBorderColor
	}
	return highlight.GutterDimForegroundHex(cfg.IsDark)
}

// styledGutterColumnSeparator returns the gutter column separator with dim foreground when color is on.
// Uses cfg.GutterMiddleSep when non-empty; falls back to gutterColumnSeparator (" │").
// Used in unified mode only — split mode uses styledSplitPanelSep.
func styledGutterColumnSeparator(cfg *RenderConfig) string {
	sep := gutterColumnSeparator // default " │"
	if cfg != nil && cfg.GutterMiddleSep != "" {
		sep = cfg.GutterMiddleSep
	}
	if cfg == nil || cfg.NoColor {
		return sep
	}
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(gutterBorderFg(cfg))).
		Render(sep)
}

// splitPanelSepPlain returns the plain (unstyled) panel separator for split mode.
// When cfg.SplitPanelSep is set, it is used directly. Otherwise, if ShowLineNumbers
// is true and GutterCellBorder is non-empty, returns "" (the border acts as separator).
// Falls back to gutterColumnSeparator (" │") in all other cases.
func splitPanelSepPlain(cfg *RenderConfig) string {
	if cfg == nil {
		return gutterColumnSeparator
	}
	if cfg.SplitPanelSep != "" {
		return cfg.SplitPanelSep
	}
	if cfg.ShowLineNumbers && cfg.GutterCellBorder != "" {
		return "" // gutter border acts as visual separator
	}
	return gutterColumnSeparator
}

// styledSplitPanelSep returns the styled panel separator for split mode.
// sep must be the pre-computed plain separator from splitPanelSepPlain.
func styledSplitPanelSep(cfg *RenderConfig, sep string) string {
	if sep == "" {
		return ""
	}
	if cfg == nil || cfg.NoColor {
		return sep
	}
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(gutterBorderFg(cfg))).
		Render(sep)
}

// styledGutterCellBorder returns the gutter cell border with the configured foreground color.
// Used in split mode when cfg.GutterCellBorder is non-empty to produce "│ NNN │" gutter cells.
// Uses GutterBorderColor when set (accent blue for DeltaTheme); falls back to dim gray.
func styledGutterCellBorder(cfg *RenderConfig) string {
	if cfg == nil || cfg.GutterCellBorder == "" {
		return ""
	}
	if cfg.NoColor {
		return cfg.GutterCellBorder
	}
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(gutterBorderFg(cfg))).
		Render(cfg.GutterCellBorder)
}

// GutterNumberRender renders a line number (or blank when n==0) in a fixed display width.
// Non-zero numbers are right-aligned within the column with [gutterNumberPadEachSide]
// spaces on the right side, and the remaining space on the left — so numbers like
// 2, 10, 100 all align on the right edge of the cell: "  2 ", " 10 ", "100 ".
// The style is applied to the whole cell so borders and backgrounds render uniformly.
func GutterNumberRender(st lipgloss.Style, width int, n int) string {
	if width < 1 {
		width = 1
	}
	if n == 0 {
		return st.Width(width).Render("")
	}
	s := strconv.Itoa(n)
	// Right-align: pad right side fixed, left side takes remaining space.
	padRight := strings.Repeat(" ", gutterNumberPadEachSide)
	inner := s + padRight
	if len(inner) >= width {
		// Number is wider than cell — just render the number directly.
		return st.Render(s)
	}
	// Pad left so that inner right-aligns within the full width.
	padLeft := strings.Repeat(" ", width-len(inner))
	return st.Render(padLeft + s + padRight)
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
