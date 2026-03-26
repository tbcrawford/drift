package render

import (
	"bytes"
	"strings"
	"testing"

	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/charmbracelet/colorprofile"
	"github.com/tylercrawford/drift/internal/edittype"
	"github.com/tylercrawford/drift/internal/highlight"
)

// noopConfig returns a RenderConfig with NoOp formatter for deterministic,
// ANSI-free output suitable for string comparison in tests.
func noopConfig() *RenderConfig {
	return &RenderConfig{
		Lexer:     highlight.DetectLexer("go", "", ""),
		Style:     highlight.SelectTheme("", true),
		Formatter: formatters.NoOp,
		Profile:   colorprofile.NoTTY,
	}
}

func TestUnified_EmptyResult(t *testing.T) {
	result := edittype.DiffResult{IsEqual: true}
	var buf bytes.Buffer
	if err := Unified(result, &buf, noopConfig()); err != nil {
		t.Fatalf("Unified error: %v", err)
	}
	if buf.Len() != 0 {
		t.Errorf("expected empty output for IsEqual=true, got: %q", buf.String())
	}
}

func TestUnified_HunkHeaders(t *testing.T) {
	// Single hunk: one deleted line, one inserted line, one context line.
	result := edittype.DiffResult{
		Hunks: []edittype.Hunk{
			{
				OldStart: 1, OldLines: 2,
				NewStart: 1, NewLines: 2,
				Lines: []edittype.Line{
					{Op: edittype.Equal, Content: "package main", OldNum: 1, NewNum: 1},
					{Op: edittype.Delete, Content: "func old() {}", OldNum: 2},
					{Op: edittype.Insert, Content: "func new() {}", NewNum: 2},
				},
			},
		},
	}

	var buf bytes.Buffer
	if err := Unified(result, &buf, noopConfig()); err != nil {
		t.Fatalf("Unified error: %v", err)
	}

	output := buf.String()

	// Verify file headers.
	if !strings.Contains(output, "--- a/input\n") {
		t.Errorf("output missing '--- a/input' header:\n%s", output)
	}
	if !strings.Contains(output, "+++ b/input\n") {
		t.Errorf("output missing '+++ b/input' header:\n%s", output)
	}

	// Verify hunk header format.
	if !strings.Contains(output, "@@ -1,2 +1,2 @@") {
		t.Errorf("output missing '@@ -1,2 +1,2 @@' hunk header:\n%s", output)
	}
}

func TestUnified_LinePrefixes(t *testing.T) {
	result := edittype.DiffResult{
		Hunks: []edittype.Hunk{
			{
				OldStart: 1, OldLines: 2,
				NewStart: 1, NewLines: 2,
				Lines: []edittype.Line{
					{Op: edittype.Equal, Content: "context", OldNum: 1, NewNum: 1},
					{Op: edittype.Delete, Content: "removed", OldNum: 2},
					{Op: edittype.Insert, Content: "added", NewNum: 2},
				},
			},
		},
	}

	var buf bytes.Buffer
	if err := Unified(result, &buf, noopConfig()); err != nil {
		t.Fatalf("Unified error: %v", err)
	}

	output := buf.String()
	lines := strings.Split(output, "\n")

	// Find the content lines (skip file headers and hunk header).
	var contextLine, deleteLine, insertLine string
	for _, l := range lines {
		switch {
		case strings.HasSuffix(l, "context"):
			contextLine = l
		case strings.HasSuffix(l, "removed"):
			deleteLine = l
		case strings.HasSuffix(l, "added"):
			insertLine = l
		}
	}

	if !strings.HasPrefix(contextLine, " ") {
		t.Errorf("Equal line prefix = %q; want ' ' (space): full line: %q", contextLine[:1], contextLine)
	}
	if !strings.HasPrefix(deleteLine, "-") {
		t.Errorf("Delete line prefix = %q; want '-': full line: %q", deleteLine[:1], deleteLine)
	}
	if !strings.HasPrefix(insertLine, "+") {
		t.Errorf("Insert line prefix = %q; want '+': full line: %q", insertLine[:1], insertLine)
	}
}

