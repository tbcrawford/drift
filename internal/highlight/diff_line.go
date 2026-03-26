package highlight

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/alecthomas/chroma/v2"
	"github.com/tylercrawford/drift/internal/edittype"
)

// DiffLineStyle returns a Lip Gloss style that applies only a full-line background
// for insert/delete lines. Colours come from [DiffLineMutedBackgroundColour] — the
// muted semantic plane — so brighter [WordSpanBackgroundColour] spans can sit on top
// in word-diff mode.
//
// The second return is false when op is Equal or when no usable background can be derived.
func DiffLineStyle(style *chroma.Style, op edittype.Op, isDark bool) (lipgloss.Style, bool) {
	if style == nil || op == edittype.Equal {
		return lipgloss.NewStyle(), false
	}

	var del bool
	switch op {
	case edittype.Delete:
		del = true
	case edittype.Insert:
		del = false
	default:
		return lipgloss.NewStyle(), false
	}

	bg := DiffLineMutedBackgroundColour(style, isDark, del)
	if !bg.IsSet() {
		return lipgloss.NewStyle(), false
	}

	return lipgloss.NewStyle().Background(lipgloss.Color(bg.String())), true
}

// ApplyDiffLineStyle wraps highlighted code (ANSI from Chroma) with a theme-derived
// line background. Chroma's TTY formatters emit \x1b[0m between tokens, which would
// clear a line-level background; those resets are downgraded to foreground-only resets
// (\x1b[39m) so the outer background persists across tokens.
func ApplyDiffLineStyle(st lipgloss.Style, highlighted string) string {
	if highlighted == "" {
		return highlighted
	}
	// Downgrade full resets so line background survives token boundaries.
	s := strings.ReplaceAll(highlighted, "\x1b[0m", "\x1b[39m")
	return st.Render(s)
}
