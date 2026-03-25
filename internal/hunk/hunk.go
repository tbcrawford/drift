// Package hunk transforms a flat sequence of diff edits into grouped Hunk
// structures suitable for unified-diff rendering and API consumers.
package hunk

import (
	"github.com/tylercrawford/drift"
)

// Build converts a slice of Edits (as returned by a diff algorithm) into a
// slice of Hunks, each representing a contiguous block of changes surrounded
// by up to contextLines unchanged lines on each side.
//
// Algorithm:
//  1. Find all non-Equal edit positions (using old-line numbers for deletes/equals,
//     new-line numbers for inserts when old is absent).
//  2. Expand each changed position into a range [pos-contextLines, pos+contextLines].
//  3. Merge overlapping or adjacent expanded ranges.
//  4. For each merged range, walk the edit sequence to build []Line.
//  5. Compute OldStart/OldLines/NewStart/NewLines from the resulting Lines.
func Build(edits []drift.Edit, oldLines, newLines []string, contextLines int) []drift.Hunk {
	if len(edits) == 0 {
		return []drift.Hunk{}
	}

	// Step 1: Collect the edit-sequence index of every non-Equal edit.
	// We'll work with edit indices (positions into the edits slice), not line numbers,
	// because the edit sequence already encodes both old and new positions.
	type editRange struct{ start, end int } // inclusive indices into edits
	var changed []int
	for i, e := range edits {
		if e.Op != drift.Equal {
			changed = append(changed, i)
		}
	}
	if len(changed) == 0 {
		return []drift.Hunk{}
	}

	// Step 2 & 3: Expand each changed-edit index by contextLines (in edit-space)
	// and merge overlapping/adjacent ranges.
	// We use edit-sequence indices rather than line numbers because the edit
	// sequence is the natural unit of iteration.
	ranges := make([]editRange, 0, len(changed))
	for _, idx := range changed {
		lo := idx - contextLines
		if lo < 0 {
			lo = 0
		}
		hi := idx + contextLines
		if hi >= len(edits) {
			hi = len(edits) - 1
		}
		if len(ranges) == 0 || lo > ranges[len(ranges)-1].end+1 {
			ranges = append(ranges, editRange{lo, hi})
		} else {
			if hi > ranges[len(ranges)-1].end {
				ranges[len(ranges)-1].end = hi
			}
		}
	}

	// Step 4 & 5: Build a Hunk for each merged range.
	result := make([]drift.Hunk, 0, len(ranges))
	for _, r := range ranges {
		lines := buildLines(edits[r.start:r.end+1], oldLines, newLines)
		if len(lines) == 0 {
			continue
		}
		result = append(result, buildHunkHeader(lines))
	}
	return result
}

// buildLines constructs the []Line slice for a contiguous sub-sequence of edits.
func buildLines(edits []drift.Edit, oldLines, newLines []string) []drift.Line {
	lines := make([]drift.Line, 0, len(edits))
	for _, e := range edits {
		switch e.Op {
		case drift.Equal:
			lines = append(lines, drift.Line{
				Op:      drift.Equal,
				Content: oldLines[e.OldLine-1],
				OldNum:  e.OldLine,
				NewNum:  e.NewLine,
			})
		case drift.Delete:
			lines = append(lines, drift.Line{
				Op:      drift.Delete,
				Content: oldLines[e.OldLine-1],
				OldNum:  e.OldLine,
				NewNum:  0,
			})
		case drift.Insert:
			lines = append(lines, drift.Line{
				Op:      drift.Insert,
				Content: newLines[e.NewLine-1],
				OldNum:  0,
				NewNum:  e.NewLine,
			})
		}
	}
	return lines
}

// buildHunkHeader computes the unified-diff @@ header fields from a []Line slice.
//
// OldStart: first OldNum found by skipping leading Insert lines.
// NewStart: first NewNum found by skipping leading Delete lines.
// OldLines: count of Equal + Delete lines.
// NewLines: count of Equal + Insert lines.
func buildHunkHeader(lines []drift.Line) drift.Hunk {
	oldStart, newStart := 0, 0
	oldCount, newCount := 0, 0

	for _, l := range lines {
		switch l.Op {
		case drift.Equal:
			if oldStart == 0 {
				oldStart = l.OldNum
			}
			if newStart == 0 {
				newStart = l.NewNum
			}
			oldCount++
			newCount++
		case drift.Delete:
			if oldStart == 0 {
				oldStart = l.OldNum
			}
			oldCount++
		case drift.Insert:
			if newStart == 0 {
				newStart = l.NewNum
			}
			newCount++
		}
	}

	// Ensure starts are at least 1 (edge case: all-insert or all-delete hunks).
	if oldStart == 0 {
		oldStart = 1
	}
	if newStart == 0 {
		newStart = 1
	}

	return drift.Hunk{
		OldStart: oldStart,
		OldLines: oldCount,
		NewStart: newStart,
		NewLines: newCount,
		Lines:    lines,
	}
}
