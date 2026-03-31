package myers_test

import (
	"strings"
	"testing"

	"github.com/tbcrawford/drift/internal/algo/myers"
	"github.com/tbcrawford/drift/internal/edittype"
)

// FuzzMyers verifies that the Myers diff algorithm never panics for any input
// and that the edit sequence it produces is structurally valid.
//
// The fuzzer discovers pathological inputs: long identical runs, lines with
// Unicode, very long lines, binary-looking content, etc.
//
// Run with: go test -fuzz=FuzzMyers -fuzztime=30s ./internal/algo/myers/...
func FuzzMyers(f *testing.F) {
	// Seed corpus: representative edge cases.
	f.Add("", "")
	f.Add("", "hello")
	f.Add("hello", "")
	f.Add("hello", "hello")
	f.Add("a\nb\nc", "a\nb\nc")
	f.Add("a\nb\nc", "a\nB\nc")
	f.Add("a\nb", "a\nb\nc\nd\ne")
	f.Add("a\nb\nc\nd\ne", "c\nd\ne")
	f.Add("line1\nline2\nline3", "line1\nlineX\nline3")
	f.Add("x\ny\nz\n", "x\nz\n")

	m := myers.New()

	f.Fuzz(func(t *testing.T, old, new string) {
		oldLines := strings.Split(old, "\n")
		newLines := strings.Split(new, "\n")

		// Must not panic.
		edits := m.Diff(oldLines, newLines)

		// Structural validity: verify edit sequence is well-formed.
		verifyEdits(t, edits, oldLines, newLines)
	})
}

// verifyEdits checks structural invariants on the edit sequence produced by Myers:
//
//  1. Every Delete edit references a valid old line number (1-indexed).
//  2. Every Insert edit references a valid new line number (1-indexed).
//  3. Every Equal edit references valid old and new line numbers.
//  4. Applying the edits reconstructs newLines from oldLines (round-trip).
func verifyEdits(t *testing.T, edits []edittype.Edit, oldLines, newLines []string) {
	t.Helper()

	for i, e := range edits {
		switch e.Op {
		case edittype.Delete:
			if e.OldLine < 1 || e.OldLine > len(oldLines) {
				t.Fatalf("edit[%d]: Delete has out-of-range OldLine=%d (len=%d)",
					i, e.OldLine, len(oldLines))
			}
		case edittype.Insert:
			if e.NewLine < 1 || e.NewLine > len(newLines) {
				t.Fatalf("edit[%d]: Insert has out-of-range NewLine=%d (len=%d)",
					i, e.NewLine, len(newLines))
			}
		case edittype.Equal:
			if e.OldLine < 1 || e.OldLine > len(oldLines) {
				t.Fatalf("edit[%d]: Equal has out-of-range OldLine=%d (len=%d)",
					i, e.OldLine, len(oldLines))
			}
			if e.NewLine < 1 || e.NewLine > len(newLines) {
				t.Fatalf("edit[%d]: Equal has out-of-range NewLine=%d (len=%d)",
					i, e.NewLine, len(newLines))
			}
			// Equal lines must have the same content.
			if oldLines[e.OldLine-1] != newLines[e.NewLine-1] {
				t.Fatalf("edit[%d]: Equal but content differs: old[%d]=%q new[%d]=%q",
					i, e.OldLine-1, oldLines[e.OldLine-1], e.NewLine-1, newLines[e.NewLine-1])
			}
		}
	}

	// Round-trip: reconstruct newLines from edits and verify.
	got := applyEdits(edits, newLines)
	if !slicesEqual(got, newLines) {
		t.Fatalf("round-trip failed: got %v, want %v (edits: %v)", got, newLines, edits)
	}
}

// applyEdits reconstructs newLines by walking the flat edit sequence directly.
// Unlike the hunk-based Apply in testdata, this works at the edit level.
func applyEdits(edits []edittype.Edit, newLines []string) []string {
	out := make([]string, 0, len(newLines))
	for _, e := range edits {
		switch e.Op {
		case edittype.Equal:
			out = append(out, newLines[e.NewLine-1])
		case edittype.Insert:
			out = append(out, newLines[e.NewLine-1])
			// Delete: skip (don't append)
		}
	}
	return out
}

// slicesEqual returns true if two string slices have identical contents.
func slicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
