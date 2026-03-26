package render

import (
	"fmt"
	"io"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/alecthomas/chroma/v2"
	"github.com/tylercrawford/drift/internal/edittype"
	"github.com/tylercrawford/drift/internal/highlight"
)

const minTermWidth = 40

type linePair struct {
	left    string
	right   string
	leftOp  edittype.Op
	rightOp edittype.Op
	// Line numbers for gutters (0 = blank cell).
	leftOldNum  int
	rightNewNum int
}

// Split writes a side-by-side split diff of result to w.
//
// Each hunk is rendered as two equal-width panels joined by gutterColumnSeparator
// (space + │ light vertical, standard TUI box stroke).
// The left panel shows old (deleted) lines; the right panel shows new (inserted)
// lines. Equal lines appear in both panels. Deleted lines with no matching
// insert (and vice versa) are paired with a blank placeholder on the opposite side.
//
// Panel width is (termWidth - 2) / 2. Lines are padded/truncated to exact panel
// width using lipgloss Style so ANSI sequences within highlighted content do not
// overflow column boundaries.
func Split(result edittype.DiffResult, w io.Writer, cfg *RenderConfig) error {
	if cfg == nil {
		cfg = &RenderConfig{}
	}
	if len(result.Hunks) == 0 {
		return nil
	}

	termWidth := cfg.TermWidth
	if termWidth == 0 {
		termWidth = 80
	}
	if termWidth < minTermWidth {
		termWidth = minTermWidth
	}

	panelWidth := (termWidth - 2) / 2
	rightPanelWidth := termWidth - 2 - panelWidth

	lexer := cfg.Lexer
	if lexer == nil {
		lexer = highlight.DetectLexer(cfg.Lang, "", "")
	}
	style := cfg.Style
	if style == nil {
		style = highlight.SelectTheme("", true)
	}
	formatter := cfg.Formatter
	if formatter == nil {
		formatter = highlight.FormatterForProfile(cfg.Profile)
	}

	for _, h := range result.Hunks {
		header := fmt.Sprintf("@@ -%d,%d +%d,%d @@", h.OldStart, h.OldLines, h.NewStart, h.NewLines)
		if _, err := fmt.Fprintln(w, header); err != nil {
			return err
		}

		pairs := pairHunkLines(h.Lines)

		var leftLines, rightLines, sepLines []string

		if !cfg.ShowLineNumbers {
			for _, pair := range pairs {
				lContent, rContent := splitHighlightPair(cfg, style, pair, lexer, formatter)
				var lBg, rBg chroma.Colour
				if cfg.LineDiffStyle && !cfg.NoColor {
					lBg, _ = highlight.DiffLineStyle(style, pair.leftOp, cfg.IsDark)
					rBg, _ = highlight.DiffLineStyle(style, pair.rightOp, cfg.IsDark)
				}
				leftLines = append(leftLines, renderPanelContent(lContent, panelWidth, lBg))
				rightLines = append(rightLines, renderPanelContent(rContent, rightPanelWidth, rBg))
				sepLines = append(sepLines, styledGutterColumnSeparator(cfg))
			}
		} else {
			oldW, newW := gutterPairWidths(pairs)
			for _, pair := range pairs {
				lContent, rContent := splitHighlightPair(cfg, style, pair, lexer, formatter)

				leftCodeW := panelWidth - oldW
				if leftCodeW < 1 {
					leftCodeW = 1
				}
				rightCodeW := rightPanelWidth - newW
				if rightCodeW < 1 {
					rightCodeW = 1
				}

				var lBg, rBg chroma.Colour
				if cfg.LineDiffStyle && !cfg.NoColor {
					lBg, _ = highlight.DiffLineStyle(style, pair.leftOp, cfg.IsDark)
					rBg, _ = highlight.DiffLineStyle(style, pair.rightOp, cfg.IsDark)
				}

				leftG := GutterNumberRender(gutterStyleForCell(style, cfg.IsDark, cfg.NoColor, true, pair.leftOp), oldW, pair.leftOldNum)
				rightG := GutterNumberRender(gutterStyleForCell(style, cfg.IsDark, cfg.NoColor, false, pair.rightOp), newW, pair.rightNewNum)

				leftLines = append(leftLines, lipgloss.JoinHorizontal(lipgloss.Top, leftG, renderPanelContent(lContent, leftCodeW, lBg)))
				rightLines = append(rightLines, lipgloss.JoinHorizontal(lipgloss.Top, rightG, renderPanelContent(rContent, rightCodeW, rBg)))
				sepLines = append(sepLines, styledGutterColumnSeparator(cfg))
			}
		}

		leftBlock := strings.Join(leftLines, "\n")
		rightBlock := strings.Join(rightLines, "\n")
		sepBlock := strings.Join(sepLines, "\n")

		row := lipgloss.JoinHorizontal(lipgloss.Top, leftBlock, sepBlock, rightBlock)
		if _, err := fmt.Fprintln(w, row); err != nil {
			return err
		}
	}

	return nil
}

