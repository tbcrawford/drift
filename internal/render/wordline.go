package render

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/alecthomas/chroma/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/tylercrawford/drift/internal/edittype"
	"github.com/tylercrawford/drift/internal/highlight"
	"github.com/tylercrawford/drift/internal/worddiff"
)

// styledDiffPrefix returns the +/-/space prefix character styled with bg when bg is set.
func styledDiffPrefix(op edittype.Op, bg chroma.Colour) string {
	p := linePrefix(op)
	if bg.IsSet() {
		return lipgloss.NewStyle().Background(lipgloss.Color(bg.String())).Render(p)
	}
	return p
}

// padLineBackground appends bg-colored spaces to content to fill width.
// This is needed because chroma token resets (\033[0m) prevent lipgloss Width()
// padding from carrying the background — the trailing spaces end up unstyled.
// If bg is not set, width <= 0, or content already fills width, returns content unchanged.
func padLineBackground(content string, width int, bg chroma.Colour) string {
	if !bg.IsSet() || width <= 0 {
		return content
	}
	visible := lipgloss.Width(content)
	padding := width - visible
	if padding <= 0 {
		return content
	}
	return content + lipgloss.NewStyle().Background(lipgloss.Color(bg.String())).Render(strings.Repeat(" ", padding))
}

// renderPanelContent pads content to panelWidth, extending bg for diff lines or
// using plain lipgloss Width() for context lines.
func renderPanelContent(content string, panelWidth int, bg chroma.Colour) string {
	if bg.IsSet() {
		return padLineBackground(content, panelWidth, bg)
	}
	return lipgloss.NewStyle().Width(panelWidth).Render(content)
}

