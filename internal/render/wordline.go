package render

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/alecthomas/chroma/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/tylercrawford/drift/internal/edittype"
	"github.com/tylercrawford/drift/internal/worddiff"
)

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
// mutedStyle and changed spans with changedTint (neutral gutter-column background).
func highlightSegmented(content string, lexer chroma.Lexer, style *chroma.Style, formatter chroma.Formatter, segs []worddiff.Segment, muted, changedTint lipgloss.Style) string {
	if content == "" {
		return ""
	}
	if len(segs) == 0 {
		return highlightPanel(content, lexer, style, formatter)
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
		hl := highlightPanel(piece, lexer, style, formatter)
		hl = strings.ReplaceAll(hl, "\x1b[0m", "\x1b[39m")
		if seg.Changed {
			hl = changedTint.Render(hl)
		} else {
			hl = muted.Render(hl)
		}
		b.WriteString(hl)
	}
	if b.Len() == 0 {
		return highlightPanel(content, lexer, style, formatter)
	}
	return b.String()
}

// splitHighlightPair returns highlighted left/right panel strings, using word-level
// segments when paired delete/insert lines are both non-empty; otherwise full-line
// diff styling via splitApplyDiffLine.
func splitHighlightPair(cfg *RenderConfig, style *chroma.Style, pair linePair, lexer chroma.Lexer, formatter chroma.Formatter) (string, string) {
	if shouldWordDiffPair(cfg, pair.left, pair.right, pair.leftOp, pair.rightOp) {
		oldSegs, newSegs := worddiff.PairSegments(pair.left, pair.right)
		muted := mutedStyle(style, cfg.IsDark)
		lTint := gutterTintStyle(cfg.IsDark, cfg.NoColor, true)
		rTint := gutterTintStyle(cfg.IsDark, cfg.NoColor, false)
		l := highlightSegmented(pair.left, lexer, style, formatter, oldSegs, muted, lTint)
		r := highlightSegmented(pair.right, lexer, style, formatter, newSegs, muted, rTint)
		return splitApplyDiffLine(cfg, style, pair, l, r)
	}
	l := highlightPanel(pair.left, lexer, style, formatter)
	r := highlightPanel(pair.right, lexer, style, formatter)
	return splitApplyDiffLine(cfg, style, pair, l, r)
}

// unifiedHighlightPair returns highlighted bodies (no +/- prefix) for a paired
// delete/insert when word diff applies.
func unifiedHighlightPair(cfg *RenderConfig, style *chroma.Style, del, ins edittype.Line, lexer chroma.Lexer, formatter chroma.Formatter) (hlDel, hlIns string, ok bool) {
	if !shouldWordDiffPair(cfg, del.Content, ins.Content, del.Op, ins.Op) {
		return "", "", false
	}
	oldSegs, newSegs := worddiff.PairSegments(del.Content, ins.Content)
	muted := mutedStyle(style, cfg.IsDark)
	lTint := gutterTintStyle(cfg.IsDark, cfg.NoColor, true)
	rTint := gutterTintStyle(cfg.IsDark, cfg.NoColor, false)
	hlDel = highlightSegmented(del.Content, lexer, style, formatter, oldSegs, muted, lTint)
	hlIns = highlightSegmented(ins.Content, lexer, style, formatter, newSegs, muted, rTint)
	return hlDel, hlIns, true
}
