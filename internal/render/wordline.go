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

func shouldWordDiffPair(cfg *RenderConfig, left, right string, leftOp, rightOp edittype.Op) bool {
	if cfg == nil || !cfg.WordDiff || cfg.NoColor {
		return false
	}
	// Only run word diff when terminal color output is active.
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

// highlightSegmented highlights all segments with lineBg so syntax foreground colours
// are consistent throughout the line, then replaces only the background ANSI escape on
// changed spans with wordBg. This preserves font colour and only changes the background
// for changed characters — the same visual contract as GitHub PR intra-line highlights.
func highlightSegmented(content string, lexer chroma.Lexer, style *chroma.Style, formatter chroma.Formatter, segs []worddiff.Segment, lineBg, wordBg chroma.Colour) string {
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
		if seg.Changed && wordBg.IsSet() && lineBg.IsSet() {
			// Swap only the background escape — foreground/emphasis unchanged.
			hl = highlight.ReplaceAnsiBackground(hl, lineBg, wordBg)
		}
		b.WriteString(hl)
	}
	if b.Len() == 0 {
		return highlightPiece(content, lexer, style, formatter, lineBg)
	}
	return b.String()
}

// highlightPiece highlights a substring; when bg is set, uses terrasort-style per-token
// lipgloss with that background so the colour extends through token boundaries.
func highlightPiece(content string, lexer chroma.Lexer, style *chroma.Style, formatter chroma.Formatter, bg chroma.Colour) string {
	if bg.IsSet() {
		h, err := highlight.HighlightLineWithLineBackground(content, lexer, style, bg)
		if err == nil {
			return h
		}
	}
	return highlightPanel(content, lexer, style, formatter)
}

// wordSpanBg returns the word-diff background colour for the given side: brighter
// semantic red (delete) or green (insert) than the full-line plane.
// Matches what the gutter uses via highlight.WordSpanBackgroundColour.
func wordSpanBg(style *chroma.Style, isDark, noColor, del bool) chroma.Colour {
	if noColor || style == nil {
		return chroma.Colour(0)
	}
	return highlight.WordSpanBackgroundColour(style, isDark, del)
}

// splitHighlightPair returns highlighted left/right panel strings, using word-level
// segments when paired delete/insert lines are both non-empty; otherwise per-line
// highlight with terrasort-style line backgrounds when LineDiffStyle is on.
func splitHighlightPair(cfg *RenderConfig, style *chroma.Style, pair linePair, lexer chroma.Lexer, formatter chroma.Formatter) (string, string) {
	if shouldWordDiffPair(cfg, pair.left, pair.right, pair.leftOp, pair.rightOp) {
		oldSegs, newSegs := worddiff.PairCharSegments(pair.left, pair.right)
		var delLineBg, insLineBg, delWordBg, insWordBg chroma.Colour
		if cfg.LineDiffStyle && !cfg.NoColor {
			delLineBg, _ = highlight.DiffLineStyle(style, edittype.Delete, cfg.IsDark)
			insLineBg, _ = highlight.DiffLineStyle(style, edittype.Insert, cfg.IsDark)
		}
		delWordBg = wordSpanBg(style, cfg.IsDark, cfg.NoColor, true)
		insWordBg = wordSpanBg(style, cfg.IsDark, cfg.NoColor, false)
		l := highlightSegmented(pair.left, lexer, style, formatter, oldSegs, delLineBg, delWordBg)
		r := highlightSegmented(pair.right, lexer, style, formatter, newSegs, insLineBg, insWordBg)
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
	oldSegs, newSegs := worddiff.PairCharSegments(del.Content, ins.Content)
	var delLineBg, insLineBg chroma.Colour
	if cfg.LineDiffStyle && !cfg.NoColor {
		delLineBg, _ = highlight.DiffLineStyle(style, edittype.Delete, cfg.IsDark)
		insLineBg, _ = highlight.DiffLineStyle(style, edittype.Insert, cfg.IsDark)
	}
	delWordBg := wordSpanBg(style, cfg.IsDark, cfg.NoColor, true)
	insWordBg := wordSpanBg(style, cfg.IsDark, cfg.NoColor, false)
	hlDel = highlightSegmented(del.Content, lexer, style, formatter, oldSegs, delLineBg, delWordBg)
	hlIns = highlightSegmented(ins.Content, lexer, style, formatter, newSegs, insLineBg, insWordBg)
	return hlDel, hlIns, true
}
