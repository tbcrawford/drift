package render

import (
	"bytes"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/charmbracelet/colorprofile"
	"github.com/tbcrawford/drift/internal/edittype"
	"github.com/tbcrawford/drift/internal/highlight"
)

// --- Split edge cases ---

// TestSplit_NilConfig verifies Split does not panic with a nil config — it
// should synthesise safe defaults and produce output.
func TestSplit_NilConfig(t *testing.T) {
	var buf bytes.Buffer
	result := edittype.DiffResult{
		Hunks: []edittype.Hunk{
			{
				OldStart: 1, OldLines: 1, NewStart: 1, NewLines: 1,
				Lines: []edittype.Line{
					{Op: edittype.Delete, Content: "old", OldNum: 1},
					{Op: edittype.Insert, Content: "new", NewNum: 1},
				},
			},
		},
	}
	if err := Split(result, &buf, nil); err != nil {
		t.Fatalf("Split(nil cfg) error: %v", err)
	}
	if buf.Len() == 0 {
		t.Fatal("expected non-empty output with nil config")
	}
}

// TestSplit_NarrowTerminal verifies that a termWidth below minTermWidth (40) is
// clamped to 40, preventing negative panel widths.
func TestSplit_NarrowTerminal(t *testing.T) {
	cfg := &RenderConfig{
		Lexer:     highlight.DetectLexer("", "", ""),
		Style:     highlight.SelectTheme("", true),
		Formatter: formatters.NoOp,
		Profile:   colorprofile.NoTTY,
		TermWidth: 10, // below minTermWidth
	}
	var buf bytes.Buffer
	result := edittype.DiffResult{
		Hunks: []edittype.Hunk{
			{
				OldStart: 1, OldLines: 1, NewStart: 1, NewLines: 1,
				Lines: []edittype.Line{
					{Op: edittype.Delete, Content: "x", OldNum: 1},
					{Op: edittype.Insert, Content: "y", NewNum: 1},
				},
			},
		},
	}
	if err := Split(result, &buf, cfg); err != nil {
		t.Fatalf("Split narrow terminal error: %v", err)
	}
	// Output should be present; each rendered line should fit within 40 cols.
	for _, line := range strings.Split(buf.String(), "\n") {
		if line == "" {
			continue
		}
		if w := lipgloss.Width(line); w > 40 {
			t.Errorf("narrow terminal: line width %d > 40: %q", w, line)
		}
	}
}

// TestSplit_ShowLineNumbers verifies that ShowLineNumbers=true triggers the gutter path.
func TestSplit_ShowLineNumbers(t *testing.T) {
	cfg := &RenderConfig{
		Lexer:           highlight.DetectLexer("", "", ""),
		Style:           highlight.SelectTheme("", true),
		Formatter:       formatters.NoOp,
		Profile:         colorprofile.NoTTY,
		TermWidth:       80,
		ShowLineNumbers: true,
		NoColor:         true,
	}
	var buf bytes.Buffer
	result := edittype.DiffResult{
		Hunks: []edittype.Hunk{
			{
				OldStart: 1, OldLines: 1, NewStart: 1, NewLines: 1,
				Lines: []edittype.Line{
					{Op: edittype.Delete, Content: "old", OldNum: 1},
					{Op: edittype.Insert, Content: "new", NewNum: 1},
				},
			},
		},
	}
	if err := Split(result, &buf, cfg); err != nil {
		t.Fatalf("Split ShowLineNumbers error: %v", err)
	}
	out := buf.String()
	// Should contain the line number "1" in the output.
	if !strings.Contains(out, "1") {
		t.Errorf("expected line number '1' in output: %q", out)
	}
}

// --- GutterNumberRender edge cases ---

// TestGutterNumberRender_widthLessThanOne verifies width<1 is clamped to 1.
func TestGutterNumberRender_widthLessThanOne(t *testing.T) {
	st := lipgloss.NewStyle()
	// Width=0 should be treated as width=1.
	out := GutterNumberRender(st, 0, 5)
	if out == "" {
		t.Fatal("expected non-empty output for width=0")
	}
	// Should not panic and contain the digit.
	if !strings.Contains(out, "5") {
		t.Errorf("expected digit '5' in output: %q", out)
	}
}

// TestGutterNumberRender_largeNumber verifies that a line number wider than the column
// falls back to plain rendering of just the digits (no padding).
func TestGutterNumberRender_largeNumber(t *testing.T) {
	st := lipgloss.NewStyle()
	// Width=2, number=99999 — inner string " 99999 " (len=7) exceeds width=2.
	// digits "99999" (len=5) also > width=2, so st.Render(s) path fires.
	out := GutterNumberRender(st, 2, 99999)
	if !strings.Contains(out, "99999") {
		t.Errorf("expected large number '99999' in output: %q", out)
	}
}

// --- gutterPairWidths coverage ---

// TestGutterPairWidths_basic verifies gutterPairWidths uses leftOldNum and rightNewNum.
func TestGutterPairWidths_basic(t *testing.T) {
	pairs := []linePair{
		{leftOldNum: 1, rightNewNum: 100},
		{leftOldNum: 5, rightNewNum: 200},
	}
	oldW, newW := gutterPairWidths(pairs)
	// 100 → 3 digits + 2 spaces = 5; 200 → 3 digits + 2 spaces = 5
	if oldW < 1 {
		t.Errorf("oldW should be >= 1, got %d", oldW)
	}
	if newW < 3 {
		t.Errorf("newW should accommodate 3-digit numbers, got %d", newW)
	}
}
