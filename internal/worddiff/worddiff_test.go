package worddiff

import (
	"testing"
)

func TestPairSegments_identical(t *testing.T) {
	t.Helper()
	o, n := PairSegments("hello world", "hello world")
	if len(o) != 1 || len(n) != 1 {
		t.Fatalf("want 1 segment each, got old=%v new=%v", o, n)
	}
	if o[0].Changed || n[0].Changed {
		t.Fatal("identical strings should not mark changed")
	}
}

func TestPairSegments_substitution(t *testing.T) {
	t.Helper()
	oldSegs, newSegs := PairSegments("foo bar baz", "foo qux baz")
	t.Logf("old=%v new=%v", oldSegs, newSegs)
	// " bar " vs " qux " should be changed regions; foo and baz unchanged.
	var oldChanged, newChanged int
	for _, s := range oldSegs {
		if s.Changed {
			oldChanged++
		}
	}
	for _, s := range newSegs {
		if s.Changed {
			newChanged++
		}
	}
	if oldChanged == 0 || newChanged == 0 {
		t.Fatalf("expected changed segments on both sides, old=%v new=%v", oldSegs, newSegs)
	}
}

func TestPairSegments_empty(t *testing.T) {
	t.Helper()
	o, n := PairSegments("", "")
	if o != nil || n != nil {
		t.Fatalf("want nil, got old=%v new=%v", o, n)
	}
}

func TestPairCharSegments_singleCharChange(t *testing.T) {
	t.Parallel()
	// Only the 'f' vs 'b' should be changed; 'oo' is common.
	oldSegs, newSegs := PairCharSegments("foo", "boo")
	if len(oldSegs) == 0 || len(newSegs) == 0 {
		t.Fatalf("expected segments, got old=%v new=%v", oldSegs, newSegs)
	}
	// The changed span on the old side must cover exactly 'f' (1 byte at offset 0).
	var found bool
	for _, s := range oldSegs {
		if s.Changed && s.Start == 0 && s.End == 1 {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected old changed span at [0,1), got %v", oldSegs)
	}
}

func TestPairCharSegments_identical(t *testing.T) {
	t.Parallel()
	o, n := PairCharSegments("abc", "abc")
	if len(o) != 1 || o[0].Changed {
		t.Fatalf("identical: want 1 unchanged segment, got %v", o)
	}
	_ = n
}

func TestPairCharSegments_unicode(t *testing.T) {
	t.Parallel()
	// '日' is 3 bytes; verify offsets are byte-accurate.
	oldSegs, _ := PairCharSegments("日本", "日x")
	for _, s := range oldSegs {
		if s.Changed && s.Start != 3 {
			t.Fatalf("expected changed segment starting at byte 3 (after '日'), got %v", oldSegs)
		}
	}
}
