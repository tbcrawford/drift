// Package render implements drift's diff output renderers.
package render

import (
	"fmt"
	"io"

	"github.com/alecthomas/chroma/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/tbcrawford/drift/internal/edittype"
	"github.com/tbcrawford/drift/internal/highlight"
)

// RenderConfig holds the resolved rendering configuration for a single
// Render call. It is populated by the public drift.Render() function
// and passed down into the internal renderer.
type RenderConfig struct {
	// OldName and NewName are the file header labels for --- and +++ lines.
	// When empty, they default to "a/input" and "b/input".
	OldName string
	NewName string

	// Lang is the explicit Chroma language override (e.g., "go"). Empty means
	// auto-detect from filename/content.
	Lang string

	// Lexer is the pre-resolved Chroma lexer. If nil, DetectLexer is called
	// using Lang.
	Lexer chroma.Lexer

	// Style is the pre-resolved Chroma style. If nil, SelectTheme is called.
	Style *chroma.Style

	// Formatter is the pre-resolved Chroma formatter. If nil, FormatterForProfile
	// is called with Profile.
	Formatter chroma.Formatter

	// Profile is the detected terminal color profile.
	Profile colorprofile.Profile

	// NoColor disables all ANSI output when true.
	NoColor bool

	// TermWidth is the resolved terminal width in columns for split rendering.
	// Zero means "use 80 columns" — the Split renderer applies this default.
	TermWidth int

	// ShowLineNumbers adds old/new gutter columns before the unified diff prefix.
	// When false, output matches pre–line-number behavior (prefix + highlighted only).
	// Callers should set this explicitly; the zero value is false.
	ShowLineNumbers bool

	// IsDark is the terminal background hint used for gutter styling (with color).
	IsDark bool

	// LineDiffStyle applies theme-derived full-line backgrounds on added/removed
	// code (prefix + highlighted content). When false, only Chroma syntax colors apply.
	LineDiffStyle bool

	// WordDiff enables word-level intra-line highlights for consecutive delete/insert
	// lines (unified) and paired split rows. When false, only line-level styling applies.
	WordDiff bool

	// GutterCache is a pre-computed cache of all gutter cell styles for this render call.
	// When non-nil and ShowLineNumbers is true, renderers use cache lookups instead of
	// constructing a new lipgloss.Style per line. Set automatically by Unified and Split
	// when ShowLineNumbers is enabled.
	GutterCache *GutterStyleCache
}

