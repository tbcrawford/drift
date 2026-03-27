package render

import (
	"bytes"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/charmbracelet/colorprofile"
	"github.com/tylercrawford/drift/internal/edittype"
	"github.com/tylercrawford/drift/internal/highlight"
)

// splitNoopConfig returns a RenderConfig suitable for deterministic split tests.
// NoOp formatter means no ANSI codes — output is plain text for string comparison.
func splitNoopConfig(termWidth int) *RenderConfig {
	return &RenderConfig{
		Lexer:     highlight.DetectLexer("go", "", ""),
		Style:     highlight.SelectTheme("", true),
		Formatter: formatters.NoOp,
		Profile:   colorprofile.NoTTY,
		TermWidth: termWidth,
	}
}

// twoLineHunk builds a minimal DiffResult with one Delete and one Insert.
func twoLineHunk() edittype.DiffResult {
	return edittype.DiffResult{
		Hunks: []edittype.Hunk{
			{
				OldStart: 1, OldLines: 1,
				NewStart: 1, NewLines: 1,
				Lines: []edittype.Line{
					{Op: edittype.Delete, Content: "old line", OldNum: 1},
					{Op: edittype.Insert, Content: "new line", NewNum: 1},
				},
			},
		},
	}
}

func TestSplit_EmptyResult(t *testing.T) {
	result := edittype.DiffResult{IsEqual: true}
	var buf bytes.Buffer
	if err := Split(result, &buf, splitNoopConfig(80)); err != nil {
		t.Fatalf("Split error: %v", err)
	}
	if buf.Len() != 0 {
		t.Errorf("expected empty output for empty result, got: %q", buf.String())
	}
}

