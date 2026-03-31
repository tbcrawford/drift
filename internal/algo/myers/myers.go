// Package myers implements the Myers diff algorithm (Eugene Myers, 1986).
// It computes the Shortest Edit Script (SES) between two slices of strings,
// producing a minimal sequence of Equal, Insert, and Delete edits.
//
// This implementation uses the Hirschberg linear-space divide-and-conquer
// variant (Myers 1986 §4): the forward and reverse passes meet in the middle
// to find a midpoint, then recurse on each half. Peak memory is O(N+M)
// rather than the O((N+M)²) of the trace-snapshot approach.
package myers

import (
	"github.com/tbcrawford/drift/internal/algo"
	"github.com/tbcrawford/drift/internal/edittype"
)

// Ensure Myers satisfies the algo.Differ interface at compile time.
var _ algo.Differ = (*Myers)(nil)

// Myers is a diff algorithm implementation using the Myers SES algorithm.
type Myers struct{}

// New returns a new Myers diff instance.
func New() *Myers {
	return &Myers{}
}

// Diff computes the minimum edit sequence between oldLines and newLines.
// Returns a slice of Edits in order from start to end of both inputs.
//
// The implementation uses Hirschberg's linear-space divide-and-conquer
// approach: midpoint() finds the center of the optimal edit path using only
// two O(N+M) V arrays, then ses() recurses on the two halves. Total space
// is O(N+M); time complexity remains O(ND).
func (m *Myers) Diff(oldLines, newLines []string) []edittype.Edit {
	N := len(oldLines)
	M := len(newLines)

	if N == 0 && M == 0 {
		return []edittype.Edit{}
	}
	if N == 0 {
		edits := make([]edittype.Edit, M)
		for i := 0; i < M; i++ {
			edits[i] = edittype.Edit{Op: edittype.Insert, OldLine: 0, NewLine: i + 1}
		}
		return edits
	}
	if M == 0 {
		edits := make([]edittype.Edit, N)
		for i := 0; i < N; i++ {
			edits[i] = edittype.Edit{Op: edittype.Delete, OldLine: i + 1, NewLine: 0}
		}
		return edits
	}

	edits := make([]edittype.Edit, 0, N+M)
	ses(oldLines, newLines, 0, 0, &edits)
	return edits
}

// ses recursively builds the Shortest Edit Script for old[0:] and new[0:]
// using the Hirschberg divide-and-conquer approach. oldOff / newOff are the
// offsets from the start of the original full inputs (used for 1-based line
// numbers in emitted edits).
func ses(old, new []string, oldOff, newOff int, edits *[]edittype.Edit) {
	N, M := len(old), len(new)

	// Trim identical prefix — these are Equal edits we can emit immediately.
	prefix := 0
	for prefix < N && prefix < M && old[prefix] == new[prefix] {
		*edits = append(*edits, edittype.Edit{
			Op:      edittype.Equal,
			OldLine: oldOff + prefix + 1,
			NewLine: newOff + prefix + 1,
		})
		prefix++
	}

	// Trim identical suffix — collect for emission after inner edits.
	suffix := 0
	for suffix < N-prefix && suffix < M-prefix &&
		old[N-1-suffix] == new[M-1-suffix] {
		suffix++
	}

	// Work on the middle section (after prefix, before suffix).
	lo := old[prefix : N-suffix]
	ln := new[prefix : M-suffix]
	loOff := oldOff + prefix
	lnOff := newOff + prefix
	nLo, nLn := len(lo), len(ln)

	if nLo == 0 && nLn == 0 {
		// Only equal lines remain (the suffix).
	} else if nLo == 0 {
		for i := 0; i < nLn; i++ {
			*edits = append(*edits, edittype.Edit{
				Op:      edittype.Insert,
				OldLine: 0,
				NewLine: lnOff + i + 1,
			})
		}
	} else if nLn == 0 {
		for i := 0; i < nLo; i++ {
			*edits = append(*edits, edittype.Edit{
				Op:      edittype.Delete,
				OldLine: loOff + i + 1,
				NewLine: 0,
			})
		}
	} else if nLo == 1 && nLn == 1 {
		// Only possibility after prefix/suffix trim: they differ.
		*edits = append(*edits,
			edittype.Edit{Op: edittype.Delete, OldLine: loOff + 1, NewLine: 0},
			edittype.Edit{Op: edittype.Insert, OldLine: 0, NewLine: lnOff + 1},
		)
	} else {
		// Divide: find the split point and recurse on each half.
		// midpoint returns (mx, my): the position in the edit grid after the
		// non-diagonal step of the middle snake. Both halves are solved
		// recursively; Equal lines are emitted via prefix/suffix stripping.
		mx, my := midpoint(lo, ln)
		ses(lo[:mx], ln[:my], loOff, lnOff, edits)
		ses(lo[mx:], ln[my:], loOff+mx, lnOff+my, edits)
	}

	// Emit the identical suffix lines.
	for i := 0; i < suffix; i++ {
		*edits = append(*edits, edittype.Edit{
			Op:      edittype.Equal,
			OldLine: oldOff + (N - suffix) + i + 1,
			NewLine: newOff + (M - suffix) + i + 1,
		})
	}
}

