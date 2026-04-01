package render

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/alecthomas/chroma/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/tbcrawford/drift/internal/edittype"
	"github.com/tbcrawford/drift/internal/highlight"
	"github.com/tbcrawford/drift/internal/worddiff"
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

// highlightLineWithSegments tokenises line exactly once (preserving full token context),
// then renders each Chroma token split at segment boundaries:
//   - unchanged positions → lineBg background, Chroma foreground/emphasis
//   - changed positions   → wordBg background, same Chroma foreground/emphasis
//
// Because every token is classified from the full line, the foreground colour is
// identical across the changed and unchanged spans — only the background differs.
// This is the "multipass" fix for single-character changes that previously caused
// font colour differences due to out-of-context re-tokenisation.
//
// Falls back to plain formatter output when lineBg is unset.
func highlightLineWithSegments(line string, lexer chroma.Lexer, style *chroma.Style, formatter chroma.Formatter, segs []worddiff.Segment, lineBg, wordBg chroma.Colour) string {
	if line == "" {
		return ""
	}
	// Without a line background we cannot apply token-level backgrounds; use the
	// standard formatter path which handles colour-only diffs gracefully.
	if !lineBg.IsSet() || lexer == nil || style == nil {
		return highlightPanel(line, lexer, style, formatter)
	}

	iterator, err := lexer.Tokenise(nil, line)
	if err != nil {
		return highlightPanel(line, lexer, style, formatter)
	}

	var b strings.Builder
	bytePos := 0 // tracks position in original line as we consume tokens

	for tok := iterator(); tok != chroma.EOF; tok = iterator() {
		if tok.Value == "" {
			continue
		}
		entry := style.Get(tok.Type)
		tokEnd := bytePos + len(tok.Value)

		// Walk the token in pieces, splitting wherever a segment boundary falls.
		cur := bytePos
		for cur < tokEnd {
			next := segBoundaryBefore(cur, tokEnd, segs)
			piece := tok.Value[cur-bytePos : next-bytePos]

			bg := lineBg
			if wordBg.IsSet() && changedAt(cur, segs) {
				bg = wordBg
			}

			s := lipgloss.NewStyle().Background(lipgloss.Color(bg.String()))
			if entry.Colour.IsSet() {
				s = s.Foreground(lipgloss.Color(entry.Colour.String()))
			}
			if entry.Bold == chroma.Yes {
				s = s.Bold(true)
			}
			if entry.Italic == chroma.Yes {
				s = s.Italic(true)
			}
			if entry.Underline == chroma.Yes {
				s = s.Underline(true)
			}
			b.WriteString(s.Render(piece))
			cur = next
		}
		bytePos = tokEnd
	}

	if b.Len() == 0 {
		return line
	}
	return b.String()
}

// changedAt reports whether byte position pos falls inside a changed segment.
func changedAt(pos int, segs []worddiff.Segment) bool {
	for _, seg := range segs {
		if seg.Changed && pos >= seg.Start && pos < seg.End {
			return true
		}
	}
	return false
}

// segBoundaryBefore returns the nearest segment boundary strictly after pos and
// at most limit. Boundaries are the Start and End of every segment.
func segBoundaryBefore(pos, limit int, segs []worddiff.Segment) int {
	next := limit
	for _, seg := range segs {
		if seg.Start > pos && seg.Start < next {
			next = seg.Start
		}
		if seg.End > pos && seg.End < next {
			next = seg.End
		}
	}
	return next
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

// splitHighlightPair returns highlighted left/right panel strings, using character-level
// segments when paired delete/insert lines are both non-empty; otherwise per-line
// highlight with terrasort-style line backgrounds when LineDiffStyle is on.
func splitHighlightPair(cfg *RenderConfig, style *chroma.Style, pair linePair, lexer chroma.Lexer, formatter chroma.Formatter) (string, string) {
	if shouldWordDiffPair(cfg, pair.left, pair.right, pair.leftOp, pair.rightOp) {
		oldSegs, newSegs := worddiff.PairCharSegments(pair.left, pair.right)
		var delLineBg, insLineBg chroma.Colour
		if cfg.LineDiffStyle && !cfg.NoColor {
			delLineBg, _ = highlight.DiffLineStyle(style, edittype.Delete, cfg.IsDark)
			insLineBg, _ = highlight.DiffLineStyle(style, edittype.Insert, cfg.IsDark)
		}
		delWordBg := wordSpanBg(style, cfg.IsDark, cfg.NoColor, true)
		insWordBg := wordSpanBg(style, cfg.IsDark, cfg.NoColor, false)
		l := highlightLineWithSegments(pair.left, lexer, style, formatter, oldSegs, delLineBg, delWordBg)
		r := highlightLineWithSegments(pair.right, lexer, style, formatter, newSegs, insLineBg, insWordBg)
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
	hlDel = highlightLineWithSegments(del.Content, lexer, style, formatter, oldSegs, delLineBg, delWordBg)
	hlIns = highlightLineWithSegments(ins.Content, lexer, style, formatter, newSegs, insLineBg, insWordBg)
	return hlDel, hlIns, true
}
