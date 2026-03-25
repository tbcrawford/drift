package algo

import "github.com/tylercrawford/drift/internal/edittype"

// Differ is the interface all diff algorithm implementations must satisfy.
type Differ interface {
	// Diff computes the minimum edit sequence between oldLines and newLines.
	// Returns a slice of Edits in order from start to end of the inputs.
	Diff(oldLines, newLines []string) []edittype.Edit
}