// Unified writes a Git-compatible unified diff of result to w.
//
// Each hunk is preceded by a "@@ -OldStart,OldLines +NewStart,NewLines @@" header.
// Lines are prefixed with "+", "-", or " " (space) for Insert, Delete, and Equal
// respectively. Syntax highlighting is applied per-line when cfg.Formatter is
// not NoOp.
//
// When LineDiffStyle is enabled, insert/delete lines use per-token lipgloss rendering
// with a line-level background (terrasort-style), not Chroma's TTY formatter with a
// background prefix.
func Unified(result edittype.DiffResult, w io.Writer, cfg *RenderConfig) error {
	if cfg == nil {
		cfg = &RenderConfig{}
	}

	// Resolve defaults for optional fields.
	oldName := cfg.OldName
	if oldName == "" {
		oldName = "a/input"
	}
	newName := cfg.NewName
	if newName == "" {
		newName = "b/input"
	}

	lexer := cfg.Lexer
	if lexer == nil {
		lexer = highlight.DetectLexer(cfg.Lang, "", "")
	}

	style := cfg.Style
	if style == nil {
		style = highlight.SelectTheme("", cfg.IsDark) // fallback; caller should always set cfg.Style
	}

	formatter := cfg.Formatter
	if formatter == nil {
		formatter = highlight.FormatterForProfile(cfg.Profile)
	}

	// Write file headers only when there are hunks to show.
	if len(result.Hunks) == 0 {
		return nil
	}

	if _, err := fmt.Fprintf(w, "--- %s\n+++ %s\n", oldName, newName); err != nil {
		return err
	}

	gutterSep := styledGutterColumnSeparator(cfg)

	// Pre-build gutter style cache once for the whole render call.
	// This eliminates per-line lipgloss.NewStyle() allocations in the hot render loop.
	if cfg.ShowLineNumbers && cfg.GutterCache == nil {
		cfg.GutterCache = NewGutterStyleCache(style, cfg.IsDark, cfg.NoColor)
	}

	termWidth := cfg.TermWidth
	if termWidth == 0 {
		termWidth = 80
	}

	for _, h := range result.Hunks {
		// Hunk header: @@ -OldStart,OldLines +NewStart,NewLines @@ [CodeFragment]
		var header string
		if h.CodeFragment != "" {
			header = fmt.Sprintf("@@ -%d,%d +%d,%d @@ %s\n",
				h.OldStart, h.OldLines, h.NewStart, h.NewLines, h.CodeFragment)
		} else {
			header = fmt.Sprintf("@@ -%d,%d +%d,%d @@\n",
				h.OldStart, h.OldLines, h.NewStart, h.NewLines)
		}
		if _, err := fmt.Fprint(w, header); err != nil {
			return err
		}

		var oldW, newW int
		if cfg.ShowLineNumbers {
			oldW, newW = gutterWidths(h.Lines)
		}

		// contentW is the width available for highlighted content (prefix already excluded).
		// We use it to extend line backgrounds to the full terminal width.
		const gutterSepWidth = 2  // " │"
		contentW := termWidth - 1 // -1 for the +/-/space prefix
		if cfg.ShowLineNumbers {
			contentW -= oldW + gutterSepWidth + newW
		}
		if contentW < 0 {
			contentW = 0
		}

		lines := h.Lines
		for i := 0; i < len(lines); i++ {
			line := lines[i]

			if i+1 < len(lines) && line.Op == edittype.Delete && lines[i+1].Op == edittype.Insert {
				if hlDel, hlIns, ok := unifiedHighlightPair(cfg, style, line, lines[i+1], lexer, formatter); ok {
					var delBg, insBg chroma.Colour
					if cfg.LineDiffStyle && !cfg.NoColor {
						delBg, _ = highlight.DiffLineStyle(style, edittype.Delete, cfg.IsDark)
						insBg, _ = highlight.DiffLineStyle(style, edittype.Insert, cfg.IsDark)
					}
					prefixDel := styledDiffPrefix(edittype.Delete, delBg)
					prefixIns := styledDiffPrefix(edittype.Insert, insBg)
					hlDel = padLineBackground(hlDel, contentW, delBg)
					hlIns = padLineBackground(hlIns, contentW, insBg)
					codeDel := prefixDel + hlDel
					codeIns := prefixIns + hlIns
					ins := lines[i+1]
					if !cfg.ShowLineNumbers {
						if _, err := fmt.Fprintf(w, "%s\n", codeDel); err != nil {
							return err
						}
						if _, err := fmt.Fprintf(w, "%s\n", codeIns); err != nil {
							return err
						}
					} else {
						goLeft := GutterNumberRender(cfg.GutterCache.Get(true, line.Op), oldW, line.OldNum)
						goRight := GutterNumberRender(cfg.GutterCache.Get(false, line.Op), newW, line.NewNum)
						if _, err := fmt.Fprintf(w, "%s%s%s%s\n", goLeft, gutterSep, goRight, codeDel); err != nil {
							return err
						}
						goLeft2 := GutterNumberRender(cfg.GutterCache.Get(true, ins.Op), oldW, ins.OldNum)
						goRight2 := GutterNumberRender(cfg.GutterCache.Get(false, ins.Op), newW, ins.NewNum)
						if _, err := fmt.Fprintf(w, "%s%s%s%s\n", goLeft2, gutterSep, goRight2, codeIns); err != nil {
							return err
						}
					}
					i++
					continue
				}
			}

			var lineBg chroma.Colour
			if cfg.LineDiffStyle && !cfg.NoColor {
				lineBg, _ = highlight.DiffLineStyle(style, line.Op, cfg.IsDark)
			}
			prefix := styledDiffPrefix(line.Op, lineBg)
			highlighted := highlightUnifiedLine(cfg, style, line, lexer, formatter)
			highlighted = padLineBackground(highlighted, contentW, lineBg)
			code := prefix + highlighted

			if !cfg.ShowLineNumbers {
				if _, err := fmt.Fprintf(w, "%s\n", code); err != nil {
					return err
				}
				continue
			}

			goLeft := GutterNumberRender(cfg.GutterCache.Get(true, line.Op), oldW, line.OldNum)
			goRight := GutterNumberRender(cfg.GutterCache.Get(false, line.Op), newW, line.NewNum)
			if _, err := fmt.Fprintf(w, "%s%s%s%s\n", goLeft, gutterSep, goRight, code); err != nil {
				return err
			}
		}
	}

	return nil
}

// linePrefix returns the unified diff prefix character for the given Op.
func linePrefix(op edittype.Op) string {
	switch op {
	case edittype.Insert:
		return "+"
	case edittype.Delete:
		return "-"
	default: // Equal
		return " "
	}
}

// highlightUnifiedLine applies syntax highlighting. For diff lines with LineDiffStyle,
// uses HighlightLineWithLineBackground (terrasort token loop); otherwise Chroma's formatter.
func highlightUnifiedLine(cfg *RenderConfig, style *chroma.Style, line edittype.Line, lexer chroma.Lexer, formatter chroma.Formatter) string {
	if cfg.LineDiffStyle && !cfg.NoColor {
		if bg, ok := highlight.DiffLineStyle(style, line.Op, cfg.IsDark); ok {
			h, err := highlight.HighlightLineWithLineBackground(line.Content, lexer, style, bg)
			if err == nil {
				return h
			}
		}
	}
	h, err := highlight.HighlightLine(line.Content, lexer, style, formatter)
	if err != nil {
		return line.Content
	}
	return h
}
