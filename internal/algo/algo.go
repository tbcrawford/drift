package algo

import "github.com/tylercrawford/drift/internal/edittype"

// Differ is the interface all diff algorithm implementations must satisfy.
type Differ interface {
	// Diff computes the minimum edit sequence between oldLines and newLines.
	// Returns a slice of Edits in order from start to end of the inputs.
	Diff(oldLines, newLines []string) []edittype.Edit
}

// ApplyOffset adjusts OldLine and NewLine in each edit by the given offsets.
// Myers returns 1-indexed lines relative to the sub-slice; this converts them
// to 1-indexed lines relative to the full file. Used by Patience and Histogram
// when falling back to Myers on a sub-region.
func ApplyOffset(edits []edittype.Edit, oldOff, newOff int) {
	for i := range edits {
		if edits[i].OldLine > 0 {
			edits[i].OldLine += oldOff
		}
		if edits[i].NewLine > 0 {
			edits[i].NewLine += newOff
		}
	}
}