// pairHunkLines pairs Delete and Insert lines for side-by-side display.
func pairHunkLines(lines []edittype.Line) []linePair {
	var pairs []linePair
	i := 0
	for i < len(lines) {
		line := lines[i]
		if line.Op == edittype.Equal {
			pairs = append(pairs, linePair{
				left:        line.Content,
				right:       line.Content,
				leftOp:      edittype.Equal,
				rightOp:     edittype.Equal,
				leftOldNum:  line.OldNum,
				rightNewNum: line.NewNum,
			})
			i++
			continue
		}

		var deletes, inserts []edittype.Line
		for i < len(lines) && lines[i].Op != edittype.Equal {
			if lines[i].Op == edittype.Delete {
				deletes = append(deletes, lines[i])
			} else {
				inserts = append(inserts, lines[i])
			}
			i++
		}

		maxLen := len(deletes)
		if len(inserts) > maxLen {
			maxLen = len(inserts)
		}
		for j := 0; j < maxLen; j++ {
			var pair linePair
			if j < len(deletes) {
				pair.left = deletes[j].Content
				pair.leftOp = edittype.Delete
				pair.leftOldNum = deletes[j].OldNum
			} else {
				pair.leftOp = edittype.Equal
			}
			if j < len(inserts) {
				pair.right = inserts[j].Content
				pair.rightOp = edittype.Insert
				pair.rightNewNum = inserts[j].NewNum
			} else {
				pair.rightOp = edittype.Equal
			}
			pairs = append(pairs, pair)
		}
	}
	return pairs
}

// highlightSplitPanel highlights one split panel line. For delete (left) / insert (right)
// with LineDiffStyle, uses terrasort-style per-token lipgloss with the line background.
func highlightSplitPanel(cfg *RenderConfig, style *chroma.Style, pair linePair, leftSide bool, content string, lexer chroma.Lexer, formatter chroma.Formatter) string {
	var op edittype.Op
	if leftSide {
		op = pair.leftOp
	} else {
		op = pair.rightOp
	}
	if cfg.LineDiffStyle && !cfg.NoColor {
		if bg, ok := highlight.DiffLineStyle(style, op, cfg.IsDark); ok {
			h, err := highlight.HighlightLineWithLineBackground(content, lexer, style, bg)
			if err == nil {
				return h
			}
		}
	}
	return highlightPanel(content, lexer, style, formatter)
}

// highlightPanel highlights a line for panel display. Fails open to plain text.
func highlightPanel(content string, lexer chroma.Lexer, style *chroma.Style, formatter chroma.Formatter) string {
	if content == "" {
		return ""
	}
	highlighted, err := highlight.HighlightLine(content, lexer, style, formatter)
	if err != nil {
		return content
	}
	return highlighted
}
