// Package render implements drift's diff output renderers.
package render

import (
	"fmt"
	"io"

	"github.com/alecthomas/chroma/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/tylercrawford/drift/internal/edittype"
	"github.com/tylercrawford/drift/internal/highlight"
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
}

// Unified writes a Git-compatible unified diff of result to w.
//
// Each hunk is preceded by a "@@ -OldStart,OldLines +NewStart,NewLines @@" header.
// Lines are prefixed with "+", "-", or " " (space) for Insert, Delete, and Equal
// respectively. Syntax highlighting is applied per-line when cfg.Formatter is
// not NoOp.
//
// The output is self-contained: ANSI reset sequences are emitted by Chroma at
// line boundaries, so each line is independently renderable.
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
		style = highlight.SelectTheme("", true) // dark default; caller sets via cfg.Style
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

	for _, h := range result.Hunks {
		// Hunk header: @@ -OldStart,OldLines +NewStart,NewLines @@
		header := fmt.Sprintf("@@ -%d,%d +%d,%d @@\n",
			h.OldStart, h.OldLines, h.NewStart, h.NewLines)
		if _, err := fmt.Fprint(w, header); err != nil {
			return err
		}

		for _, line := range h.Lines {
			prefix := linePrefix(line.Op)

			highlighted, err := highlight.HighlightLine(line.Content, lexer, style, formatter)
			if err != nil {
				// Fail-open: use plain content on highlight error.
				highlighted = line.Content
			}

			if _, err := fmt.Fprintf(w, "%s%s\n", prefix, highlighted); err != nil {
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
