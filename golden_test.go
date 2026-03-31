package drift_test

import (
	"bytes"
	"testing"

	"github.com/sebdah/goldie/v2"
	"github.com/tylercrawford/drift"
)

// goOldSrc and goNewSrc are fixed Go source snippets used as stable golden
// test inputs. They exercise real rendering paths: gutter numbers, hunk
// headers, deletion/insertion line prefixes, and syntax highlighting paths.
// All tests use WithNoColor() for plain-text, CI-portable fixtures.
const goOldSrc = `package main

import "fmt"

func greet(name string) string {
	return fmt.Sprintf("Hello, %s!", name)
}

func main() {
	fmt.Println(greet("world"))
}
`

const goNewSrc = `package main

import (
	"fmt"
	"strings"
)

func greet(name string) string {
	return fmt.Sprintf("Hello, %s!", strings.TrimSpace(name))
}

func main() {
	fmt.Println(greet("world"))
}
`

// TestGolden_UnifiedRenderer snapshots unified diff output for a known Go
// source diff. Re-generate with: go test -run TestGolden_UnifiedRenderer -update .
func TestGolden_UnifiedRenderer(t *testing.T) {
	g := goldie.New(t, goldie.WithFixtureDir("testdata/golden"))
	result, err := drift.Diff(goOldSrc, goNewSrc)
	if err != nil {
		t.Fatalf("Diff: %v", err)
	}
	var buf bytes.Buffer
	if err := drift.Render(result, &buf, drift.WithNoColor(), drift.WithLang("go")); err != nil {
		t.Fatalf("Render: %v", err)
	}
	g.Assert(t, "unified_go", buf.Bytes())
}

// TestGolden_SplitRenderer snapshots split diff output (120-col, no-color).
// Re-generate with: go test -run TestGolden_SplitRenderer -update .
func TestGolden_SplitRenderer(t *testing.T) {
	g := goldie.New(t, goldie.WithFixtureDir("testdata/golden"))
	result, err := drift.Diff(goOldSrc, goNewSrc)
	if err != nil {
		t.Fatalf("Diff: %v", err)
	}
	var buf bytes.Buffer
	if err := drift.Render(result, &buf,
		drift.WithNoColor(),
		drift.WithLang("go"),
		drift.WithSplit(),
		drift.WithTermWidth(120),
	); err != nil {
		t.Fatalf("Render: %v", err)
	}
	g.Assert(t, "split_go", buf.Bytes())
}

// TestGolden_NoColorOutput snapshots plain-text unified diff for a minimal
// change. Verifies the no-color path produces consistent, ANSI-free output.
// Re-generate with: go test -run TestGolden_NoColorOutput -update .
func TestGolden_NoColorOutput(t *testing.T) {
	g := goldie.New(t, goldie.WithFixtureDir("testdata/golden"))
	result, err := drift.Diff(
		"line one\nline two\nline three\n",
		"line one\nline TWO\nline three\n",
	)
	if err != nil {
		t.Fatalf("Diff: %v", err)
	}
	var buf bytes.Buffer
	if err := drift.Render(result, &buf, drift.WithNoColor()); err != nil {
		t.Fatalf("Render: %v", err)
	}
	g.Assert(t, "nocolor_basic", buf.Bytes())
}
