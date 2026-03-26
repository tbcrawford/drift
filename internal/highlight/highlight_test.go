package highlight

import (
	"bytes"
	"strings"
	"testing"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/charmbracelet/colorprofile"
)

// renderLine runs the formatter on a simple Go line and returns the output.
func renderLine(f chroma.Formatter) (string, error) {
	style := styles.Get("monokai")
	lexer := chroma.Coalesce(lexers.Get("go"))
	iter, err := lexer.Tokenise(nil, "func main() {}")
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := f.Format(&buf, style, iter); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func TestFormatterForProfile(t *testing.T) {
	tests := []struct {
		profile   colorprofile.Profile
		wantColor bool // true = should produce ANSI codes, false = plain text
		name      string
	}{
		{colorprofile.TrueColor, true, "TrueColor"},
		{colorprofile.ANSI256, true, "ANSI256"},
		{colorprofile.ANSI, true, "ANSI"},
		{colorprofile.Ascii, false, "Ascii"},
		{colorprofile.NoTTY, false, "NoTTY"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := FormatterForProfile(tc.profile)
			if got == nil {
				t.Fatal("FormatterForProfile returned nil")
			}
			output, err := renderLine(got)
			if err != nil {
				t.Fatalf("formatter render error: %v", err)
			}
			hasANSI := strings.Contains(output, "\033[")
			if tc.wantColor && !hasANSI {
				t.Errorf("%s: expected ANSI codes in output but got none", tc.name)
			}
			if !tc.wantColor && hasANSI {
				t.Errorf("%s: expected plain text but got ANSI codes: %q", tc.name, output)
			}
		})
	}
}

func TestFormatterForProfile_NoopIsNoOp(t *testing.T) {
	// Ascii and NoTTY should produce identical output as formatters.NoOp.
	noopFmt := formatters.NoOp
	asciiOut, err := renderLine(FormatterForProfile(colorprofile.Ascii))
	if err != nil {
		t.Fatalf("Ascii formatter error: %v", err)
	}
	noopOut, err := renderLine(noopFmt)
	if err != nil {
		t.Fatalf("NoOp formatter error: %v", err)
	}
	if asciiOut != noopOut {
		t.Errorf("Ascii output differs from NoOp:\n  got:  %q\n  want: %q", asciiOut, noopOut)
	}
}

func TestSelectTheme_DarkDefault(t *testing.T) {
	style := SelectTheme("", true)
	if style == nil {
		t.Fatal("SelectTheme returned nil for dark default")
	}
	if style.Name != "github-dark" {
		t.Errorf("dark default theme = %q; want %q", style.Name, "github-dark")
	}
}

func TestSelectTheme_LightDefault(t *testing.T) {
	style := SelectTheme("", false)
	if style == nil {
		t.Fatal("SelectTheme returned nil for light default")
	}
	if style.Name != "github" {
		t.Errorf("light default theme = %q; want %q", style.Name, "github")
	}
}

func TestSelectTheme_ExplicitOverride(t *testing.T) {
	style := SelectTheme("dracula", true)
	if style == nil {
		t.Fatal("SelectTheme returned nil for explicit 'dracula'")
	}
	if style.Name != "dracula" {
		t.Errorf("explicit theme = %q; want %q", style.Name, "dracula")
	}
}

func TestSelectTheme_UnknownFallsBackToAutoDetect(t *testing.T) {
	style := SelectTheme("nonexistent-theme-xyz", true)
	if style == nil {
		t.Fatal("SelectTheme returned nil for unknown theme with dark=true")
	}
	if style.Name != "github-dark" {
		t.Errorf("unknown theme fallback (dark) = %q; want %q", style.Name, "github-dark")
	}
}

func TestHighlightLine_NoopFormatter_PlainText(t *testing.T) {
	lexer := lexers.Get("go")
	if lexer == nil {
		t.Skip("Go lexer not available")
	}
	style := SelectTheme("", true) // monokai

	result, err := HighlightLine(`func main() {}`, lexer, style, formatters.NoOp)
	if err != nil {
		t.Fatalf("HighlightLine error: %v", err)
	}
	if strings.Contains(result, "\033[") {
		t.Errorf("NoOp formatter produced ANSI codes in output: %q", result)
	}
	if !strings.Contains(result, "func") {
		t.Errorf("output missing 'func' keyword: %q", result)
	}
}