// mutedStyle returns a foreground-only style for unchanged intra-line segments.
func mutedStyle(style *chroma.Style, isDark bool) lipgloss.Style {
	if style == nil {
		return lipgloss.NewStyle()
	}
	e := style.Get(chroma.Comment)
	if e.Colour.IsSet() {
		return lipgloss.NewStyle().Foreground(lipgloss.Color(e.Colour.String()))
	}
	if isDark {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#8b8b8b"))
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#57606a"))
}

func shouldWordDiffPair(cfg *RenderConfig, left, right string, leftOp, rightOp edittype.Op) bool {
	if cfg == nil || !cfg.WordDiff || cfg.NoColor {
		return false
	}
	// Chroma formatter is not comparable (FormatterFunc); use profile so we
	// only run word diff when terminal color output is active.
	switch cfg.Profile {
	case colorprofile.TrueColor, colorprofile.ANSI256, colorprofile.ANSI:
	default:
		return false
	}
	if left == "" || right == "" {
		return false
	}
	return leftOp == edittype.Delete && rightOp == edittype.Insert
}

// highlightSegmented applies Chroma per segment and wraps unchanged spans with
// mutedStyle and changed spans with semantic word-span background (brighter red/green).
// When lineBg is set, the base highlight uses terrasort-style per-token lipgloss with
// that line background so the full-line diff tint survives token boundaries.
func highlightSegmented(content string, lexer chroma.Lexer, style *chroma.Style, formatter chroma.Formatter, segs []worddiff.Segment, muted, changedWord lipgloss.Style, lineBg chroma.Colour) string {
	if content == "" {
		return ""
	}
	if len(segs) == 0 {
		return highlightPiece(content, lexer, style, formatter, lineBg)
	}
	var b strings.Builder
	for _, seg := range segs {
		if seg.Start < 0 || seg.End > len(content) || seg.Start >= seg.End {
			continue
		}
		piece := content[seg.Start:seg.End]
		if piece == "" {
			continue
		}
		hl := highlightPiece(piece, lexer, style, formatter, lineBg)
		hl = strings.ReplaceAll(hl, "\x1b[0m", "\x1b[39m")
		if seg.Changed {
			hl = changedWord.Render(hl)
		} else {
			hl = muted.Render(hl)
		}
		b.WriteString(hl)
	}
	if b.Len() == 0 {
		return highlightPiece(content, lexer, style, formatter, lineBg)
	}
	return b.String()
}

// highlightPiece highlights a substring; when lineBg is set, uses terrasort-style per-token
// lipgloss with that line background.
func highlightPiece(content string, lexer chroma.Lexer, style *chroma.Style, formatter chroma.Formatter, lineBg chroma.Colour) string {
	if lineBg.IsSet() {
		h, err := highlight.HighlightLineWithLineBackground(content, lexer, style, lineBg)
		if err == nil {
			return h
		}
	}
	return highlightPanel(content, lexer, style, formatter)
}

// wordSpanStyle returns a background-only style for intra-line changed segments:
// brighter semantic red (delete) or green (insert) than the muted full-line plane.
func wordSpanStyle(style *chroma.Style, isDark, noColor, del bool) lipgloss.Style {
	if noColor || style == nil {
		return lipgloss.NewStyle()
	}
	c := highlight.WordSpanBackgroundColour(style, isDark, del)
	if !c.IsSet() {
		return lipgloss.NewStyle()
	}
	return lipgloss.NewStyle().Background(lipgloss.Color(c.String()))
}

// splitHighlightPair returns highlighted left/right panel strings, using word-level
// segments when paired delete/insert lines are both non-empty; otherwise per-line
// highlight with terrasort-style line backgrounds when LineDiffStyle is on.
func splitHighlightPair(cfg *RenderConfig, style *chroma.Style, pair linePair, lexer chroma.Lexer, formatter chroma.Formatter) (string, string) {
	if shouldWordDiffPair(cfg, pair.left, pair.right, pair.leftOp, pair.rightOp) {
		oldSegs, newSegs := worddiff.PairSegments(pair.left, pair.right)
		muted := mutedStyle(style, cfg.IsDark)
		lWord := wordSpanStyle(style, cfg.IsDark, cfg.NoColor, true)
		rWord := wordSpanStyle(style, cfg.IsDark, cfg.NoColor, false)
		var delBg, insBg chroma.Colour
		if cfg.LineDiffStyle && !cfg.NoColor {
			delBg, _ = highlight.DiffLineStyle(style, edittype.Delete, cfg.IsDark)
			insBg, _ = highlight.DiffLineStyle(style, edittype.Insert, cfg.IsDark)
		}
		l := highlightSegmented(pair.left, lexer, style, formatter, oldSegs, muted, lWord, delBg)
		r := highlightSegmented(pair.right, lexer, style, formatter, newSegs, muted, rWord, insBg)
		return l, r
	}
	l := highlightSplitPanel(cfg, style, pair, true, pair.left, lexer, formatter)
	r := highlightSplitPanel(cfg, style, pair, false, pair.right, lexer, formatter)
	return l, r
}

// unifiedHighlightPair returns highlighted bodies (no +/- prefix) for a paired
// delete/insert when word diff applies.
func unifiedHighlightPair(cfg *RenderConfig, style *chroma.Style, del, ins edittype.Line, lexer chroma.Lexer, formatter chroma.Formatter) (hlDel, hlIns string, ok bool) {
	if !shouldWordDiffPair(cfg, del.Content, ins.Content, del.Op, ins.Op) {
		return "", "", false
	}
	oldSegs, newSegs := worddiff.PairSegments(del.Content, ins.Content)
	muted := mutedStyle(style, cfg.IsDark)
	lWord := wordSpanStyle(style, cfg.IsDark, cfg.NoColor, true)
	rWord := wordSpanStyle(style, cfg.IsDark, cfg.NoColor, false)
	var delBg, insBg chroma.Colour
	if cfg.LineDiffStyle && !cfg.NoColor {
		delBg, _ = highlight.DiffLineStyle(style, edittype.Delete, cfg.IsDark)
		insBg, _ = highlight.DiffLineStyle(style, edittype.Insert, cfg.IsDark)
	}
	hlDel = highlightSegmented(del.Content, lexer, style, formatter, oldSegs, muted, lWord, delBg)
	hlIns = highlightSegmented(ins.Content, lexer, style, formatter, newSegs, muted, rWord, insBg)
	return hlDel, hlIns, true
}
