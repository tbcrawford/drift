// Package edittype defines the core types shared between the root drift package
// and internal algorithm/hunk implementations. This package exists solely to
// break the import cycle that would arise if internal packages (algo/myers,
// hunk) imported the root drift package, which in turn imports them.
//
// Consumers should use the re-exported types in the public library package:
//
//	import "github.com/tylercrawford/drift"
//	drift.Op, drift.Edit, drift.Line, drift.Span, drift.Hunk, drift.DiffResult
package edittype

// Op represents the operation type for a diff edit or line.
type Op int

const (
	// Equal indicates the line is unchanged.
	Equal Op = iota
	// Insert indicates the line was added.
	Insert
	// Delete indicates the line was removed.
	Delete
)

// Edit represents a single line-level change from the diff algorithm.
// OldLine and NewLine are 1-indexed line numbers; zero means not applicable.
type Edit struct {
	Op      Op
	OldLine int // 1-indexed; 0 for Insert (no old line)
	NewLine int // 1-indexed; 0 for Delete (no new line)
}

// Span marks a character range within a Line for intra-line highlighting.
// Start and End are byte offsets into Line.Content. Reserved for v1.x.
type Span struct {
	Start int
	End   int
	Op    Op
}

// Line represents a single rendered line in a diff hunk.
// It carries the operation, content, and original line numbers.
type Line struct {
	Op      Op
	Content string // line text without trailing newline
	OldNum  int    // 1-indexed old file line number; 0 if inserted
	NewNum  int    // 1-indexed new file line number; 0 if deleted
	// Spans holds intra-line word-level diff spans (nil in v1.0; reserved for v1.x).
	Spans []Span
}

// Hunk represents a contiguous block of changes with surrounding context lines.
// OldStart/OldLines and NewStart/NewLines are used to generate @@ hunk headers.
type Hunk struct {
	OldStart int    // 1-indexed start line in old file
	OldLines int    // number of lines from old file in this hunk
	NewStart int    // 1-indexed start line in new file
	NewLines int    // number of lines from new file in this hunk
	Lines    []Line // all lines in the hunk (context + changes)
}

// DiffResult is the structured output of a diff operation.
// When IsEqual is true, Hunks is empty and no edits were found.
type DiffResult struct {
	Hunks   []Hunk
	IsEqual bool // true when both inputs were identical
}