func TestHighlightLine_TrueColorFormatter_ContainsANSI(t *testing.T) {
	lexer := lexers.Get("go")
	if lexer == nil {
		t.Skip("Go lexer not available")
	}
	style := SelectTheme("", true) // monokai

	result, err := HighlightLine(`func main() {}`, lexer, style, formatters.TTY16m)
	if err != nil {
		t.Fatalf("HighlightLine error: %v", err)
	}
	if !strings.Contains(result, "\033[") {
		t.Errorf("TrueColor formatter produced no ANSI codes: %q", result)
	}
}

func TestHighlightLine_FallbackOnError(t *testing.T) {
	lexer := lexers.Fallback
	style := SelectTheme("", true)
	result, err := HighlightLine("plain text line", lexer, style, formatters.NoOp)
	if err != nil {
		t.Fatalf("HighlightLine with fallback lexer error: %v", err)
	}
	if !strings.Contains(result, "plain text line") {
		t.Errorf("fallback output missing original text: %q", result)
	}
}

func TestDetectLexer_ExplicitLang(t *testing.T) {
	lexer := DetectLexer("go", "", "")
	if lexer == nil {
		t.Fatal("DetectLexer with explicit lang='go' returned nil")
	}
	iter, err := lexer.Tokenise(nil, "package main")
	if err != nil {
		t.Fatalf("tokenize error: %v", err)
	}
	tokens := iter.Tokens()
	if len(tokens) == 0 {
		t.Error("expected tokens for 'package main' but got none")
	}
}

func TestFilenameForLexer_gitDisplaySuffixes(t *testing.T) {
	if got := FilenameForLexer("main.tf (HEAD)"); got != "main.tf" {
		t.Fatalf("FilenameForLexer(main.tf (HEAD)) = %q, want main.tf", got)
	}
	if got := FilenameForLexer("foo.go (working tree)"); got != "foo.go" {
		t.Fatalf("FilenameForLexer = %q, want foo.go", got)
	}
	if got := FilenameForLexer("unchanged.go"); got != "unchanged.go" {
		t.Fatalf("FilenameForLexer = %q", got)
	}
}

func TestDetectLexer_FilenameExtension(t *testing.T) {
	lexer := DetectLexer("", "main.go", "")
	if lexer == nil {
		t.Fatal("DetectLexer with filename='main.go' returned nil")
	}
}

func TestDetectLexer_gitModeDisplayNamesStillMatchExtension(t *testing.T) {
	plain := DetectLexer("", "main.tf (HEAD)", "")
	if plain == nil {
		t.Fatal("DetectLexer returned nil")
	}
	if plain.Config().Name == "plaintext" || plain.Config().Name == "fallback" {
		t.Fatalf("expected non-plaintext lexer for .tf after sanitizing display name, got %q", plain.Config().Name)
	}
}

func TestDetectLexer_PythonOverride(t *testing.T) {
	lexer := DetectLexer("python", "", "")
	if lexer == nil {
		t.Fatal("DetectLexer with explicit lang='python' returned nil")
	}
}

func TestDetectLexer_UnknownExtensionFallback(t *testing.T) {
	lexer := DetectLexer("", "unknown.xyzzy", "")
	if lexer == nil {
		t.Fatal("DetectLexer with unknown extension returned nil; expected Fallback lexer")
	}
}

func TestDetectLexer_ContentAnalysis(t *testing.T) {
	goContent := "package main\n\nfunc main() {\n\tprintln(\"hello\")\n}\n"
	lexer := DetectLexer("", "", goContent)
	if lexer == nil {
		t.Fatal("DetectLexer with Go content returned nil")
	}
}

func TestDetectLexer_ExplicitPriorityOverFilename(t *testing.T) {
	lexer := DetectLexer("python", "main.go", "")
	if lexer == nil {
		t.Fatal("DetectLexer returned nil")
	}
	_, err := lexer.Tokenise(nil, "def foo(): pass")
	if err != nil {
		t.Fatalf("tokenize Python code with Python lexer error: %v", err)
	}
}
