package render

import (
	"io"
	"os"
	"strconv"

	"github.com/charmbracelet/x/term"
)

// TerminalWidth returns the terminal column count for the given writer.
//
// Resolution order:
//  1. If w is an *os.File and the file descriptor is a TTY, query the terminal
//     size via term.GetSize. Returns the width when > 0.
//  2. Read the COLUMNS environment variable (set by shells and respected by diff
//     tools such as delta and bat). Parses as a decimal integer; ignores
//     non-positive or non-numeric values.
//  3. Default to 80 columns — the industry-standard safe width for piped output.
//
// TerminalWidth never panics. It is safe to call with any io.Writer, including
// nil (returns 80).
//
// Note: runewidth.EastAsianWidth is intentionally left at its default (false).
// Lip Gloss v2 uses rivo/uniseg for grapheme-cluster-aware width measurement,
// which handles combining characters, ZWJ sequences, and emoji correctly without
// treating ambiguous-width characters as double-width.
func TerminalWidth(w io.Writer) int {
	if f, ok := w.(*os.File); ok {
		if width, _, err := term.GetSize(f.Fd()); err == nil && width > 0 {
			return width
		}
	}

	if cols := os.Getenv("COLUMNS"); cols != "" {
		if n, err := strconv.Atoi(cols); err == nil && n > 0 {
			return n
		}
	}

	return 80
}
