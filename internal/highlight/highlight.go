// Package highlight provides Chroma v2 syntax highlighting for drift diff output.
package highlight

import (
	"bytes"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	chromastyles "github.com/alecthomas/chroma/v2/styles"
	"github.com/charmbracelet/colorprofile"
)

// HighlightLine applies Chroma syntax highlighting to a single line of text.
//
// The line is tokenized using the provided lexer, styled with style, and
// formatted using formatter. On tokenization or formatting error, the original
// line is returned unmodified (fail-open: plain text is always preferable to
// an error that blocks rendering).
//
// Note: per-line tokenization loses multi-line context (e.g., a string
// spanning multiple lines). This is acceptable for v1 unified diff output
// where individual lines are the unit of display.
func HighlightLine(line string, lexer chroma.Lexer, style *chroma.Style, formatter chroma.Formatter) (string, error) {
	iterator, err := lexer.Tokenise(nil, line)
	if err != nil {
		return line, err
	}
	var buf bytes.Buffer
	if err := formatter.Format(&buf, style, iterator); err != nil {
		return line, err
	}
	// Strip the trailing newline that Chroma formatters append.
	return strings.TrimRight(buf.String(), "\n"), nil
}

// FormatterForProfile returns the Chroma formatter appropriate for the given
// terminal color profile. The mapping is:
//
//	TrueColor → terminal16m (24-bit)
//	ANSI256   → terminal256 (8-bit)
//	ANSI      → terminal16  (4-bit)
//	Ascii     → noop        (plain text, NO_COLOR compliant)
//	NoTTY     → noop        (plain text, output is piped)
func FormatterForProfile(p colorprofile.Profile) chroma.Formatter {
	switch p {
	case colorprofile.TrueColor:
		return formatters.TTY16m
	case colorprofile.ANSI256:
		return formatters.TTY256
	case colorprofile.ANSI:
		return formatters.TTY16
	default: // Ascii, NoTTY
		return formatters.NoOp
	}
}

// SelectTheme returns the Chroma style for the given theme name and dark/light
// terminal preference.
//
// If requested is non-empty, it is looked up by exact name in the registry.
// If the name is unknown, SelectTheme falls back to auto-detection.
// Auto-detection uses "monokai" for dark terminals and "github" for light terminals.
// chromastyles.Fallback ("swapoff") is used as the last resort — it is always registered.
func SelectTheme(requested string, isDark bool) *chroma.Style {
	if requested != "" {
		if s, ok := chromastyles.Registry[requested]; ok {
			return s
		}
		// Unknown theme name: fall through to auto-detect.
	}
	name := "monokai"
	if !isDark {
		name = "github"
	}
	if s, ok := chromastyles.Registry[name]; ok {
		return s
	}
	return chromastyles.Fallback
}

// DetectLexer returns the best Chroma lexer for the given language name,
// filename, and content. The selection priority is:
//  1. Explicit language name via lexers.Get(lang) — e.g., "go", "python"
//  2. Filename/extension match via lexers.Match(filename) — e.g., "main.go"
//  3. Content analysis via lexers.Analyse(content)
//  4. lexers.Fallback (plaintext) if nothing matches
//
// The returned lexer is always wrapped with chroma.Coalesce to merge adjacent
// same-type tokens, reducing ANSI escape sequence noise in output.
func DetectLexer(lang, filename, content string) chroma.Lexer {
	var lexer chroma.Lexer

	if lang != "" {
		lexer = lexers.Get(lang)
	}
	if lexer == nil && filename != "" {
		lexer = lexers.Match(filename)
	}
	if lexer == nil && content != "" {
		lexer = lexers.Analyse(content)
	}
	if lexer == nil {
		lexer = lexers.Fallback
	}
	return chroma.Coalesce(lexer)
}