// midpoint finds the split point (mx, my) of the optimal edit path between
// old[0:N] and new[0:M] using the Hirschberg linear-space technique
// (Myers 1986 §4b).
//
// It returns a position (mx, my) on the optimal path such that:
//   - old[:mx] and new[:my] form the left sub-problem
//   - old[mx:] and new[my:] form the right sub-problem
//
// The position corresponds to the end of the non-diagonal (delete/insert)
// step of the middle snake; any diagonal (equal) lines following it are
// handled by prefix stripping in the recursive ses call on the right half.
//
// Arrays:
//   - vf[k+offset]: furthest x reached on forward diagonal k, scanning from (0,0)
//   - vb[c+offset]: furthest y reached on backward diagonal c, scanning from (N,M)
//     going in reverse (y decreases). "Furthest" for backward means lowest y.
//
// Overlap conditions (Myers §4b):
//   - delta odd  → the optimal path length is odd; check during forward pass:
//     when forward diagonal k overlaps backward diagonal c = k-delta,
//     the overlap is y >= vb[c] (forward y has passed the backward y).
//   - delta even → the optimal path length is even; check during backward pass:
//     when backward diagonal c overlaps forward diagonal k = c+delta,
//     the overlap is x <= vf[k] (forward x has passed the backward x).
//
// Preconditions (guaranteed by ses):
//   - N > 0 && M > 0
//   - Not all-equal (prefix/suffix stripping has already occurred)
func midpoint(old, new []string) (mx, my int) {
	N, M := len(old), len(new)
	// maxD: upper bound on edit distance. Used as array offset so negative
	// diagonal indices are valid.
	maxD := N + M
	half := (maxD + 1) / 2
	offset := maxD

	// vf[k+offset] = furthest x on forward diagonal k (from top-left).
	// Seed vf[1] = 0: at d=0 the algorithm selects x = vf[k+1] = 0 for k=0.
	vf := make([]int, 2*maxD+1)
	vf[1+offset] = 0

	// vb[c+offset] = lowest y on backward diagonal c (from bottom-right).
	// Seed vb[1] = M: at d=0 the backward algorithm selects y = vb[c+1] = M for c=0.
	vb := make([]int, 2*maxD+1)
	vb[1+offset] = M

	delta := N - M

	for d := 0; d <= half; d++ {
		// ── Forward pass ──────────────────────────────────────────────────
		// Iterate k in descending order (prefer upper diagonals = deletions
		// first, mirroring Git's convention).
		for k := d; k >= -d; k -= 2 {
			var x int
			if k == -d || (k != d && vf[k-1+offset] < vf[k+1+offset]) {
				// Downward step (insert new[y]): start at same x as k+1.
				x = vf[k+1+offset]
			} else {
				// Rightward step (delete old[x]): advance x from k-1.
				x = vf[k-1+offset] + 1
			}
			y := x - k
			// Extend the snake (equal lines).
			for x < N && y < M && old[x] == new[y] {
				x++
				y++
			}
			vf[k+offset] = x

			// When delta is odd the optimal D is odd. The forward pass at
			// step d can meet the backward pass which ran d-1 steps.
			// Backward diagonals in range [-(d-1), d-1] have been computed.
			if delta%2 != 0 {
				c := k - delta
				if c >= -(d-1) && c <= d-1 {
					// Overlap when forward y >= backward y on same diagonal.
					if y >= vb[c+offset] {
						// mx, my = position just after the non-diagonal step
						// (the start of the diagonal run we just extended).
						// x - y == k, so the diagonal started at x_start where
						// x_start - y_start == k and x_start = x - diag_len.
						// We return x, y (end of snake) — the right sub-problem
						// ses(old[x:], new[y:]) will handle the rest.
						return x, y
					}
				}
			}
		}

		// ── Backward pass ─────────────────────────────────────────────────
		// vb tracks y-positions (minimise y = prefer insertions trailing).
		for c := d; c >= -d; c -= 2 {
			k := c + delta // corresponding forward diagonal

			var y int
			if c == -d || (c != d && vb[c-1+offset] > vb[c+1+offset]) {
				// Leftward step (insert in reverse, i.e., accept new[y]):
				// same y as diagonal c+1.
				y = vb[c+1+offset]
			} else {
				// Upward step (delete in reverse, i.e., accept old[x]):
				// decrease y from diagonal c-1.
				y = vb[c-1+offset] - 1
			}
			x := y + k
			// Extend the reverse snake (equal lines going backward).
			for x > 0 && y > 0 && old[x-1] == new[y-1] {
				x--
				y--
			}
			vb[c+offset] = y

			// When delta is even the optimal D is even. The backward pass at
			// step d can meet the forward pass which also ran d steps.
			if delta%2 == 0 {
				if k >= -d && k <= d {
					// Overlap when backward x <= forward x on same diagonal.
					if x <= vf[k+offset] {
						return x, y
					}
				}
			}
		}
	}

	// Fallback — should be unreachable for valid non-empty inputs after
	// prefix/suffix trimming. Return a balanced split to avoid infinite recursion.
	return N / 2, M / 2
}
