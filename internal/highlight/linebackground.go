package highlight

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/alecthomas/chroma/v2"
)

// HighlightLineWithLineBackground renders syntax highlighting the same way terrasort does
// for diff lines (see terrasort internal/highlight/highlight.go Highlight): each Chroma
// token is emitted as a separate lipgloss span with the line-level background applied on
// every token, plus foreground (and emphasis) from the Chroma style entry.
//
// This does not use Chroma's TTY formatter: those formatters end each token with \x1b[0m,
// which resets both foreground and background and defeats any line-level background applied
// only as a prefix or outer lipgloss wrap.
func HighlightLineWithLineBackground(line string, lexer chroma.Lexer, style *chroma.Style, lineBg chroma.Colour) (string, error) {
	if !lineBg.IsSet() || lexer == nil || style == nil {
		return line, nil
	}
	iterator, err := lexer.Tokenise(nil, line)
	if err != nil {
		return line, err
	}
	var b strings.Builder
	for tok := iterator(); tok != chroma.EOF; tok = iterator() {
		if tok.Value == "" {
			continue
		}
		entry := style.Get(tok.Type)
		s := lipgloss.NewStyle().Background(lipgloss.Color(lineBg.String()))
		if entry.Bold == chroma.Yes {
			s = s.Bold(true)
		}
		if entry.Italic == chroma.Yes {
			s = s.Italic(true)
		}
		if entry.Underline == chroma.Yes {
			s = s.Underline(true)
		}
		if entry.Colour.IsSet() {
			s = s.Foreground(lipgloss.Color(entry.Colour.String()))
		}
		b.WriteString(s.Render(tok.Value))
	}
	if b.Len() == 0 {
		return line, nil
	}
	return b.String(), nil
}

// ReplaceAnsiBackground replaces every occurrence of the TrueColor background
// sub-sequence for `from` with the sub-sequence for `to` in s. It handles both
// standalone escapes (\x1b[48;2;R;G;Bm) and combined fg+bg SGR sequences that
// lipgloss v2 emits (\x1b[38;2;R;G;B;48;2;R;G;Bm). All other ANSI codes (foreground,
// bold, italic, etc.) are preserved unchanged.
//
// This is used by the word-diff renderer to swap the line background for the brighter
// word-span background on changed character spans without altering syntax foreground colours.
//
// The function only operates on the 24-bit TrueColor sub-sequence "48;2;R;G;B".
// When the string uses a lower colour depth the call is a no-op — an acceptable
// degradation because character-level highlights are most valuable in TrueColor terminals.
func ReplaceAnsiBackground(s string, from, to chroma.Colour) string {
	if !from.IsSet() || !to.IsSet() || from == to {
		return s
	}
	fromPart := fmt.Sprintf("48;2;%d;%d;%d", from.Red(), from.Green(), from.Blue())
	toPart := fmt.Sprintf("48;2;%d;%d;%d", to.Red(), to.Green(), to.Blue())
	return strings.ReplaceAll(s, fromPart, toPart)
}
