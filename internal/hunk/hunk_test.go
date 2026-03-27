package hunk_test

import (
	"testing"

	"github.com/tylercrawford/drift/drift"
	"github.com/tylercrawford/drift/internal/hunk"
)

// makeLines builds a simple []string slice from a variadic list of strings.
func makeLines(lines ...string) []string { return lines }

// numberedLines builds lines like "line01", "line02", …
func numberedLines(n int) []string {
	lines := make([]string, n)
	for i := range lines {
		lines[i] = "line" + itoa(i+1)
	}
	return lines
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	digits := []byte{}
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}

// TestBuild_NoChanges verifies that identical inputs produce no hunks.
func TestBuild_NoChanges(t *testing.T) {
	lines := makeLines("a", "b", "c")
	edits := []drift.Edit{
		{Op: drift.Equal, OldLine: 1, NewLine: 1},
		{Op: drift.Equal, OldLine: 2, NewLine: 2},
		{Op: drift.Equal, OldLine: 3, NewLine: 3},
	}

	hunks := hunk.Build(edits, lines, lines, 3)

	if len(hunks) != 0 {
		t.Errorf("expected 0 hunks for identical input, got %d", len(hunks))
	}
}

// TestBuild_SingleChangeMiddle verifies a single change in the middle of a 7-line file
// with context=3. The whole file fits within context, so we get 1 hunk covering all 7 lines.
func TestBuild_SingleChangeMiddle(t *testing.T) {
	old := makeLines("1", "2", "3", "4", "5", "6", "7")
	newL := makeLines("1", "2", "3", "X", "5", "6", "7")

	edits := []drift.Edit{
		{Op: drift.Equal, OldLine: 1, NewLine: 1},
		{Op: drift.Equal, OldLine: 2, NewLine: 2},
		{Op: drift.Equal, OldLine: 3, NewLine: 3},
		{Op: drift.Delete, OldLine: 4, NewLine: 0},
		{Op: drift.Insert, OldLine: 0, NewLine: 4},
		{Op: drift.Equal, OldLine: 5, NewLine: 5},
		{Op: drift.Equal, OldLine: 6, NewLine: 6},
		{Op: drift.Equal, OldLine: 7, NewLine: 7},
	}

	hunks := hunk.Build(edits, old, newL, 3)

	if len(hunks) != 1 {
		t.Fatalf("expected 1 hunk, got %d", len(hunks))
	}
	h := hunks[0]
	if h.OldStart != 1 {
		t.Errorf("OldStart: got %d, want 1", h.OldStart)
	}
	if h.OldLines != 7 {
		t.Errorf("OldLines: got %d, want 7", h.OldLines)
	}
	if h.NewStart != 1 {
		t.Errorf("NewStart: got %d, want 1", h.NewStart)
	}
	if h.NewLines != 7 {
		t.Errorf("NewLines: got %d, want 7", h.NewLines)
	}
}

// TestBuild_SingleChangeAtStart verifies a change at line 1 with context=3.
// Context window after the change covers through line 4.
func TestBuild_SingleChangeAtStart(t *testing.T) {
	old := makeLines("old", "b", "c", "d", "e", "f")
	newL := makeLines("new", "b", "c", "d", "e", "f")

	edits := []drift.Edit{
		{Op: drift.Delete, OldLine: 1, NewLine: 0},
		{Op: drift.Insert, OldLine: 0, NewLine: 1},
		{Op: drift.Equal, OldLine: 2, NewLine: 2},
		{Op: drift.Equal, OldLine: 3, NewLine: 3},
		{Op: drift.Equal, OldLine: 4, NewLine: 4},
		{Op: drift.Equal, OldLine: 5, NewLine: 5},
		{Op: drift.Equal, OldLine: 6, NewLine: 6},
	}

	hunks := hunk.Build(edits, old, newL, 3)

	if len(hunks) != 1 {
		t.Fatalf("expected 1 hunk, got %d", len(hunks))
	}
	h := hunks[0]
	if h.OldStart != 1 {
		t.Errorf("OldStart: got %d, want 1", h.OldStart)
	}
	if h.NewStart != 1 {
		t.Errorf("NewStart: got %d, want 1", h.NewStart)
	}
	// Should include line 1 (changed) + 3 context lines after = lines 1-4
	// OldLines = 1 delete + 3 equal = 4; NewLines = 1 insert + 3 equal = 4
	if h.OldLines != 4 {
		t.Errorf("OldLines: got %d, want 4", h.OldLines)
	}
	if h.NewLines != 4 {
		t.Errorf("NewLines: got %d, want 4", h.NewLines)
	}
}