func TestSplit_ContainsSeparator(t *testing.T) {
	var buf bytes.Buffer
	if err := Split(twoLineHunk(), &buf, splitNoopConfig(80)); err != nil {
		t.Fatalf("Split error: %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "│") {
		t.Errorf("expected split output to contain '│' separator; got:\n%s", output)
	}
}

func TestSplit_NoColorSeparator(t *testing.T) {
	cfg := splitNoopConfig(80)
	cfg.NoColor = true

	var buf bytes.Buffer
	if err := Split(twoLineHunk(), &buf, cfg); err != nil {
		t.Fatalf("Split error: %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "│") {
		t.Errorf("NoColor output must use '│' separator; got:\n%s", output)
	}
}

func TestSplit_HunkHeaderPresent(t *testing.T) {
	var buf bytes.Buffer
	if err := Split(twoLineHunk(), &buf, splitNoopConfig(80)); err != nil {
		t.Fatalf("Split error: %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "@@ -1,1 +1,1 @@") {
		t.Errorf("expected hunk header '@@ -1,1 +1,1 @@' in output; got:\n%s", output)
	}
}

func TestSplit_BothPanelContents(t *testing.T) {
	var buf bytes.Buffer
	if err := Split(twoLineHunk(), &buf, splitNoopConfig(80)); err != nil {
		t.Fatalf("Split error: %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "old line") {
		t.Errorf("expected left panel to contain 'old line'; got:\n%s", output)
	}
	if !strings.Contains(output, "new line") {
		t.Errorf("expected right panel to contain 'new line'; got:\n%s", output)
	}
}

func TestSplit_Width80_NoLineOverflow(t *testing.T) {
	result := edittype.DiffResult{
		Hunks: []edittype.Hunk{
			{
				OldStart: 1, OldLines: 1, NewStart: 1, NewLines: 1,
				Lines: []edittype.Line{
					{Op: edittype.Delete, Content: strings.Repeat("x", 100), OldNum: 1},
					{Op: edittype.Insert, Content: strings.Repeat("y", 100), NewNum: 1},
				},
			},
		},
	}

	var buf bytes.Buffer
	if err := Split(result, &buf, splitNoopConfig(80)); err != nil {
		t.Fatalf("Split error: %v", err)
	}

	for _, line := range strings.Split(buf.String(), "\n") {
		if line == "" {
			continue
		}
		w := lipgloss.Width(line)
		if w > 80 {
			t.Errorf("output line exceeds 80 columns (got %d): %q", w, line)
		}
	}
}

func TestSplit_Width120_NoLineOverflow(t *testing.T) {
	result := edittype.DiffResult{
		Hunks: []edittype.Hunk{
			{
				OldStart: 1, OldLines: 1, NewStart: 1, NewLines: 1,
				Lines: []edittype.Line{
					{Op: edittype.Delete, Content: strings.Repeat("a", 200), OldNum: 1},
					{Op: edittype.Insert, Content: strings.Repeat("b", 200), NewNum: 1},
				},
			},
		},
	}

	var buf bytes.Buffer
	if err := Split(result, &buf, splitNoopConfig(120)); err != nil {
		t.Fatalf("Split error: %v", err)
	}

	for _, line := range strings.Split(buf.String(), "\n") {
		if line == "" {
			continue
		}
		w := lipgloss.Width(line)
		if w > 120 {
			t.Errorf("output line exceeds 120 columns (got %d): %q", w, line)
		}
	}
}

func TestSplit_WordDiffPairedDeleteInsert(t *testing.T) {
	result := edittype.DiffResult{
		Hunks: []edittype.Hunk{
			{
				OldStart: 1, OldLines: 1, NewStart: 1, NewLines: 1,
				Lines: []edittype.Line{
					{Op: edittype.Delete, Content: "let x = foo bar baz", OldNum: 1},
					{Op: edittype.Insert, Content: "let x = foo qux baz", NewNum: 1},
				},
			},
		},
	}
	cfg := &RenderConfig{
		Lexer:         highlight.DetectLexer("go", "", ""),
		Style:         styles.Get("github"),
		Formatter:     formatters.TTY16m,
		Profile:       colorprofile.TrueColor,
		TermWidth:     120,
		WordDiff:      true,
		LineDiffStyle: true,
		IsDark:        false,
	}
	if cfg.Style == nil {
		cfg.Style = highlight.SelectTheme("", false)
	}
	var buf bytes.Buffer
	if err := Split(result, &buf, cfg); err != nil {
		t.Fatalf("Split error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "\033[") {
		t.Fatalf("expected ANSI in word-diff split output:\n%s", out)
	}
	if !strings.Contains(out, "bar") || !strings.Contains(out, "qux") {
		t.Fatalf("expected both sides of substitution in output:\n%s", out)
	}
}

func TestSplit_ANSIWidthNotInflated(t *testing.T) {
	plain := "func main() {}"
	cfg := &RenderConfig{
		Lexer:     highlight.DetectLexer("go", "", ""),
		Style:     styles.Get("monokai"),
		Formatter: formatters.TTY16m,
		Profile:   colorprofile.TrueColor,
		TermWidth: 80,
	}
	if cfg.Style == nil {
		cfg.Style = highlight.SelectTheme("", true)
	}

	highlighted := highlightPanel(plain, cfg.Lexer, cfg.Style, cfg.Formatter)

	if len(highlighted) == len(plain) {
		t.Error("expected highlighted string to contain ANSI bytes (len > plain len)")
	}
	if lipgloss.Width(highlighted) != lipgloss.Width(plain) {
		t.Errorf("lipgloss.Width(highlighted)=%d; want %d (same as plain display width)",
			lipgloss.Width(highlighted), lipgloss.Width(plain))
	}
}

func TestPairHunkLines_EqualLines(t *testing.T) {
	lines := []edittype.Line{
		{Op: edittype.Equal, Content: "same"},
	}
	pairs := pairHunkLines(lines)
	if len(pairs) != 1 {
		t.Fatalf("expected 1 pair, got %d", len(pairs))
	}
	if pairs[0].left != "same" || pairs[0].right != "same" {
		t.Errorf("Equal line: left=%q right=%q; want both 'same'", pairs[0].left, pairs[0].right)
	}
}

func TestPairHunkLines_MoreDeletesThanInserts(t *testing.T) {
	lines := []edittype.Line{
		{Op: edittype.Delete, Content: "del1"},
		{Op: edittype.Delete, Content: "del2"},
		{Op: edittype.Insert, Content: "ins1"},
	}
	pairs := pairHunkLines(lines)
	if len(pairs) != 2 {
		t.Fatalf("expected 2 pairs, got %d", len(pairs))
	}
	if pairs[0].left != "del1" || pairs[0].right != "ins1" {
		t.Errorf("pair[0]: left=%q right=%q; want left='del1' right='ins1'", pairs[0].left, pairs[0].right)
	}
	if pairs[1].left != "del2" || pairs[1].right != "" {
		t.Errorf("pair[1]: left=%q right=%q; want left='del2' right=''", pairs[1].left, pairs[1].right)
	}
}

func TestPairHunkLines_MoreInsertsThanDeletes(t *testing.T) {
	lines := []edittype.Line{
		{Op: edittype.Delete, Content: "del1"},
		{Op: edittype.Insert, Content: "ins1"},
		{Op: edittype.Insert, Content: "ins2"},
	}
	pairs := pairHunkLines(lines)
	if len(pairs) != 2 {
		t.Fatalf("expected 2 pairs, got %d", len(pairs))
	}
	if pairs[0].left != "del1" || pairs[0].right != "ins1" {
		t.Errorf("pair[0]: left=%q right=%q; want left='del1' right='ins1'", pairs[0].left, pairs[0].right)
	}
	if pairs[1].left != "" || pairs[1].right != "ins2" {
		t.Errorf("pair[1]: left=%q right=%q; want left='' right='ins2'", pairs[1].left, pairs[1].right)
	}
}

func TestPairHunkLines_OnlyDeletes(t *testing.T) {
	lines := []edittype.Line{
		{Op: edittype.Delete, Content: "d1"},
		{Op: edittype.Delete, Content: "d2"},
	}
	pairs := pairHunkLines(lines)
	if len(pairs) != 2 {
		t.Fatalf("expected 2 pairs, got %d", len(pairs))
	}
	for i, p := range pairs {
		if p.right != "" {
			t.Errorf("pair[%d].right = %q; want '' (blank) for Delete-only hunk", i, p.right)
		}
	}
}

func TestPairHunkLines_OnlyInserts(t *testing.T) {
	lines := []edittype.Line{
		{Op: edittype.Insert, Content: "i1"},
		{Op: edittype.Insert, Content: "i2"},
	}
	pairs := pairHunkLines(lines)
	if len(pairs) != 2 {
		t.Fatalf("expected 2 pairs, got %d", len(pairs))
	}
	for i, p := range pairs {
		if p.left != "" {
			t.Errorf("pair[%d].left = %q; want '' (blank) for Insert-only hunk", i, p.left)
		}
	}
}
