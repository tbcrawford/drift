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

func TestGutterNeutralTint_LightOldColumn_gutterTintStyle(t *testing.T) {
	t.Parallel()
	st := gutterTintStyle(false, false, true)
	out := st.Render("chg")
	if !strings.Contains(out, "e4e4e4") && !strings.Contains(out, "228") {
		t.Fatalf("expected light old-column neutral (#e4e4e4 / 228) in gutter tint output, got %q", out)
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
