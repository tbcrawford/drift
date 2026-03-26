package render

import (
	"strconv"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/alecthomas/chroma/v2"
	"github.com/tylercrawford/drift/internal/edittype"
	"github.com/tylercrawford/drift/internal/highlight"
)

// gutterNumberPadEachSide is the number of ASCII spaces placed before and after the
// line number digits inside the gutter cell (see [GutterNumberRender]).
const gutterNumberPadEachSide = 1

// gutterWidths returns minimum column widths for old and new line numbers in a hunk.
// Widths are based on the longest decimal string among lines with OldNum > 0 / NewNum > 0,
// plus one space on each side of the number.
// If no such line exists for a side, width is 1.
func gutterWidths(lines []edittype.Line) (oldW, newW int) {
	oldW, newW = 1, 1
	for _, ln := range lines {
		if ln.OldNum > 0 {
			if w := len(strconv.Itoa(ln.OldNum)) + 2*gutterNumberPadEachSide; w > oldW {
				oldW = w
			}
		}
		if ln.NewNum > 0 {
			if w := len(strconv.Itoa(ln.NewNum)) + 2*gutterNumberPadEachSide; w > newW {
				newW = w
			}
		}
	}
	return oldW, newW
}

// gutterStyleForCell returns a Lip Gloss style for one gutter column on one logical line.
// Delete rows on the old column and insert rows on the new column use the same semantic
// background as [highlight.WordSpanBackgroundColour] (matches highlighted changed words).
// Context lines use neutral [highlight.GutterBackgroundHex]. Callers should use
// [GutterNumberRender] so Width + center alignment fill the full gutter with background.
func gutterStyleForCell(style *chroma.Style, isDark, noColor bool, oldColumn bool, lineOp edittype.Op) lipgloss.Style {
	if noColor {
		return lipgloss.NewStyle()
	}
	var bg string
	switch {
	case style != nil && oldColumn && lineOp == edittype.Delete:
		c := highlight.WordSpanBackgroundColour(style, isDark, true)
		if c.IsSet() {
			bg = c.String()
		} else {
			bg = highlight.GutterBackgroundHex(isDark, true)
		}
	case style != nil && !oldColumn && lineOp == edittype.Insert:
		c := highlight.WordSpanBackgroundColour(style, isDark, false)
		if c.IsSet() {
			bg = c.String()
		} else {
			bg = highlight.GutterBackgroundHex(isDark, false)
		}
	default:
		bg = highlight.GutterBackgroundHex(isDark, oldColumn)
	}
	if isDark {
		return lipgloss.NewStyle().Background(lipgloss.Color(bg)).Foreground(lipgloss.Color("252"))
	}
	return lipgloss.NewStyle().Background(lipgloss.Color(bg)).Foreground(lipgloss.Color("240"))
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
	oldW, newW = 1, 1
	for _, p := range pairs {
		if p.leftOldNum > 0 {
			if w := len(strconv.Itoa(p.leftOldNum)) + 2*gutterNumberPadEachSide; w > oldW {
				oldW = w
			}
		}
		if p.rightNewNum > 0 {
			if w := len(strconv.Itoa(p.rightNewNum)) + 2*gutterNumberPadEachSide; w > newW {
				newW = w
			}
		}
	}
	return oldW, newW
}
