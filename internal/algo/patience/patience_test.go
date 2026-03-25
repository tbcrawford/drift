package patience_test

import (
	"testing"

	"github.com/tylercrawford/drift/internal/algo/patience"
)

// TestBothEmpty — minimal bootstrap test required by task 02-01-01 acceptance criteria.
// Full test suite is in task 02-01-02.
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
