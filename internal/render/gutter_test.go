package render

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/charmbracelet/x/ansi"
	"github.com/tbcrawford/drift/internal/edittype"
)

func TestGutterNumberRender_padsSingleDigitWithSpaces(t *testing.T) {
	t.Parallel()
	style := styles.Get("github")
	if style == nil {
		t.Fatal("github style")
	}
	st := gutterStyleForCell(style, false, false, true, edittype.Equal)
	out := GutterNumberRender(st, 3, 1)
	plain := ansi.Strip(out)
	if plain != " 1 " {
		t.Fatalf("want padded \" 1 \", got %q", plain)
	}
}

func TestGutterNumberRender_blankFillsFullColumnWidth(t *testing.T) {
	t.Parallel()
	style := styles.Get("github")
	if style == nil {
		t.Fatal("github style")
	}
	st := gutterStyleForCell(style, false, false, true, edittype.Equal)
	out := GutterNumberRender(st, 5, 0)
	if lipgloss.Width(out) != 5 {
		t.Fatalf("blank gutter cell width: got %d want 5 (output %q)", lipgloss.Width(out), out)
	}
}

func TestGutterDeleteCell_usesAccentForeground(t *testing.T) {
	t.Parallel()
	style := styles.Get("github-dark")
	if style == nil {
		t.Fatal("github-dark style")
	}
	st := gutterStyleForCell(style, true, false, true, edittype.Delete)
	out := GutterNumberRender(st, 3, 2)
	// Delete gutter cells now use accent blue foreground (color 63 / 38;5;63) with NO background.
	if strings.Contains(out, ";48;") || strings.Contains(out, ";48:") {
		t.Fatalf("delete gutter should not set background color, got %q", out)
	}
	// Should contain foreground ANSI (38).
	if !strings.Contains(out, "\x1b[38;") {
		t.Fatalf("expected foreground ANSI (38) in delete gutter, got %q", out)
	}
}

func TestGutterStyleForCell_deleteOldColumn_hasAccentForeground(t *testing.T) {
	t.Parallel()
	style := styles.Get("github")
	if style == nil {
		t.Fatal("github style")
	}
	st := gutterStyleForCell(style, false, false, true, edittype.Delete)
	out := st.Render("1")
	// Changed: delete gutter now uses accent foreground only — no background.
	if strings.Contains(out, ";48;") || strings.Contains(out, ";48:") {
		t.Fatalf("delete gutter should not set background color, got %q", out)
	}
	// Must still have a foreground color (the accent blue).
	if !strings.Contains(out, "\x1b[38;") {
		t.Fatalf("expected foreground ANSI (38) in delete gutter output, got %q", out)
	}
}

func TestWordSpanBg_delete_hasBackgroundANSI(t *testing.T) {
	t.Parallel()
	style := styles.Get("github-dark")
	if style == nil {
		t.Fatal("github-dark style")
	}
	c := wordSpanBg(style, true, false, true)
	if !c.IsSet() {
		t.Fatal("expected set word span background colour")
	}
	out := lipgloss.NewStyle().Background(lipgloss.Color(c.String())).Render("chg")
	if !strings.Contains(out, "\x1b[48;") {
		t.Fatalf("expected background CSI 48 in word-span output, got %q", out)
	}
}

func TestGutterStyleForCell_contextHasNoBackgroundColor(t *testing.T) {
	t.Parallel()
	style := styles.Get("github")
	if style == nil {
		t.Fatal("github style")
	}
	st := gutterStyleForCell(style, false, false, true, edittype.Equal)
	out := st.Render(" ")
	if strings.Contains(out, "\x1b[48;") {
		t.Fatalf("context gutter should not set background (no gray wash), got %q", out)
	}
}

// TestGutterStyleCache_cacheCorrectness verifies that NewGutterStyleCache pre-populates
// all 6 style variants (2 sides × 3 ops) and that Get() returns styles producing
// identical ANSI output to gutterStyleForCell called directly.
func TestGutterStyleCache_cacheCorrectness(t *testing.T) {
	t.Parallel()
	style := styles.Get("github-dark")
	if style == nil {
		t.Fatal("github-dark style")
	}
	cache := NewGutterStyleCache(style, true, false)

	// Verify all 6 combinations produce same output as direct gutterStyleForCell.
	for _, oldCol := range []bool{true, false} {
		for _, op := range []edittype.Op{edittype.Equal, edittype.Delete, edittype.Insert} {
			directSt := gutterStyleForCell(style, true, false, oldCol, op)
			cachedSt := cache.Get(oldCol, op)
			directOut := GutterNumberRender(directSt, 5, 42)
			cachedOut := GutterNumberRender(cachedSt, 5, 42)
			if directOut != cachedOut {
				t.Errorf("cache mismatch for oldColumn=%v op=%v:\ndirect: %q\ncached: %q",
					oldCol, op, directOut, cachedOut)
			}
		}
	}
}

// TestGutterStyleCache_noColor verifies that noColor cache returns plain styles.
func TestGutterStyleCache_noColor(t *testing.T) {
	t.Parallel()
	style := styles.Get("github")
	if style == nil {
		t.Fatal("github style")
	}
	cache := NewGutterStyleCache(style, false, true) // noColor=true
	st := cache.Get(true, edittype.Delete)
	out := GutterNumberRender(st, 3, 1)
	// No-color output should have no ANSI background sequences.
	if strings.Contains(out, "\x1b[") {
		t.Errorf("noColor cache produced ANSI escape sequences: %q", out)
	}
}
