package algo

import (
	"testing"

	"github.com/tbcrawford/drift/internal/edittype"
)

func TestApplyOffset_adjustsPositiveLines(t *testing.T) {
	edits := []edittype.Edit{
		{Op: edittype.Delete, OldLine: 1, NewLine: 0},
		{Op: edittype.Insert, OldLine: 0, NewLine: 1},
		{Op: edittype.Equal, OldLine: 2, NewLine: 2},
	}
	ApplyOffset(edits, 10, 20)

	if edits[0].OldLine != 11 {
		t.Errorf("Delete OldLine: got %d want 11", edits[0].OldLine)
	}
	if edits[0].NewLine != 0 {
		t.Errorf("Delete NewLine: should stay 0 (was 0 before offset), got %d", edits[0].NewLine)
	}
	if edits[1].OldLine != 0 {
		t.Errorf("Insert OldLine: should stay 0, got %d", edits[1].OldLine)
	}
	if edits[1].NewLine != 21 {
		t.Errorf("Insert NewLine: got %d want 21", edits[1].NewLine)
	}
	if edits[2].OldLine != 12 {
		t.Errorf("Equal OldLine: got %d want 12", edits[2].OldLine)
	}
	if edits[2].NewLine != 22 {
		t.Errorf("Equal NewLine: got %d want 22", edits[2].NewLine)
	}
}

func TestApplyOffset_zeroLinesNotAdjusted(t *testing.T) {
	// Zero means "no line number assigned" — must not be offset.
	edits := []edittype.Edit{
		{Op: edittype.Delete, OldLine: 0, NewLine: 0},
	}
	ApplyOffset(edits, 5, 5)
	if edits[0].OldLine != 0 {
		t.Errorf("zero OldLine should remain 0, got %d", edits[0].OldLine)
	}
	if edits[0].NewLine != 0 {
		t.Errorf("zero NewLine should remain 0, got %d", edits[0].NewLine)
	}
}

func TestApplyOffset_zeroOffsets(t *testing.T) {
	// Zero offsets must not change any line numbers.
	edits := []edittype.Edit{
		{Op: edittype.Equal, OldLine: 3, NewLine: 7},
	}
	ApplyOffset(edits, 0, 0)
	if edits[0].OldLine != 3 {
		t.Errorf("OldLine: got %d want 3", edits[0].OldLine)
	}
	if edits[0].NewLine != 7 {
		t.Errorf("NewLine: got %d want 7", edits[0].NewLine)
	}
}

func TestApplyOffset_emptySlice(t *testing.T) {
	// Must not panic on empty input.
	ApplyOffset(nil, 10, 20)
	ApplyOffset([]edittype.Edit{}, 10, 20)
}

func TestApplyOffset_largeOffset(t *testing.T) {
	edits := []edittype.Edit{
		{Op: edittype.Equal, OldLine: 1, NewLine: 1},
	}
	ApplyOffset(edits, 9999, 8888)
	if edits[0].OldLine != 10000 {
		t.Errorf("OldLine: got %d want 10000", edits[0].OldLine)
	}
	if edits[0].NewLine != 8889 {
		t.Errorf("NewLine: got %d want 8889", edits[0].NewLine)
	}
}
