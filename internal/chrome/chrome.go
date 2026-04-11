// Package chrome provides the ChromeTheme interface and its implementations for
// decorating diff output with styled file headers and hunk headers.
//
// Two themes are available:
//   - DriftTheme: the default drift style — a slate-blue ▸ chevron before the
//     filename and a full-width ─ rule below it.
//   - DeltaTheme: a style inspired by delta — Δ before the filename, a full-width
//     rule below, and a Unicode box wrapping the hunk header function context.
//
// Callers select a theme via [ParseChromeTheme] and pass it to
// [ChromeTheme.RenderFileHeader] and [ChromeTheme.RenderHunkHeader] at render time.
package chrome

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"charm.land/lipgloss/v2"
)

// ChromeTheme renders the decorative "chrome" around diff sections
// (file headers, hunk headers, separator rules). Implementations: [DriftTheme], [DeltaTheme].
type ChromeTheme interface {
	// RenderFileHeader returns the fully formatted file header string
	// (including trailing newlines) for the given filename.
	// noColor disables all ANSI codes. termWidth is the terminal column count;
	// 0 means use 80 as default.
	RenderFileHeader(name string, noColor bool, termWidth int) string

	// RenderHunkHeader returns the fully formatted hunk header string
	// (including trailing newline) for the given line number and code fragment.
	// When the returned string is empty, the caller falls back to the standard
	// "@@ -old,n +new,n @@ fragment" format.
	// noColor disables all ANSI codes.
	RenderHunkHeader(lineNum int, codeFragment string, noColor bool) string

	// GutterSeparators returns the strings used between and after gutter number columns.
	// middleSep is the separator between the old and new number columns (e.g. " │" or " ⋮ ").
	// rightBorder is the string appended after the new number column before line content
	// (e.g. "" or " │").
	// noColor disables ANSI styling.
	GutterSeparators(noColor bool) (middleSep, rightBorder string)

	// Name returns the theme identifier (e.g. "drift", "delta").
	Name() string
}

const fallbackWidth = 80

// driftGutterSep is the default gutter column separator (space + U+2502 BOX DRAWINGS LIGHT VERTICAL).
// This matches the gutterColumnSeparator constant in internal/render/gutter.go.
const driftGutterSep = " │"

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

// RenderHunkHeader returns "" — DriftTheme uses the standard "@@ ... @@" format.
func (DriftTheme) RenderHunkHeader(_ int, _ string, _ bool) string { return "" }

// GutterSeparators returns the DriftTheme gutter separator strings.
// middleSep is " │" (existing behavior); rightBorder is "" (no border after new column).
func (DriftTheme) GutterSeparators(_ bool) (string, string) {
	return driftGutterSep, ""
}

// DeltaTheme is a chrome inspired by delta's visual style.
// File headers use Δ + filename + full-width rule (matching DriftTheme structure).
// Hunk headers with a code fragment render a Unicode box around the function context.
type DeltaTheme struct{}

// Name returns "delta".
func (DeltaTheme) Name() string { return "delta" }

// RenderFileHeader renders the delta-style file header.
//
// Styled (color) format:
//
//	Δ filename
//	────────────────────────────────────────────────────────────
//
// Plain (noColor) format:
//
//	Δ filename
//	------------------------------------------------------------
//
// A blank line follows the rule so the diff hunk below has breathing room.
func (DeltaTheme) RenderFileHeader(name string, noColor bool, termWidth int) string {
	width := resolveWidth(termWidth)
	if noColor {
		return "Δ " + name + "\n" + strings.Repeat("-", width) + "\n\n"
	}
	// Accent color for the Δ glyph — muted slate-blue (ANSI 256 #63).
	chevronStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("63")).Bold(true)
	// Filename in a muted foreground (bright white on dark / dark gray on light).
	nameStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("250"))
	// Rule in a dimmer tone so it recedes behind the filename.
	ruleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("238"))
	rule := strings.Repeat("─", width)
	return chevronStyle.Render("Δ") + " " + nameStyle.Render(name) + "\n" +
		ruleStyle.Render(rule) + "\n\n"
}

// RenderHunkHeader renders the delta-style hunk header box. Always renders the box
// — uses "• N:" when codeFragment is empty, or "• N: fragment" when non-empty.
//
// Styled (color) format:
//
//	───────────────────────────────────┐
//	• 111: func name {                 │
//	───────────────────────────────────┘
//
// Plain (noColor) format:
//
//	-----------------------------------+
//	• 111: func name {                 |
//	-----------------------------------+
func (DeltaTheme) RenderHunkHeader(lineNum int, codeFragment string, noColor bool) string {
	var content string
	if codeFragment == "" {
		content = fmt.Sprintf("• %d:", lineNum)
	} else {
		content = fmt.Sprintf("• %d: %s", lineNum, codeFragment)
	}
	n := utf8.RuneCountInString(content)
	if noColor {
		top := strings.Repeat("-", n+1) + "+"
		middle := content + " |"
		bottom := strings.Repeat("-", n+1) + "+"
		return top + "\n" + middle + "\n" + bottom + "\n"
	}
	borderStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("63"))   // slate blue
	contentStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("250")) // muted white
	top := borderStyle.Render(strings.Repeat("─", n+1) + "┐")
	middle := contentStyle.Render(content) + borderStyle.Render(" │")
	bottom := borderStyle.Render(strings.Repeat("─", n+1) + "┘")
	return top + "\n" + middle + "\n" + bottom + "\n"
}

// GutterSeparators returns the DeltaTheme gutter separator strings.
// Delta format: middleSep = " ⋮ " (vertical ellipsis), rightBorder = " │" (light vertical).
// The Unicode characters render without ANSI codes, so they are the same for both
// color and noColor paths.
func (DeltaTheme) GutterSeparators(_ bool) (string, string) {
	return " ⋮ ", " │"
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