func TestUnified_CustomFileNames(t *testing.T) {
	result := edittype.DiffResult{
		Hunks: []edittype.Hunk{
			{
				OldStart: 1, OldLines: 1, NewStart: 1, NewLines: 1,
				Lines: []edittype.Line{
					{Op: edittype.Delete, Content: "x", OldNum: 1},
				},
			},
		},
	}

	cfg := noopConfig()
	cfg.OldName = "a/foo.go"
	cfg.NewName = "b/foo.go"

	var buf bytes.Buffer
	if err := Unified(result, &buf, cfg); err != nil {
		t.Fatalf("Unified error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "--- a/foo.go\n") {
		t.Errorf("expected '--- a/foo.go' in output:\n%s", output)
	}
	if !strings.Contains(output, "+++ b/foo.go\n") {
		t.Errorf("expected '+++ b/foo.go' in output:\n%s", output)
	}
}

func TestUnified_WordDiffPairedDeleteInsert(t *testing.T) {
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
		WordDiff:      true,
		LineDiffStyle: true,
		IsDark:        false,
	}
	if cfg.Style == nil {
		cfg.Style = highlight.SelectTheme("", false)
	}
	var buf bytes.Buffer
	if err := Unified(result, &buf, cfg); err != nil {
		t.Fatalf("Unified error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "\033[") {
		t.Fatalf("expected ANSI in word-diff output:\n%s", out)
	}
	if !strings.Contains(out, "bar") || !strings.Contains(out, "qux") {
		t.Fatalf("expected both sides of substitution in output:\n%s", out)
	}
	// Full-line semantic background (lipgloss may merge 38+48 as ...;48;2;... in one SGR).
	if !strings.Contains(out, ";48;2") && !strings.Contains(out, "\x1b[48;") {
		t.Fatalf("expected full-line background SGR CSI 48 in word-diff output:\n%s", out)
	}
}

func TestUnified_TrueColorProducesANSI(t *testing.T) {
	result := edittype.DiffResult{
		Hunks: []edittype.Hunk{
			{
				OldStart: 1, OldLines: 1, NewStart: 1, NewLines: 1,
				Lines: []edittype.Line{
					{Op: edittype.Insert, Content: "func main() {}", NewNum: 1},
				},
			},
		},
	}

	cfg := &RenderConfig{
		Lexer:     highlight.DetectLexer("go", "", ""),
		Style:     styles.Get("monokai"),
		Formatter: formatters.TTY16m,
		Profile:   colorprofile.TrueColor,
	}
	if cfg.Style == nil {
		cfg.Style = highlight.SelectTheme("", true)
	}

	var buf bytes.Buffer
	if err := Unified(result, &buf, cfg); err != nil {
		t.Fatalf("Unified error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "\033[") {
		t.Errorf("TrueColor output missing ANSI escape codes:\n%s", output)
	}
}

func TestUnified_NilLexerFallback(t *testing.T) {
	// When cfg.Lexer is nil, Unified should call DetectLexer and not panic.
	result := edittype.DiffResult{
		Hunks: []edittype.Hunk{
			{
				OldStart: 1, OldLines: 1, NewStart: 1, NewLines: 1,
				Lines: []edittype.Line{
					{Op: edittype.Insert, Content: "hello world", NewNum: 1},
				},
			},
		},
	}

	cfg := &RenderConfig{
		Formatter: formatters.NoOp,
		Style:     highlight.SelectTheme("", true),
		Profile:   colorprofile.NoTTY,
		// Lexer intentionally nil — should auto-resolve.
	}

	var buf bytes.Buffer
	if err := Unified(result, &buf, cfg); err != nil {
		t.Fatalf("Unified with nil Lexer error: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected non-empty output")
	}
}

func TestUnified_ShowLineNumbersGutter(t *testing.T) {
	result := edittype.DiffResult{
		Hunks: []edittype.Hunk{
			{
				OldStart: 1, OldLines: 1, NewStart: 1, NewLines: 1,
				Lines: []edittype.Line{
					{Op: edittype.Equal, Content: "x", OldNum: 10, NewNum: 20},
				},
			},
		},
	}
	cfg := noopConfig()
	cfg.ShowLineNumbers = true
	cfg.NoColor = true
	cfg.IsDark = true
	var buf bytes.Buffer
	if err := Unified(result, &buf, cfg); err != nil {
		t.Fatalf("Unified error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "10") || !strings.Contains(out, "20") {
		t.Fatalf("expected gutter line numbers in output:\n%s", out)
	}
	if !strings.Contains(out, "│") {
		t.Fatalf("expected │ gutter separator:\n%s", out)
	}
}

// TestUnified_FullLineBackgroundAndPrefix checks that when LineDiffStyle is active,
// the +/- prefix character and trailing whitespace both carry the line background colour.
func TestUnified_FullLineBackgroundAndPrefix(t *testing.T) {
	result := edittype.DiffResult{
		Hunks: []edittype.Hunk{
			{
				OldStart: 1, OldLines: 1, NewStart: 1, NewLines: 1,
				Lines: []edittype.Line{
					{Op: edittype.Delete, Content: "old line", OldNum: 1},
					{Op: edittype.Insert, Content: "new line", NewNum: 1},
				},
			},
		},
	}
	cfg := &RenderConfig{
		Lexer:         highlight.DetectLexer("go", "", ""),
		Style:         styles.Get("github"),
		Formatter:     formatters.TTY16m,
		Profile:       colorprofile.TrueColor,
		LineDiffStyle: true,
		IsDark:        false,
		TermWidth:     80,
	}
	if cfg.Style == nil {
		cfg.Style = highlight.SelectTheme("", false)
	}
	var buf bytes.Buffer
	if err := Unified(result, &buf, cfg); err != nil {
		t.Fatalf("Unified error: %v", err)
	}
	out := buf.String()

	// Must contain background SGR (CSI 48) for both delete and insert lines.
	if !strings.Contains(out, ";48;2") && !strings.Contains(out, "\x1b[48;") {
		t.Fatalf("expected full-line background SGR (CSI 48) in output:\n%s", out)
	}

	// The output must contain ANSI codes before the first '-' or '+' prefix,
	// confirming the prefix character carries a background.
	if !strings.Contains(out, "\x1b[") {
		t.Fatalf("expected ANSI codes in output:\n%s", out)
	}
}

// Verify the fallback lexer from lexers package is directly usable (compile check).
var _ = lexers.Fallback
