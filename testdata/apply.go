// Package testdata provides test helpers for drift property-based tests.
package testdata

import "github.com/tylercrawford/drift"

// Apply reconstructs the "new" file from a DiffResult and the original "old" lines.
// Used to verify the round-trip invariant: Apply(Diff(old, new), oldLines) == newLines.
//
// Algorithm:
//  1. Walk hunks in order of OldStart (they are already sorted).
//  2. Before each hunk, emit old lines that are not covered by the hunk.
//  3. Within a hunk, emit Equal and Insert lines; skip Delete lines.
//  4. After all hunks, emit remaining old lines.
func Apply(result drift.DiffResult, oldLines []string) []string {
	if result.IsEqual {
		// Identical files: new == old, return a copy of oldLines.
		out := make([]string, len(oldLines))
		copy(out, oldLines)
		return out
	}

	var out []string
	// oldIdx is a 0-indexed cursor tracking which old line we've consumed.
	oldIdx := 0

	for _, h := range result.Hunks {
		// h.OldStart is 1-indexed; convert to 0-indexed.
		hunkOldStart := h.OldStart - 1

		// Emit old lines before this hunk that are not covered by any hunk.
		// These are context lines that fell outside the hunk window.
		for oldIdx < hunkOldStart {
			out = append(out, oldLines[oldIdx])
			oldIdx++
		}

		// Emit lines from within the hunk.
		for _, l := range h.Lines {
			switch l.Op {
			case drift.Equal:
				// Line exists in both files — emit it and advance old cursor.
				out = append(out, l.Content)
				oldIdx++
			case drift.Insert:
				// New line added — emit it, old cursor unchanged.
				out = append(out, l.Content)
			case drift.Delete:
				// Old line removed — skip it, advance old cursor.
				oldIdx++
			}
		}
	}

	// Emit remaining old lines after the last hunk.
	for oldIdx < len(oldLines) {
		out = append(out, oldLines[oldIdx])
		oldIdx++
	}

	// Preserve empty-slice semantics: if result is non-nil diff but
	// new file is empty, return empty slice rather than nil.
	if out == nil {
		return []string{}
	}
	return out
}
