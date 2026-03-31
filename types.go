package drift

import "github.com/tbcrawford/drift/internal/edittype"

// Op represents the operation type for a diff edit or line.
// It is an alias for edittype.Op to break the import cycle between the
// root package and internal algorithm/hunk implementations.
type Op = edittype.Op

const (
	// Equal indicates the line is unchanged.
	Equal = edittype.Equal
	// Insert indicates the line was added.
	Insert = edittype.Insert
	// Delete indicates the line was removed.
	Delete = edittype.Delete
)

// Edit represents a single line-level change from the diff algorithm.
// OldLine and NewLine are 1-indexed line numbers; zero means not applicable.
// It is an alias for edittype.Edit.
type Edit = edittype.Edit

// Line represents a single rendered line in a diff hunk.
// It carries the operation, content, and original line numbers.
// It is an alias for edittype.Line.
type Line = edittype.Line

// Hunk represents a contiguous block of changes with surrounding context lines.
// OldStart/OldLines and NewStart/NewLines are used to generate @@ hunk headers.
// It is an alias for edittype.Hunk.
type Hunk = edittype.Hunk

// DiffResult is the structured output of a diff operation.
// When IsEqual is true, Hunks is empty and no edits were found.
// It is an alias for edittype.DiffResult.
type DiffResult = edittype.DiffResult
