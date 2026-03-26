package render

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/charmbracelet/x/ansi"
	"github.com/tylercrawford/drift/internal/edittype"
	"github.com/tylercrawford/drift/internal/highlight"
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

func TestGutterDeleteCell_usesSemanticWordSpanColour(t *testing.T) {
	t.Parallel()
	style := styles.Get("github-dark")
	if style == nil {
		t.Fatal("github-dark style")
	}
	if !highlight.WordSpanBackgroundColour(style, true, true).IsSet() {
		t.Fatal("expected word-span colour")
	}
	st := gutterStyleForCell(style, true, false, true, edittype.Delete)
	out := GutterNumberRender(st, 3, 2)
	// Foreground may precede background (38…48…); require embedded 48 (bg) sequence.
	if !strings.Contains(out, ";48;") && !strings.Contains(out, ";48:") {
		t.Fatalf("expected background SGR (48) in delete gutter (same pipeline as word spans): %q", out)
	}
}

func TestGutterStyleForCell_deleteOldColumn_hasBackgroundANSI(t *testing.T) {
	t.Parallel()
	style := styles.Get("github")
	if style == nil {
		t.Fatal("github style")
	}
	st := gutterStyleForCell(style, false, false, true, edittype.Delete)
	out := st.Render("1")
	if !strings.Contains(out, ";48;") && !strings.Contains(out, ";48:") {
		t.Fatalf("expected background SGR (48) in gutter output, got %q", out)
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
