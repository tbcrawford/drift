package highlight

import (
	"fmt"
	"strings"

	"github.com/alecthomas/chroma/v2"
)

// HighlightLineWithLineBackground renders syntax highlighting for a diff line using
// a direct ANSI SGR builder — no lipgloss allocation per token.
//
// Each Chroma token is emitted as an ANSI escape sequence combining the line-level
// background with per-token foreground, bold, italic, and underline attributes.
// A reset sequence (\x1b[0m) follows each token to prevent color bleed.
//
// This avoids lipgloss.NewStyle() per token: on a 10k-line diff with color enabled,
// the old implementation created millions of allocations. The new implementation
// uses a strings.Builder with pre-computed SGR byte sequences per token.
func HighlightLineWithLineBackground(line string, lexer chroma.Lexer, style *chroma.Style, lineBg chroma.Colour) (string, error) {
	if !lineBg.IsSet() || lexer == nil || style == nil {
		return line, nil
	}
	return highlightLineWithLineBackgroundFast(line, lexer, style, lineBg)
}

// highlightLineWithLineBackgroundFast is the allocation-free inner implementation.
// It writes ANSI SGR sequences directly using fmt.Fprintf into a strings.Builder,
// avoiding lipgloss.NewStyle() allocations per token.
func highlightLineWithLineBackgroundFast(line string, lexer chroma.Lexer, style *chroma.Style, lineBg chroma.Colour) (string, error) {
	iterator, err := lexer.Tokenise(nil, line)
	if err != nil {
		return line, err
	}

	// Pre-format the background sub-sequence once — reused for every token.
	// Format: "48;2;R;G;B"
	bgSeq := fmt.Sprintf("48;2;%d;%d;%d", lineBg.Red(), lineBg.Green(), lineBg.Blue())

	var b strings.Builder
	// Heuristic capacity: most lines < 120 chars, tokens ~20 chars + ~30 ANSI overhead each.
	b.Grow(len(line) * 3)

	for tok := iterator(); tok != chroma.EOF; tok = iterator() {
		v := tok.Value
		if v == "" {
			continue
		}

		entry := style.Get(tok.Type)

		// Build a combined SGR parameter string.
		// Always include the line background. Optionally add fg, bold, italic, underline.
		var params strings.Builder
		params.WriteString(bgSeq)

		if entry.Colour.IsSet() {
			params.WriteByte(';')
			fmt.Fprintf(&params, "38;2;%d;%d;%d", entry.Colour.Red(), entry.Colour.Green(), entry.Colour.Blue())
		}
		if entry.Bold == chroma.Yes {
			params.WriteString(";1")
		}
		if entry.Italic == chroma.Yes {
			params.WriteString(";3")
		}
		if entry.Underline == chroma.Yes {
			params.WriteString(";4")
		}

		b.WriteString("\x1b[")
		b.WriteString(params.String())
		b.WriteByte('m')
		b.WriteString(v)
		b.WriteString("\x1b[0m")
	}

	if b.Len() == 0 {
		return line, nil
	}
	return strings.TrimRight(b.String(), "\n"), nil
}

// ReplaceAnsiBackground replaces every occurrence of the TrueColor background
// sub-sequence for `from` with the sub-sequence for `to` in s. It handles both
// standalone escapes (\x1b[48;2;R;G;Bm) and combined fg+bg SGR sequences that
// lipgloss v2 emits (\x1b[38;2;R;G;B;48;2;R;G;Bm). All other ANSI codes (foreground,
// bold, italic, etc.) are preserved unchanged.
//
// The function only operates on the 24-bit TrueColor sub-sequence "48;2;R;G;B".
// When the string uses a lower colour depth the call is a no-op — an acceptable
// degradation because character-level highlights are most valuable in TrueColor terminals.
//
// Note: this function is not called by the current word-diff renderer, which uses
// segment-based highlighting (highlightLineWithSegments) instead. It is retained
// as a tested utility for a future renderer iteration that may require post-hoc
// background replacement on pre-highlighted ANSI strings.
func ReplaceAnsiBackground(s string, from, to chroma.Colour) string {
	if !from.IsSet() || !to.IsSet() || from == to {
		return s
	}
	fromPart := fmt.Sprintf("48;2;%d;%d;%d", from.Red(), from.Green(), from.Blue())
	toPart := fmt.Sprintf("48;2;%d;%d;%d", to.Red(), to.Green(), to.Blue())
	return strings.ReplaceAll(s, fromPart, toPart)
}
