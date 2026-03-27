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
