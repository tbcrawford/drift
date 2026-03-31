package highlight

import (
	"fmt"
	"strings"
	"testing"

	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/tbcrawford/drift/internal/edittype"
)

func TestDiffLineStyle_Equal(t *testing.T) {
	t.Helper()
	style := styles.Get("github")
	_, ok := DiffLineStyle(style, edittype.Equal, true)
	if ok {
		t.Fatal("expected no style for Equal lines")
	}
}

func TestHighlightLineWithLineBackground_githubDeleteHasBackground(t *testing.T) {
	t.Helper()
	style := styles.Get("github")
	bg, ok := DiffLineStyle(style, edittype.Delete, false)
	if !ok {
		t.Fatal("expected github GenericDeleted background")
	}
	lexer := lexers.Fallback
	out, err := HighlightLineWithLineBackground("hello", lexer, style, bg)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, ";48;2") && !strings.Contains(out, "\x1b[48;") && !strings.Contains(out, "\x1b[48:") {
		t.Fatalf("expected background ANSI in output, got %q", out)
	}
}

func TestHighlightLineWithLineBackground_githubDarkGoSyntax(t *testing.T) {
	t.Helper()
	style := styles.Get("github-dark")
	if style == nil {
		t.Fatal("github-dark")
	}
	bg, ok := DiffLineStyle(style, edittype.Insert, true)
	if !ok {
		t.Fatal("expected DiffLineStyle for insert")
	}
	lexer := lexers.Get("go")
	if lexer == nil {
		t.Fatal("go lexer")
	}
	out, err := HighlightLineWithLineBackground(`return "x"`, lexer, style, bg)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, ";48;2") && !strings.Contains(out, "\x1b[48;") && !strings.Contains(out, "\x1b[48:") {
		t.Fatalf("expected line-level background CSI 48 in output, got:\n%q", out)
	}
}

func TestReplaceAnsiBackground_swapsOnlyBackground(t *testing.T) {
	t.Parallel()
	style := styles.Get("github-dark")
	if style == nil {
		t.Fatal("github-dark style")
	}
	lexer := lexers.Get("go")
	if lexer == nil {
		t.Fatal("go lexer")
	}
	lineBg := DiffLineBackgroundColour(style, true, true) // #490202
	wordBg := WordSpanBackgroundColour(style, true, true) // brighter red
	if !lineBg.IsSet() || !wordBg.IsSet() {
		t.Fatal("expected set colours")
	}

	out, err := HighlightLineWithLineBackground("return", lexer, style, lineBg)
	if err != nil {
		t.Fatal(err)
	}
	replaced := ReplaceAnsiBackground(out, lineBg, wordBg)

	wantBg := fmt.Sprintf("48;2;%d;%d;%d", wordBg.Red(), wordBg.Green(), wordBg.Blue())
	if !strings.Contains(replaced, wantBg) {
		t.Fatalf("expected word-span background %q in replaced output, got:\n%q", wantBg, replaced)
	}
	oldBg := fmt.Sprintf("48;2;%d;%d;%d", lineBg.Red(), lineBg.Green(), lineBg.Blue())
	if strings.Contains(replaced, oldBg) {
		t.Fatalf("line background %q should be absent after replacement, got:\n%q", oldBg, replaced)
	}
}
