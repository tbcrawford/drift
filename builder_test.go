package drift

import (
	"bytes"
	"reflect"
	"testing"
)

func TestBuilderDiffParity(t *testing.T) {
	old := "alpha\nbeta\ngamma\n"
	newText := "alpha\nbeta\ndelta\n"

	want, err := Diff(old, newText, WithAlgorithm(Patience), WithTheme("monokai"))
	if err != nil {
		t.Fatalf("Diff with options: %v", err)
	}
	got, err := New().Algorithm(Patience).Theme("monokai").Diff(old, newText)
	if err != nil {
		t.Fatalf("Builder.Diff: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("DiffResult mismatch\ngot:  %+v\nwant: %+v", got, want)
	}
}

func TestBuilderRenderParity(t *testing.T) {
	old := "package main\n\nfunc main() {}\n"
	newText := "package main\n\nfunc main() { println() }\n"

	result, err := Diff(old, newText)
	if err != nil {
		t.Fatalf("Diff: %v", err)
	}

	var bufOpts, bufBuilder bytes.Buffer
	if err := Render(result, &bufOpts, WithNoColor()); err != nil {
		t.Fatalf("Render with options: %v", err)
	}
	if err := New().NoColor().Render(result, &bufBuilder); err != nil {
		t.Fatalf("Builder.Render: %v", err)
	}
	if !bytes.Equal(bufBuilder.Bytes(), bufOpts.Bytes()) {
		t.Fatalf("rendered output mismatch\nopts:    %q\nbuilder: %q", bufOpts.String(), bufBuilder.String())
	}
}

// --- Builder API coverage tests ---

func TestBuilder_Context(t *testing.T) {
	old := "a\nb\nc\nd\ne\n"
	newText := "a\nb\nX\nd\ne\n"

	wantResult, err := Diff(old, newText, WithContext(0))
	if err != nil {
		t.Fatalf("Diff: %v", err)
	}
	gotResult, err := New().Context(0).Diff(old, newText)
	if err != nil {
		t.Fatalf("Builder.Context.Diff: %v", err)
	}
	if !reflect.DeepEqual(gotResult, wantResult) {
		t.Fatalf("Context(0) mismatch\ngot:  %+v\nwant: %+v", gotResult, wantResult)
	}
}

func TestBuilder_Lang(t *testing.T) {
	old := "x := 1\n"
	newText := "x := 2\n"

	result, err := New().Lang("go").NoColor().Diff(old, newText)
	if err != nil {
		t.Fatalf("Builder.Lang.Diff: %v", err)
	}
	var buf bytes.Buffer
	if err := New().Lang("go").NoColor().Render(result, &buf); err != nil {
		t.Fatalf("Builder.Lang.Render: %v", err)
	}
	if buf.Len() == 0 {
		t.Fatal("expected non-empty output")
	}
}

func TestBuilder_ThemeResolved(t *testing.T) {
	old := "hello\n"
	newText := "world\n"

	var resolved string
	result, err := New().ThemeResolved(func(name string) { resolved = name }).Diff(old, newText)
	if err != nil {
		t.Fatalf("Builder.ThemeResolved.Diff: %v", err)
	}
	var buf bytes.Buffer
	if err := New().ThemeResolved(func(name string) { resolved = name }).NoColor().Render(result, &buf); err != nil {
		t.Fatalf("Builder.ThemeResolved.Render: %v", err)
	}
	if resolved == "" {
		t.Fatal("ThemeResolved callback was not invoked")
	}
}

func TestBuilder_Split(t *testing.T) {
	old := "foo\n"
	newText := "bar\n"

	result, err := New().Split().NoColor().Diff(old, newText)
	if err != nil {
		t.Fatalf("Builder.Split.Diff: %v", err)
	}
	var buf bytes.Buffer
	if err := New().Split().NoColor().Render(result, &buf); err != nil {
		t.Fatalf("Builder.Split.Render: %v", err)
	}
	// Split output contains │ separator
	if !bytes.Contains(buf.Bytes(), []byte("│")) {
		t.Fatalf("split output missing │ separator: %q", buf.String())
	}
}

func TestBuilder_LineNumbers(t *testing.T) {
	old := "alpha\n"
	newText := "beta\n"

	result, err := New().LineNumbers(true).NoColor().Diff(old, newText)
	if err != nil {
		t.Fatalf("Builder.LineNumbers.Diff: %v", err)
	}
	var buf bytes.Buffer
	if err := New().LineNumbers(true).NoColor().Render(result, &buf); err != nil {
		t.Fatalf("Builder.LineNumbers.Render: %v", err)
	}
	if buf.Len() == 0 {
		t.Fatal("expected non-empty output with line numbers")
	}
}

func TestBuilder_WithoutLineNumbers(t *testing.T) {
	old := "alpha\n"
	newText := "beta\n"

	result, err := New().NoColor().Diff(old, newText)
	if err != nil {
		t.Fatalf("Diff: %v", err)
	}

	var bufWith, bufWithout bytes.Buffer
	if err := New().LineNumbers(true).NoColor().Render(result, &bufWith); err != nil {
		t.Fatalf("Render with line numbers: %v", err)
	}
	if err := New().WithoutLineNumbers().NoColor().Render(result, &bufWithout); err != nil {
		t.Fatalf("Builder.WithoutLineNumbers.Render: %v", err)
	}
	// Output without line numbers should differ (shorter or different)
	if bytes.Equal(bufWith.Bytes(), bufWithout.Bytes()) {
		t.Fatal("expected different output when line numbers are toggled")
	}
}

func TestBuilder_LineDiffStyle(t *testing.T) {
	old := "foo\n"
	newText := "bar\n"

	result, err := New().LineDiffStyle(false).NoColor().Diff(old, newText)
	if err != nil {
		t.Fatalf("Builder.LineDiffStyle.Diff: %v", err)
	}
	var buf bytes.Buffer
	if err := New().LineDiffStyle(false).NoColor().Render(result, &buf); err != nil {
		t.Fatalf("Builder.LineDiffStyle.Render: %v", err)
	}
	if buf.Len() == 0 {
		t.Fatal("expected non-empty output")
	}
}

func TestBuilder_WordDiff(t *testing.T) {
	old := "hello world\n"
	newText := "hello earth\n"

	result, err := New().WordDiff(true).NoColor().Diff(old, newText)
	if err != nil {
		t.Fatalf("Builder.WordDiff.Diff: %v", err)
	}
	var buf bytes.Buffer
	if err := New().WordDiff(true).NoColor().Render(result, &buf); err != nil {
		t.Fatalf("Builder.WordDiff.Render: %v", err)
	}
	if buf.Len() == 0 {
		t.Fatal("expected non-empty output")
	}
}

func TestBuilder_RenderWithNames(t *testing.T) {
	old := "before\n"
	newText := "after\n"

	result, err := New().NoColor().Diff(old, newText)
	if err != nil {
		t.Fatalf("Diff: %v", err)
	}
	var buf bytes.Buffer
	if err := New().NoColor().RenderWithNames(result, &buf, "old.txt", "new.txt"); err != nil {
		t.Fatalf("Builder.RenderWithNames: %v", err)
	}
	out := buf.String()
	if !bytes.Contains([]byte(out), []byte("old.txt")) {
		t.Errorf("expected 'old.txt' in output: %q", out)
	}
	if !bytes.Contains([]byte(out), []byte("new.txt")) {
		t.Errorf("expected 'new.txt' in output: %q", out)
	}
}
