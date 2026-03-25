// Package myers implements the Myers diff algorithm (Eugene Myers, 1986).
// It computes the Shortest Edit Script (SES) between two slices of strings,
// producing a minimal sequence of Equal, Insert, and Delete edits.
package myers

import (
	"github.com/tylercrawford/drift"
	"github.com/tylercrawford/drift/internal/algo"
)

// Ensure Myers satisfies the algo.Differ interface at compile time.
var _ algo.Differ = (*Myers)(nil)

// Myers is a diff algorithm implementation using the Myers SES algorithm.
type Myers struct{}

// New returns a new Myers diff instance.
func New() *Myers {
	return &Myers{}
}

// copyV returns a copy of the V array.
func copyV(v []int) []int {
	cp := make([]int, len(v))
	copy(cp, v)
	return cp
}

// Diff computes the minimum edit sequence between oldLines and newLines.
// Returns a slice of Edits in order from start to end of both inputs.
//
// The Myers algorithm finds the Shortest Edit Script (SES) via a forward
// pass over edit-distance diagonals, recording a trace at the END of each
// d-loop iteration. Backtracking through the trace reconstructs the edits.
//
// CRITICAL: trace is saved at the END of the d-loop (after the k-loop), NOT
// at the top. Saving at the top causes an off-by-one invisible on short inputs
// but wrong on 100+ line files.
func (m *Myers) Diff(oldLines, newLines []string) []drift.Edit {
	N := len(oldLines)
	M := len(newLines)

	// Edge case: both empty
	if N == 0 && M == 0 {
		return []drift.Edit{}
	}

	// Edge case: old empty — all inserts
	if N == 0 {
		edits := make([]drift.Edit, M)
		for i := 0; i < M; i++ {
			edits[i] = drift.Edit{Op: drift.Insert, OldLine: 0, NewLine: i + 1}
		}
		return edits
	}

	// Edge case: new empty — all deletes
	if M == 0 {
		edits := make([]drift.Edit, N)
		for i := 0; i < N; i++ {
			edits[i] = drift.Edit{Op: drift.Delete, OldLine: i + 1, NewLine: 0}
		}
		return edits
	}

	maxD := N + M

	// V array indexed by diagonal k = x - y, offset by maxD so k can be negative.
	// V[k+maxD] = furthest-reaching x on diagonal k.
	// Pre-allocate to 2*maxD+1 to avoid repeated allocations.
	v := make([]int, 2*maxD+1)

	// trace stores a snapshot of V at the END of each d-loop iteration.
	// trace[d] is the V state after all k-diagonals at edit-distance d are processed.
	// CRITICAL: saved at END of d-loop, not at top.
	trace := make([][]int, 0, maxD+1)

	// Forward pass: find the edit distance and record trace.
	// For each edit distance d, extend all reachable diagonals to their
	// furthest x position. The V array is updated in place. A snapshot is
	// saved at the END of each d-iteration so backtracking can recover the
	// exact path. (Saving at the TOP would give the state BEFORE this d's
	// extensions, causing an off-by-one on the backtrack.)
outer:
	for d := 0; d <= maxD; d++ {
		for k := -d; k <= d; k += 2 {
			// Determine whether to move down (insert from new) or right (delete from old).
			var x int
			if k == -d || (k != d && v[k-1+maxD] < v[k+1+maxD]) {
				// Move down: come from diagonal k+1 (insert a line from newLines)
				x = v[k+1+maxD]
			} else {
				// Move right: come from diagonal k-1 (delete a line from oldLines)
				x = v[k-1+maxD] + 1
			}
			y := x - k

			// Follow the snake: advance while lines match (Equal edits)
			for x < N && y < M && oldLines[x] == newLines[y] {
				x++
				y++
			}

			v[k+maxD] = x

			// Reached the endpoint: save the final trace entry and stop.
			if x == N && y == M {
				trace = append(trace, copyV(v)) // END of d-loop — correct position
				break outer
			}
		}

		// Save V snapshot at END of d-loop (after all k-diagonals updated).
		trace = append(trace, copyV(v))
	}

	// Backtrack through trace to reconstruct the edit sequence in reverse.
	// We walk backwards from (x=N, y=M) to (0,0), reading from trace[d].
	// trace[d] tells us the V state AFTER processing edit distance d,
	// which allows us to determine what move was made to reach current (x,y).
	edits := make([]drift.Edit, 0, N+M)
	x, y := N, M

	for d := len(trace) - 1; d >= 0; d-- {
		vd := trace[d]
		k := x - y

		// Determine the diagonal we came from at this step.
		var prevK int
		if k == -d || (k != d && vd[k-1+maxD] < vd[k+1+maxD]) {
			prevK = k + 1 // came from down (insert)
		} else {
			prevK = k - 1 // came from right (delete)
		}

		// Previous position before the edit + snake
		prevX := vd[prevK+maxD]
		prevY := prevX - prevK

		// Emit Equal edits for the snake portion (diagonal moves)
		// The snake goes from (prevX, prevY) to (x, y) via Equal moves.
		for x > prevX && y > prevY {
			x--
			y--
			edits = append(edits, drift.Edit{
				Op:      drift.Equal,
				OldLine: x + 1, // convert to 1-indexed
				NewLine: y + 1,
			})
		}

		// Emit the Insert or Delete that preceded the snake.
		if d > 0 {
			if prevK == k+1 {
				// Came from down: Insert (y decreased by 1, x unchanged)
				// prevX == x, prevY == y-1 → the new line at y (0-indexed) was inserted
				y--
				edits = append(edits, drift.Edit{
					Op:      drift.Insert,
					OldLine: 0,
					NewLine: y + 1, // convert to 1-indexed
				})
			} else {
				// Came from right: Delete (x decreased by 1, y unchanged)
				// prevX == x-1, prevY == y → the old line at x (0-indexed) was deleted
				x--
				edits = append(edits, drift.Edit{
					Op:      drift.Delete,
					OldLine: x + 1, // convert to 1-indexed
					NewLine: 0,
				})
			}
		}
	}

	// Reverse to get forward order (backtracking produces reversed output)
	for i, j := 0, len(edits)-1; i < j; i, j = i+1, j-1 {
		edits[i], edits[j] = edits[j], edits[i]
	}

	return edits
}