// TestBuild_TwoDistantChanges verifies that changes far apart produce two hunks.
// 20 lines; change at line 2 and line 19; context=3.
// First hunk covers lines 1-5 (1 context before + changed line 2 + 3 context after).
// Second hunk covers lines 16-20 (3 context before + changed line 19 + 1 context after).
func TestBuild_TwoDistantChanges(t *testing.T) {
	old := numberedLines(20)
	newL := numberedLines(20)
	newL[1] = "CHANGED_2"   // line 2
	newL[18] = "CHANGED_19" // line 19

	// Build edits manually: equal except at positions 2 and 19
	edits := make([]drift.Edit, 0, 22)
	for i := 1; i <= 20; i++ {
		if i == 2 || i == 19 {
			edits = append(edits,
				drift.Edit{Op: drift.Delete, OldLine: i, NewLine: 0},
				drift.Edit{Op: drift.Insert, OldLine: 0, NewLine: i},
			)
		} else {
			edits = append(edits, drift.Edit{Op: drift.Equal, OldLine: i, NewLine: i})
		}
	}

	hunks := hunk.Build(edits, old, newL, 3)

	if len(hunks) != 2 {
		t.Fatalf("expected 2 hunks, got %d", len(hunks))
	}

	h1 := hunks[0]
	h2 := hunks[1]

	// First hunk: lines 1-5 (line 2 change ± 3 context, clamped at line 1)
	if h1.OldStart != 1 {
		t.Errorf("hunk1 OldStart: got %d, want 1", h1.OldStart)
	}
	// Last equal line in hunk1 should be line 5 (line 2 + 3 context)
	if h1.OldStart+h1.OldLines-1 < 5 {
		t.Errorf("hunk1 should extend at least to line 5, ends at %d", h1.OldStart+h1.OldLines-1)
	}
	// Hunk1 must end before line 16 (where hunk2 starts)
	hunk1End := h1.OldStart + h1.OldLines - 1
	if hunk1End >= 16 {
		t.Errorf("hunk1 should end before line 16, ends at line %d", hunk1End)
	}

	// Second hunk: starts at line 16 (line 19 - 3 context)
	if h2.OldStart != 16 {
		t.Errorf("hunk2 OldStart: got %d, want 16", h2.OldStart)
	}
}

// TestBuild_TwoNearbyChanges verifies that nearby changes are merged into one hunk.
// 10 lines; changes at line 3 and line 7; context=3.
// Context window for line 3 ends at line 6; context window for line 7 starts at line 4.
// They overlap (4 <= 6), so they must be merged into a single hunk.
func TestBuild_TwoNearbyChanges(t *testing.T) {
	old := numberedLines(10)
	newL := numberedLines(10)
	newL[2] = "CHANGED_3" // line 3
	newL[6] = "CHANGED_7" // line 7

	edits := make([]drift.Edit, 0, 12)
	for i := 1; i <= 10; i++ {
		if i == 3 || i == 7 {
			edits = append(edits,
				drift.Edit{Op: drift.Delete, OldLine: i, NewLine: 0},
				drift.Edit{Op: drift.Insert, OldLine: 0, NewLine: i},
			)
		} else {
			edits = append(edits, drift.Edit{Op: drift.Equal, OldLine: i, NewLine: i})
		}
	}

	hunks := hunk.Build(edits, old, newL, 3)

	if len(hunks) != 1 {
		t.Fatalf("expected 1 merged hunk, got %d", len(hunks))
	}
}

// TestBuild_ZeroContext verifies that WithContext(0) produces hunks with only changed lines.
func TestBuild_ZeroContext(t *testing.T) {
	old := makeLines("a", "b", "c", "d", "e")
	newL := makeLines("a", "B", "c", "d", "e")

	edits := []drift.Edit{
		{Op: drift.Equal, OldLine: 1, NewLine: 1},
		{Op: drift.Delete, OldLine: 2, NewLine: 0},
		{Op: drift.Insert, OldLine: 0, NewLine: 2},
		{Op: drift.Equal, OldLine: 3, NewLine: 3},
		{Op: drift.Equal, OldLine: 4, NewLine: 4},
		{Op: drift.Equal, OldLine: 5, NewLine: 5},
	}

	hunks := hunk.Build(edits, old, newL, 0)

	if len(hunks) != 1 {
		t.Fatalf("expected 1 hunk, got %d", len(hunks))
	}

	h := hunks[0]
	for _, line := range h.Lines {
		if line.Op == drift.Equal {
			t.Errorf("expected no Equal lines with context=0, found Equal line: %q", line.Content)
		}
	}
}
