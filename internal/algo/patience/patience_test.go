package patience_test

import (
	"testing"

	"github.com/tbcrawford/drift/internal/algo/patience"
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

// applyEditsWithNew reconstructs the new file using both old and new line content.
func applyEditsWithNew(edits []edittype.Edit, old, new []string) []string {
	result := make([]string, 0, len(new))
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

// slicesEqual reports whether two string slices are element-wise equal.
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

// TestBothEmpty — both empty returns empty non-nil slice.
func TestBothEmpty(t *testing.T) {
	p := patience.New()
	edits := p.Diff([]string{}, []string{})
	if edits == nil {
		t.Fatal("expected non-nil empty slice, got nil")
	}
	if len(edits) != 0 {
		t.Fatalf("expected 0 edits, got %d", len(edits))
	}
}

// TestOldEmptyAllInserts — old empty → all Insert edits with correct line numbers.
func TestOldEmptyAllInserts(t *testing.T) {
	p := patience.New()
	edits := p.Diff([]string{}, []string{"x", "y"})
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

// TestNewEmptyAllDeletes — new empty → all Delete edits with correct line numbers.
func TestNewEmptyAllDeletes(t *testing.T) {
	p := patience.New()
	edits := p.Diff([]string{"x", "y"}, []string{})
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

// TestIdenticalInputs — identical inputs → all Equal edits.
func TestIdenticalInputs(t *testing.T) {
	p := patience.New()
	edits := p.Diff([]string{"a", "b", "c"}, []string{"a", "b", "c"})
	if len(edits) != 3 {
		t.Fatalf("expected 3 edits, got %d", len(edits))
	}
	for i, e := range edits {
		if e.Op != edittype.Equal {
			t.Errorf("edit[%d]: expected Equal, got %v", i, e.Op)
		}
		if e.OldLine != i+1 {
			t.Errorf("edit[%d]: expected OldLine=%d, got %d", i, i+1, e.OldLine)
		}
		if e.NewLine != i+1 {
			t.Errorf("edit[%d]: expected NewLine=%d, got %d", i, i+1, e.NewLine)
		}
	}
}

// TestLineInvariant_Patience — table-driven line invariant: Equal+Delete==len(old), Equal+Insert==len(new).
func TestLineInvariant_Patience(t *testing.T) {
	p := patience.New()
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
		{"mixed", []string{"a", "b", "c", "d", "e"}, []string{"a", "x", "c", "y", "e"}},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			edits := p.Diff(tc.old, tc.new)
			eq, ins, del := countOps(edits)

			if eq+del != len(tc.old) {
				t.Errorf("old invariant: Equal(%d)+Delete(%d)=%d != len(old)=%d",
					eq, del, eq+del, len(tc.old))
			}
			if eq+ins != len(tc.new) {
				t.Errorf("new invariant: Equal(%d)+Insert(%d)=%d != len(new)=%d",
					eq, ins, eq+ins, len(tc.new))
			}
		})
	}
}

// TestPatienceFallback_NoUniqueLines — no panic and correct result when no unique lines exist.
func TestPatienceFallback_NoUniqueLines(t *testing.T) {
	p := patience.New()
	old := []string{"a", "a", "a", "b", "b"}
	new := []string{"b", "b", "a", "a", "a"}
	edits := p.Diff(old, new)

	eq, ins, del := countOps(edits)
	if eq+del != len(old) {
		t.Errorf("old invariant violated: Equal(%d)+Delete(%d)=%d != %d", eq, del, eq+del, len(old))
	}
	if eq+ins != len(new) {
		t.Errorf("new invariant violated: Equal(%d)+Insert(%d)=%d != %d", eq, ins, eq+ins, len(new))
	}
}

// TestRoundTrip_Patience — apply(Diff(old, new), old) == new for a known case.
func TestRoundTrip_Patience(t *testing.T) {
	p := patience.New()
	old := []string{"a", "b", "c", "d"}
	new := []string{"a", "c", "b", "e"}
	edits := p.Diff(old, new)

	got := applyEditsWithNew(edits, old, new)
	if !slicesEqual(got, new) {
		t.Errorf("round-trip failed:\n  old: %v\n  new: %v\n  got: %v", old, new, got)
		for i, e := range edits {
			t.Logf("  edit[%d]: Op=%v OldLine=%d NewLine=%d", i, e.Op, e.OldLine, e.NewLine)
		}
	}
}

// TestPatienceSuperiority_FunctionMove — canonical patience vs myers test case from research section 6.1.
// Verifies correctness (apply round-trip) on the C code function-move example.
func TestPatienceSuperiority_FunctionMove(t *testing.T) {
	p := patience.New()

	oldLines := []string{
		"#include <stdio.h>",
		"",
		"// Frobs foo heartily",
		"int frobnitz(int foo) {",
		"    int i;",
		"    for(i = 0; i < 10; i++) {",
		"        printf(\"Your answer is: \");",
		"        printf(\"%d\\n\", foo);",
		"    }",
		"}",
		"",
		"int fact(int n) {",
		"    if(n > 1) {",
		"        return fact(n-1) * n;",
		"    }",
		"    return 1;",
		"}",
		"",
		"int main(int argc, char **argv) {",
		"    frobnitz(fact(10));",
		"}",
	}

	newLines := []string{
		"#include <stdio.h>",
		"",
		"// Frobs foo heartily",
		"int frobnitz(int foo) {",
		"    int i;",
		"    for(i = 0; i < 10; i++) {",
		"        printf(\"%d\\n\", foo);",
		"    }",
		"}",
		"",
		"int fib(int n) {",
		"    if(n > 2) {",
		"        return fib(n-1) + fib(n-2);",
		"    }",
		"    return 1;",
		"}",
		"",
		"int main(int argc, char **argv) {",
		"    frobnitz(fib(10));",
		"}",
	}

	edits := p.Diff(oldLines, newLines)

	// Correctness: apply round-trip
	got := applyEditsWithNew(edits, oldLines, newLines)
	if !slicesEqual(got, newLines) {
		t.Errorf("round-trip failed on function-move example")
		for i, e := range edits {
			t.Logf("  edit[%d]: Op=%v OldLine=%d NewLine=%d", i, e.Op, e.OldLine, e.NewLine)
		}
	}

	// Invariant check
	eq, ins, del := countOps(edits)
	if eq+del != len(oldLines) {
		t.Errorf("old invariant: Equal(%d)+Delete(%d)=%d != %d", eq, del, eq+del, len(oldLines))
	}
	if eq+ins != len(newLines) {
		t.Errorf("new invariant: Equal(%d)+Insert(%d)=%d != %d", eq, ins, eq+ins, len(newLines))
	}
}
