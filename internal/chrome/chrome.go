// Package chrome provides the ChromeTheme interface and its implementations for
// decorating diff output with styled file headers.
//
// Two themes are available:
//   - DriftTheme: the default drift style — a slate-blue ▸ chevron before the
//     filename and a full-width ─ rule below it.
//   - DeltaTheme: a box-decorated style inspired by delta's file-decoration-style —
//     wraps the filename in a ┌─ ... ─┐ / └───────────┘ Unicode box.
//
// Callers select a theme via [ParseChromeTheme] and pass it to
// [ChromeTheme.RenderFileHeader] at render time.
package chrome

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
)

// ChromeTheme renders the decorative "chrome" around diff sections
// (file headers, separator rules). Implementations: [DriftTheme], [DeltaTheme].
type ChromeTheme interface {
	// RenderFileHeader returns the fully formatted file header string
	// (including trailing newlines) for the given filename.
	// noColor disables all ANSI codes. termWidth is the terminal column count;
	// 0 means use 80 as default.
	RenderFileHeader(name string, noColor bool, termWidth int) string

	// Name returns the theme identifier (e.g. "drift", "delta").
	Name() string
}

const fallbackWidth = 80

func resolveWidth(termWidth int) int {
	if termWidth <= 0 {
		return fallbackWidth
	}
	return termWidth
}

// DriftTheme is the default drift chrome: a slate-blue ▸ chevron before the
// filename and a full-width ─ rule below it.
type DriftTheme struct{}

// Name returns "drift".
func (DriftTheme) Name() string { return "drift" }

// RenderFileHeader renders the drift-style file header.
//
// Styled (color) format:
//
//	▸ filename
//	────────────────────────────────────────────────────────────
//
// Plain (noColor) format:
//
//	▸ filename
//	------------------------------------------------------------
//
// A blank line follows the rule so the diff hunk below has breathing room.
func (DriftTheme) RenderFileHeader(name string, noColor bool, termWidth int) string {
	width := resolveWidth(termWidth)
	if noColor {
		return "▸ " + name + "\n" + strings.Repeat("-", width) + "\n\n"
	}
	// Accent color for the ▸ glyph — muted slate-blue (ANSI 256 #63).
	chevronStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("63")).Bold(true)
	// Filename in a muted foreground (bright white on dark / dark gray on light).
	nameStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("250"))
	// Rule in a dimmer tone so it recedes behind the filename.
	ruleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("238"))
	rule := strings.Repeat("─", width)
	return chevronStyle.Render("▸") + " " + nameStyle.Render(name) + "\n" +
		ruleStyle.Render(rule) + "\n\n"
}

// DeltaTheme is a box-decorated chrome inspired by delta's file-decoration-style.
// It wraps the filename in a ┌─ ... ─┐ / └───────────┘ Unicode box.
type DeltaTheme struct{}

// Name returns "delta".
func (DeltaTheme) Name() string { return "delta" }

// RenderFileHeader renders the delta-style box-decorated file header.
//
// Styled (color) format:
//
//	┌─ filename ─────────────────────────────────────────────────┐
//	└────────────────────────────────────────────────────────────┘
//
// Plain (noColor) format:
//
//	+-- filename ------------------------------------------------+
//	+------------------------------------------------------------+
//
// A blank line follows the box so the diff hunk below has breathing room.
func (DeltaTheme) RenderFileHeader(name string, noColor bool, termWidth int) string {
	width := resolveWidth(termWidth)
	if noColor {
		// Plain ASCII box: +-- filename --...+ / +---------...+
		inner := "-- " + name + " "
		padding := width - 2 - len(inner)
		if padding < 0 {
			padding = 0
		}
		top := "+" + inner + strings.Repeat("-", padding) + "+"
		bottom := "+" + strings.Repeat("-", width-2) + "+"
		return top + "\n" + bottom + "\n\n"
	}
	// Styled box using slate-blue (matches drift chevron color) for borders.
	boxColor := lipgloss.Color("63")   // slate blue
	nameColor := lipgloss.Color("250") // muted white
	boxStyle := lipgloss.NewStyle().Foreground(boxColor)
	nameStyle := lipgloss.NewStyle().Foreground(nameColor)

	// Build top bar: ┌─ filename ──────────────────────────────────────┐
	// visibleLabel is the unstyled label to measure its visible width.
	visibleLabel := "─ " + name + " "
	padding := width - 2 - len(visibleLabel)
	if padding < 0 {
		padding = 0
	}
	// label uses styled name inside the border run.
	label := "─ " + nameStyle.Render(name) + " "
	top := boxStyle.Render("┌" + label + strings.Repeat("─", padding) + "┐")
	bottom := boxStyle.Render("└" + strings.Repeat("─", width-2) + "┘")
	return top + "\n" + bottom + "\n\n"
}

// ParseChromeTheme maps a name string to a [ChromeTheme]. Returns an error
// for unknown names. Matching is case-insensitive.
//
// Valid names: "drift" (or ""), "delta".
func ParseChromeTheme(name string) (ChromeTheme, error) {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "drift", "":
		return DriftTheme{}, nil
	case "delta":
		return DeltaTheme{}, nil
	default:
		return nil, fmt.Errorf("drift: unknown chrome theme %q (use: drift, delta)", name)
	}
}
