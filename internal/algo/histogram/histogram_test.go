package histogram_test

import (
	"testing"

	"github.com/tbcrawford/drift/internal/algo/histogram"
	"github.com/tbcrawford/drift/internal/edittype"
)

// countOps counts edits of each Op type.
func countOps(edits []edittype.Edit) (equal, insert, del int) {
	for _, e := range edits {
		switch e.Op {
		case edittype.Equal:
			equal++
		case edittype.Insert:
			insert++
		case edittype.Delete:
			del++
		}
	}
	return
}

// applyEditsWithNew reconstructs the new file from edits, old, and new slices.
func applyEditsWithNew(edits []edittype.Edit, old, new []string) []string {
	var result []string
	for _, e := range edits {
		switch e.Op {
		case edittype.Equal:
			result = append(result, old[e.OldLine-1])
		case edittype.Insert:
			result = append(result, new[e.NewLine-1])
		case edittype.Delete:
			// skip
		}
	}
	return result
}

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

// TestBothEmpty — both empty returns non-nil empty slice.
func TestBothEmpty(t *testing.T) {
	edits := histogram.New().Diff([]string{}, []string{})
	if edits == nil {
		t.Fatal("expected non-nil empty slice, got nil")
	}
	if len(edits) != 0 {
		t.Fatalf("expected 0 edits, got %d", len(edits))
	}
}

// TestOldEmptyAllInserts — old empty → all inserts with correct line numbers.
func TestOldEmptyAllInserts(t *testing.T) {
	edits := histogram.New().Diff([]string{}, []string{"x", "y"})
	if len(edits) != 2 {
		t.Fatalf("expected 2 edits, got %d", len(edits))
	}
	for i, e := range edits {
		if e.Op != edittype.Insert {
			t.Errorf("edit[%d]: expected Insert, got %v", i, e.Op)
		}
		if e.OldLine != 0 {
			t.Errorf("edit[%d]: expected OldLine=0, got %d", i, e.OldLine)
		}
		if e.NewLine != i+1 {
			t.Errorf("edit[%d]: expected NewLine=%d, got %d", i, i+1, e.NewLine)
		}
	}
}

// TestNewEmptyAllDeletes — new empty → all deletes with correct line numbers.
func TestNewEmptyAllDeletes(t *testing.T) {
	edits := histogram.New().Diff([]string{"x", "y"}, []string{})
	if len(edits) != 2 {
		t.Fatalf("expected 2 edits, got %d", len(edits))
	}
	for i, e := range edits {
		if e.Op != edittype.Delete {
			t.Errorf("edit[%d]: expected Delete, got %v", i, e.Op)
		}
		if e.OldLine != i+1 {
			t.Errorf("edit[%d]: expected OldLine=%d, got %d", i, i+1, e.OldLine)
		}
		if e.NewLine != 0 {
			t.Errorf("edit[%d]: expected NewLine=0, got %d", i, e.NewLine)
		}
	}
}

// TestIdenticalInputs — identical inputs return all Equal edits.
func TestIdenticalInputs(t *testing.T) {
	edits := histogram.New().Diff([]string{"a", "b", "c"}, []string{"a", "b", "c"})
	if len(edits) != 3 {
		t.Fatalf("expected 3 edits, got %d", len(edits))
	}
	for i, e := range edits {
		if e.Op != edittype.Equal {
			t.Errorf("edit[%d]: expected Equal, got %v", i, e.Op)
		}
	}
}

// TestLineInvariant_Histogram — table-driven: Equal+Delete==len(old), Equal+Insert==len(new).
func TestLineInvariant_Histogram(t *testing.T) {
	h := histogram.New()
	cases := []struct {
		name string
		old  []string
		new  []string
	}{
		{"identical", []string{"a", "b", "c"}, []string{"a", "b", "c"}},
		{"both empty", []string{}, []string{}},
		{"old empty", []string{}, []string{"x", "y"}},
		{"new empty", []string{"x", "y"}, []string{}},
		{"simple insert", []string{"a", "b", "d"}, []string{"a", "b", "c", "d"}},
		{"simple delete", []string{"a", "b", "c", "d"}, []string{"a", "b", "d"}},
		{"all replaced", []string{"a", "b", "c"}, []string{"x", "y", "z"}},
		{"repetitive", []string{"}", "}", "}"}, []string{"}", "x", "}"}},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			edits := h.Diff(tc.old, tc.new)
			eq, ins, del := countOps(edits)

			if eq+del != len(tc.old) {
				t.Errorf("old invariant violated: Equal(%d)+Delete(%d)=%d != len(old)=%d",
					eq, del, eq+del, len(tc.old))
			}
			if eq+ins != len(tc.new) {
				t.Errorf("new invariant violated: Equal(%d)+Insert(%d)=%d != len(new)=%d",
					eq, ins, eq+ins, len(tc.new))
			}
		})
	}
}

// TestHistogramFallback_AllIdenticalLines — 200 identical lines triggers Myers fallback.
// Verifies no panic and line invariant holds.
func TestHistogramFallback_AllIdenticalLines(t *testing.T) {
	old := make([]string, 200)
	for i := range old {
		old[i] = "}"
	}
	newLines := make([]string, 200)
	copy(newLines, old)
	newLines[100] = "// inserted"

	edits := histogram.New().Diff(old, newLines)
	eq, ins, del := countOps(edits)

	if eq+del != len(old) {
		t.Errorf("old invariant violated: Equal(%d)+Delete(%d)=%d != len(old)=%d",
			eq, del, eq+del, len(old))
	}
	if eq+ins != len(newLines) {
		t.Errorf("new invariant violated: Equal(%d)+Insert(%d)=%d != len(new)=%d",
			eq, ins, eq+ins, len(newLines))
	}
}

// TestRoundTrip_Histogram — apply(Diff(old,new), old) == new for repetitive Go function file.
func TestRoundTrip_Histogram(t *testing.T) {
	old := []string{
		"func A() {",
		"    doThing1()",
		"}",
		"",
		"func B() {",
		"    doThing2()",
		"}",
	}
	new := []string{
		"func A() {",
		"    doThing1()",
		"    doThing3()",
		"}",
		"",
		"func B() {",
		"    doThing2()",
		"}",
	}

	edits := histogram.New().Diff(old, new)
	got := applyEditsWithNew(edits, old, new)

	if !slicesEqual(got, new) {
		t.Errorf("round-trip failed:\n  got:  %v\n  want: %v", got, new)
	}
}

// TestHistogramFallback_NoPanic — pathological input where every line appears > 64 times.
// Verifies no panic and line invariant holds.
func TestHistogramFallback_NoPanic(t *testing.T) {
	old := make([]string, 65)
	for i := range old {
		old[i] = "x"
	}
	newLines := make([]string, 65)
	copy(newLines, old)
	// Swap first and last element
	newLines[0], newLines[64] = newLines[64], newLines[0]

	edits := histogram.New().Diff(old, newLines)
	eq, ins, del := countOps(edits)

	if eq+del != len(old) {
		t.Errorf("old invariant violated: Equal(%d)+Delete(%d)=%d != len(old)=%d",
			eq, del, eq+del, len(old))
	}
	if eq+ins != len(newLines) {
		t.Errorf("new invariant violated: Equal(%d)+Insert(%d)=%d != len(new)=%d",
			eq, ins, eq+ins, len(newLines))
	}
}
