// Package histogram implements Git's Histogram diff algorithm.
// It finds the longest contiguous matching block with the lowest occurrence
// count in old, then recurses on before/after regions. Falls back to Myers
// when all old lines in the region exceed histogramMaxOccurrences = 64.
package histogram

import (
	"sort"

	"github.com/tylercrawford/drift/internal/algo"
	"github.com/tylercrawford/drift/internal/algo/myers"
	"github.com/tylercrawford/drift/internal/edittype"
)

const histogramMaxOccurrences = 64

// Compile-time interface check.
var _ algo.Differ = (*Histogram)(nil)

// Histogram is a diff algorithm implementation using the Histogram algorithm.
type Histogram struct{}

// New returns a new Histogram diff instance.
func New() *Histogram { return &Histogram{} }

// frame is a region of old[os:oe] vs new[ns:ne] to be processed.
type frame struct {
	os, oe int // old start/end (0-indexed, exclusive end)
	ns, ne int // new start/end (0-indexed, exclusive end)
}

// matchResult is the best contiguous matching block found within a frame.
type matchResult struct {
	os, oe int // matched old range [os, oe) — 0-indexed into full old slice
	ns, ne int // matched new range [ns, ne) — 0-indexed into full new slice
}

// Diff computes the edit sequence between old and new using the Histogram
// algorithm. Returns a slice of Edits in file order.
func (h *Histogram) Diff(old, new []string) []edittype.Edit {
	N := len(old)
	M := len(new)

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

	stack := []frame{{0, N, 0, M}}
	edits := make([]edittype.Edit, 0, N+M)

	for len(stack) > 0 {
		f := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if f.os == f.oe && f.ns == f.ne {
			continue
		}

		if f.os == f.oe {
			for i := 0; i < f.ne-f.ns; i++ {
				edits = append(edits, edittype.Edit{Op: edittype.Insert, OldLine: 0, NewLine: f.ns + i + 1})
			}
			continue
		}

		if f.ns == f.ne {
			for i := 0; i < f.oe-f.os; i++ {
				edits = append(edits, edittype.Edit{Op: edittype.Delete, OldLine: f.os + i + 1, NewLine: 0})
			}
			continue
		}

		counts := make(map[string]int, f.oe-f.os)
		for i := f.os; i < f.oe; i++ {
			counts[old[i]]++
		}

		match, found := findBestMatch(old, new, f, counts)

		if !found {
			fallback := myers.New().Diff(old[f.os:f.oe], new[f.ns:f.ne])
			applyOffset(fallback, f.os, f.ns)
			edits = append(edits, fallback...)
			continue
		}

		// Push after-match region first (processed last — file order maintained via LIFO).
		if match.oe < f.oe || match.ne < f.ne {
			stack = append(stack, frame{match.oe, f.oe, match.ne, f.ne})
		}
		// Push before-match region second (processed first).
		if f.os < match.os || f.ns < match.ns {
			stack = append(stack, frame{f.os, match.os, f.ns, match.ns})
		}

		// Emit Equal edits for the matched block.
		for i := 0; i < match.oe-match.os; i++ {
			edits = append(edits, edittype.Edit{
				Op:      edittype.Equal,
				OldLine: match.os + i + 1,
				NewLine: match.ns + i + 1,
			})
		}
	}

	// Sort edits into file order by new-line position (Insert uses NewLine;
	// Delete/Equal use OldLine). Use new-file position as primary key because
	// it is monotonically defined for all ops; for Deletes (NewLine==0) sort
	// by the OldLine-derived new-file cursor position tracked via newCursor.
	sortEdits(edits)
	return edits
}

// findBestMatch finds the longest contiguous matching block between old[f.os:f.oe]
// and new[f.ns:f.ne] where the old lines appear with the fewest occurrences,
// following the jgit HistogramDiff approach.
func findBestMatch(old, new []string, f frame, counts map[string]int) (matchResult, bool) {
	lowcount := histogramMaxOccurrences + 1
	var best matchResult
	found := false

	for ni := f.ns; ni < f.ne; ni++ {
		line := new[ni]
		cnt, ok := counts[line]
		if !ok || cnt > lowcount {
			continue
		}

		for oi := f.os; oi < f.oe; oi++ {
			if old[oi] != line {
				continue
			}
			if counts[line] > lowcount {
				continue
			}

			matchOs, matchNs := oi, ni
			matchOe, matchNe := oi+1, ni+1

			for matchOs > f.os && matchNs > f.ns && old[matchOs-1] == new[matchNs-1] {
				matchOs--
				matchNs--
			}
			for matchOe < f.oe && matchNe < f.ne && old[matchOe] == new[matchNe] {
				matchOe++
				matchNe++
			}

			regionLow := histogramMaxOccurrences + 1
			for k := matchOs; k < matchOe; k++ {
				if c := counts[old[k]]; c < regionLow {
					regionLow = c
				}
			}

			matchLen := matchOe - matchOs
			bestLen := best.oe - best.os
			if !found || matchLen > bestLen || (matchLen == bestLen && regionLow < lowcount) {
				best = matchResult{matchOs, matchOe, matchNs, matchNe}
				lowcount = regionLow
				found = true
			}
		}
	}

	return best, found
}

// applyOffset adjusts OldLine/NewLine in a Myers fallback edit slice by the
// subslice offsets used when calling Myers on a sub-region.
func applyOffset(edits []edittype.Edit, oldOff, newOff int) {
	for i := range edits {
		if edits[i].OldLine > 0 {
			edits[i].OldLine += oldOff
		}
		if edits[i].NewLine > 0 {
			edits[i].NewLine += newOff
		}
	}
}

// sortEdits reorders edits into canonical unified-diff file order.
// Sort key: for Insert use NewLine; for Equal/Delete use OldLine.
// When keys are equal, Equal/Delete come before Insert.
func sortEdits(edits []edittype.Edit) {
	sort.SliceStable(edits, func(i, j int) bool {
		ei, ej := edits[i], edits[j]
		ki := sortKey(ei)
		kj := sortKey(ej)
		if ki != kj {
			return ki < kj
		}
		return editTieBreak(ei) < editTieBreak(ej)
	})
}

// sortKey returns the new-file line position for an edit.
// Insert uses NewLine directly. Equal/Delete use NewLine if available,
// otherwise fall back to OldLine (Delete has NewLine=0).
// For Equal, NewLine == the new-file position which is what we want.
// For Delete (NewLine=0), we use OldLine as a proxy — but Deletes sort
// after Equals/Inserts at the same new-file position by tiebreak.
func sortKey(e edittype.Edit) int {
	if e.NewLine > 0 {
		return e.NewLine
	}
	// Delete: NewLine == 0. Use OldLine as position proxy.
	return e.OldLine
}

func editTieBreak(e edittype.Edit) int {
	if e.Op == edittype.Insert {
		return 1
	}
	return 0
}
