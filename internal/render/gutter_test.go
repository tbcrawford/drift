package render

import (
	"strings"
	"testing"

	"github.com/alecthomas/chroma/v2/styles"
	"github.com/tylercrawford/drift/internal/edittype"
)

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

func TestWordSpanStyle_delete_hasBackgroundANSI(t *testing.T) {
	t.Parallel()
	style := styles.Get("github-dark")
	if style == nil {
		t.Fatal("github-dark style")
	}
	st := wordSpanStyle(style, true, false, true)
	out := st.Render("chg")
	if !strings.Contains(out, "\x1b[48;") {
		t.Fatalf("expected full-line background CSI 48 in word-span style output, got %q", out)
	}
}

func TestGutterStyleForCell_contextUsesNeutralGray(t *testing.T) {
	t.Parallel()
	style := styles.Get("github")
	if style == nil {
		t.Fatal("github style")
	}
	st := gutterStyleForCell(style, false, false, true, edittype.Equal)
	out := st.Render(" ")
	// Neutral github light gutter uses #e4e4e4 for old column.
	if !strings.Contains(out, "e4e4e4") && !strings.Contains(out, "228") {
		t.Fatalf("expected neutral light gray in gutter output, got %q", out)
	}
}
