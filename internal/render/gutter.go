package render

import (
	"strconv"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/alecthomas/chroma/v2"
	"github.com/tylercrawford/drift/internal/edittype"
	"github.com/tylercrawford/drift/internal/highlight"
)

// gutterWidths returns minimum column widths for old and new line numbers in a hunk.
// Widths are based on the longest decimal string among lines with OldNum > 0 / NewNum > 0.
// If no such line exists for a side, width is 1.
func gutterWidths(lines []edittype.Line) (oldW, newW int) {
	oldW, newW = 1, 1
	for _, ln := range lines {
		if ln.OldNum > 0 {
			if w := len(strconv.Itoa(ln.OldNum)); w > oldW {
				oldW = w
			}
		}
		if ln.NewNum > 0 {
			if w := len(strconv.Itoa(ln.NewNum)); w > newW {
				newW = w
			}
		}
	}
	return oldW, newW
}

// centerLineNumber renders n centered in a field of display width width.
// When n == 0, returns width spaces (blank gutter cell).
func centerLineNumber(n int, width int) string {
	if width < 1 {
		width = 1
	}
	if n == 0 {
		return strings.Repeat(" ", width)
	}
	s := strconv.Itoa(n)
	if len(s) >= width {
		return s
	}
	pad := width - len(s)
	left := pad / 2
	right := pad - left
	return strings.Repeat(" ", left) + s + strings.Repeat(" ", right)
}

// gutterStyleForCell returns a Lip Gloss style for one gutter column on one logical line.
// Line-number columns use neutral grays from [highlight.GutterBackgroundHex] for every
// row; semantic add/remove backgrounds are applied to the full code line (prefix + body)
// by the unified/split renderers, not in the gutter cells.
func gutterStyleForCell(_ *chroma.Style, isDark, noColor bool, oldColumn bool, _ edittype.Op) lipgloss.Style {
	if noColor {
		return lipgloss.NewStyle()
	}
	bg := lipgloss.Color(highlight.GutterBackgroundHex(isDark, oldColumn))
	if isDark {
		return lipgloss.NewStyle().Background(bg).Foreground(lipgloss.Color("252"))
	}
	return lipgloss.NewStyle().Background(bg).Foreground(lipgloss.Color("240"))
}

func gutterPairWidths(pairs []linePair) (oldW, newW int) {
	oldW, newW = 1, 1
	for _, p := range pairs {
		if p.leftOldNum > 0 {
			if w := len(strconv.Itoa(p.leftOldNum)); w > oldW {
				oldW = w
			}
		}
		if p.rightNewNum > 0 {
			if w := len(strconv.Itoa(p.rightNewNum)); w > newW {
				newW = w
			}
		}
	}
	return oldW, newW
}
