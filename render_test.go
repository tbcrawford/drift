package drift_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/tylercrawford/drift"
)

// TestRender_WithNoColor verifies that WithNoColor() produces plain text
// output with no ANSI escape sequences, regardless of color support.
func TestRender_WithNoColor(t *testing.T) {
	result, err := drift.Diff(
		"package main\nfunc old() {}\n",
		"package main\nfunc new() {}\n",
	)
	if err != nil {
		t.Fatalf("Diff error: %v", err)
	}

	var buf bytes.Buffer
	if err := drift.Render(result, &buf, drift.WithNoColor()); err != nil {
		t.Fatalf("Render error: %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Fatal("expected non-empty output for differing inputs")
	}
	if strings.Contains(output, "\033[") {
		t.Errorf("WithNoColor output contains ANSI codes:\n%s", output)
	}
}

// TestRender_PlainWriter verifies that writing to a bytes.Buffer (non-*os.File)
// produces plain text output (treated as NoTTY profile).
func TestRender_PlainWriter(t *testing.T) {
	result, err := drift.Diff(
		"line one\nline two\n",
		"line one\nline THREE\n",
	)
	if err != nil {
		t.Fatalf("Diff error: %v", err)
	}

	var buf bytes.Buffer
	if err := drift.Render(result, &buf); err != nil {
		t.Fatalf("Render error: %v", err)
	}

	output := buf.String()
	// bytes.Buffer is non-TTY → NoTTY profile → NoOp formatter → no ANSI codes.
	if strings.Contains(output, "\033[") {
		t.Errorf("non-file writer output contains ANSI codes:\n%s", output)
	}
	if !strings.Contains(output, "@@ ") {
		t.Errorf("output missing hunk header:\n%s", output)
	}
}

// TestRender_HunkHeaderFormat verifies the exact @@ -a,b +c,d @@ format.
func TestRender_HunkHeaderFormat(t *testing.T) {
	result, err := drift.Diff(
		"line one\nline two\nline three\n",
		"line one\nline TWO\nline three\n",
	)
	if err != nil {
		t.Fatalf("Diff error: %v", err)
	}

	var buf bytes.Buffer
	if err := drift.Render(result, &buf, drift.WithNoColor()); err != nil {
		t.Fatalf("Render error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "@@ -") {
		t.Errorf("output missing @@ hunk header:\n%s", output)
	}
	if !strings.Contains(output, "+line TWO") {
		t.Errorf("output missing '+line TWO' inserted line:\n%s", output)
	}
	if !strings.Contains(output, "-line two") {
		t.Errorf("output missing '-line two' deleted line:\n%s", output)
	}
}

// TestRender_EqualInputsNoOutput verifies that identical inputs produce empty output.
func TestRender_EqualInputsNoOutput(t *testing.T) {
	result, err := drift.Diff("same\ncontent\n", "same\ncontent\n")
	if err != nil {
		t.Fatalf("Diff error: %v", err)
	}
	if !result.IsEqual {
		t.Fatal("expected IsEqual=true for identical inputs")
	}

	var buf bytes.Buffer
	if err := drift.Render(result, &buf, drift.WithNoColor()); err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if buf.Len() != 0 {
		t.Errorf("expected empty output for equal inputs, got: %q", buf.String())
	}
}

// TestRender_WithLang verifies that WithLang overrides language detection.
// The test passes without error — correctness of lexer selection is verified
// in internal/highlight tests.
func TestRender_WithLang(t *testing.T) {
	result, err := drift.Diff(
		"def foo(): pass\n",
		"def foo(): return 1\n",
	)
	if err != nil {
		t.Fatalf("Diff error: %v", err)
	}

	var buf bytes.Buffer
	if err := drift.Render(result, &buf, drift.WithNoColor(), drift.WithLang("python")); err != nil {
		t.Fatalf("Render with WithLang error: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected non-empty output")
	}
}

// TestRender_WithTheme verifies that WithTheme does not cause an error.
// Visual correctness of the theme is a manual verification item.
func TestRender_WithTheme(t *testing.T) {
	result, err := drift.Diff("old\n", "new\n")
	if err != nil {
		t.Fatalf("Diff error: %v", err)
	}

	var buf bytes.Buffer
	if err := drift.Render(result, &buf, drift.WithNoColor(), drift.WithTheme("dracula")); err != nil {
		t.Fatalf("Render with WithTheme error: %v", err)
	}
}

// TestRenderWithNames verifies file header labels appear in output.
func TestRenderWithNames(t *testing.T) {
	result, err := drift.Diff("old content\n", "new content\n")
	if err != nil {
		t.Fatalf("Diff error: %v", err)
	}

	var buf bytes.Buffer
	if err := drift.RenderWithNames(result, &buf, "a/main.go", "b/main.go", drift.WithNoColor()); err != nil {
		t.Fatalf("RenderWithNames error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "--- a/main.go\n") {
		t.Errorf("output missing '--- a/main.go':\n%s", output)
	}
	if !strings.Contains(output, "+++ b/main.go\n") {
		t.Errorf("output missing '+++ b/main.go':\n%s", output)
	}
}

// TestRender_NoColorEnvVar verifies that when NO_COLOR is set, resolveProfile
// returns colorprofile.Ascii (via os.Getenv) even for a non-*os.File writer,
// so output has no ANSI escape codes.
func TestRender_NoColorEnvVar(t *testing.T) {
	t.Setenv("NO_COLOR", "1")

	result, err := drift.Diff("old\n", "new\n")
	if err != nil {
		t.Fatalf("Diff error: %v", err)
	}

	var buf bytes.Buffer
	if err := drift.Render(result, &buf); err != nil {
		t.Fatalf("Render error: %v", err)
	}

	output := buf.String()
	if strings.Contains(output, "\033[") {
		t.Errorf("NO_COLOR=1 output contains ANSI codes:\n%s", output)
	}
}
