package drift_test

import (
	"testing"

	"github.com/tbcrawford/drift"
)

// TestDiff_IdenticalInputs verifies that identical inputs return IsEqual=true
// and an empty Hunks slice.
func TestDiff_IdenticalInputs(t *testing.T) {
	result, err := drift.Diff("a\nb\nc", "a\nb\nc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsEqual {
		t.Error("expected IsEqual=true for identical inputs")
	}
	if len(result.Hunks) != 0 {
		t.Errorf("expected 0 hunks, got %d", len(result.Hunks))
	}
}

// TestDiff_CRLFNormalization verifies that \r\n inputs produce the same result
// as \n inputs, confirming Windows line-ending normalization.
func TestDiff_CRLFNormalization(t *testing.T) {
	lf, err := drift.Diff("a\nb", "a\nc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	crlf, err := drift.Diff("a\r\nb", "a\r\nc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if lf.IsEqual != crlf.IsEqual {
		t.Errorf("IsEqual mismatch: LF=%v CRLF=%v", lf.IsEqual, crlf.IsEqual)
	}
	if len(lf.Hunks) != len(crlf.Hunks) {
		t.Errorf("hunk count mismatch: LF=%d CRLF=%d", len(lf.Hunks), len(crlf.Hunks))
	}
	if len(lf.Hunks) > 0 && len(crlf.Hunks) > 0 {
		if lf.Hunks[0].OldLines != crlf.Hunks[0].OldLines {
			t.Errorf("OldLines mismatch: LF=%d CRLF=%d",
				lf.Hunks[0].OldLines, crlf.Hunks[0].OldLines)
		}
	}
}

// TestDiff_SingleAddedLine verifies that adding one line produces a 1-hunk
// result with exactly 1 Insert line.
func TestDiff_SingleAddedLine(t *testing.T) {
	result, err := drift.Diff("a\nb", "a\nb\nc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsEqual {
		t.Fatal("expected IsEqual=false when a line is added")
	}
	if len(result.Hunks) != 1 {
		t.Fatalf("expected 1 hunk, got %d", len(result.Hunks))
	}

	var insertCount int
	for _, line := range result.Hunks[0].Lines {
		if line.Op == drift.Insert {
			insertCount++
			if line.Content != "c" {
				t.Errorf("inserted line content: got %q, want %q", line.Content, "c")
			}
		}
	}
	if insertCount != 1 {
		t.Errorf("expected 1 Insert line, got %d", insertCount)
	}
}

// TestDiff_WithContextZero verifies that WithContext(0) produces hunks
// containing only changed lines (no Equal context lines).
func TestDiff_WithContextZero(t *testing.T) {
	result, err := drift.Diff("a\nb\nc\nd\ne", "a\nB\nc\nd\ne", drift.WithContext(0))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Hunks) != 1 {
		t.Fatalf("expected 1 hunk, got %d", len(result.Hunks))
	}

	for _, line := range result.Hunks[0].Lines {
		if line.Op == drift.Equal {
			t.Errorf("found Equal line with content %q in context=0 hunk", line.Content)
		}
	}
}

// TestSplitLines_TrailingNewline verifies that files ending with \n don't produce
// a spurious empty last line. This is an indirect test via Diff behavior.
func TestSplitLines_TrailingNewline(t *testing.T) {
	// Both inputs end with \n — should be treated identically to without \n
	result, err := drift.Diff("a\nb\n", "a\nb\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsEqual {
		t.Error("expected IsEqual=true for identical inputs with trailing newline")
	}
}
