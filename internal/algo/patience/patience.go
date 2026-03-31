// Package patience implements the Patience diff algorithm (Bram Cohen, 2007).
// It anchors diffs on unique lines — lines appearing exactly once in each file —
// computes their LCS to form anchors, and recurses on inter-anchor gaps.
// When no unique lines exist in a sub-range, it falls back to Myers.
package patience

import (
	"github.com/tbcrawford/drift/internal/algo"
	"github.com/tbcrawford/drift/internal/algo/myers"
	"github.com/tbcrawford/drift/internal/edittype"
)

// Ensure Patience satisfies the algo.Differ interface at compile time.
var _ algo.Differ = (*Patience)(nil)

// Patience is a diff algorithm implementation using the Patience diff algorithm.
type Patience struct{}

// New returns a new Patience diff instance.
func New() *Patience { return &Patience{} }

// stackItem is a tagged union pushed onto the processing stack.
// Either it holds a frame to process (isEmit=false) or a pre-built edit to emit (isEmit=true).
type stackItem struct {
	isEmit bool
	edit   edittype.Edit
	// frame fields (used when isEmit=false)
	os, oe int // old start/end (0-indexed, exclusive end)
	ns, ne int // new start/end (0-indexed, exclusive end)
}

// Diff computes the minimum edit sequence between oldLines and newLines using
// the Patience algorithm. Returns edits in forward file order.
func (p *Patience) Diff(old, new []string) []edittype.Edit {
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

	stack := []stackItem{{isEmit: false, os: 0, oe: N, ns: 0, ne: M}}
	edits := make([]edittype.Edit, 0, N+M)
	m := myers.New()

	for len(stack) > 0 {
		item := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		// Pre-built emit item
		if item.isEmit {
			edits = append(edits, item.edit)
			continue
		}

		f := struct{ os, oe, ns, ne int }{item.os, item.oe, item.ns, item.ne}

		// Head trim: consume matching prefix lines (emit directly — order is safe here)
		for f.os < f.oe && f.ns < f.ne && old[f.os] == new[f.ns] {
			edits = append(edits, edittype.Edit{
				Op:      edittype.Equal,
				OldLine: f.os + 1,
				NewLine: f.ns + 1,
			})
			f.os++
			f.ns++
		}

		// Tail trim: collect matching suffix edits to emit last
		te := 0
		for te < (f.oe-f.os) && te < (f.ne-f.ns) && old[f.oe-1-te] == new[f.ne-1-te] {
			te++
		}
		// Build tail edits in forward order
		tailEdits := make([]edittype.Edit, te)
		for i := 0; i < te; i++ {
			tailEdits[i] = edittype.Edit{
				Op:      edittype.Equal,
				OldLine: f.oe - te + i + 1,
				NewLine: f.ne - te + i + 1,
			}
		}
		f.oe -= te
		f.ne -= te

		// Empty after trim
		if f.os == f.oe && f.ns == f.ne {
			edits = append(edits, tailEdits...)
			continue
		}

		// Old empty after trim: all inserts
		if f.os == f.oe {
			for i := 0; i < f.ne-f.ns; i++ {
				edits = append(edits, edittype.Edit{
					Op:      edittype.Insert,
					OldLine: 0,
					NewLine: f.ns + i + 1,
				})
			}
			edits = append(edits, tailEdits...)
			continue
		}

		// New empty after trim: all deletes
		if f.ns == f.ne {
			for i := 0; i < f.oe-f.os; i++ {
				edits = append(edits, edittype.Edit{
					Op:      edittype.Delete,
					OldLine: f.os + i + 1,
					NewLine: 0,
				})
			}
			edits = append(edits, tailEdits...)
			continue
		}

		// Find unique elements: lines appearing exactly once in each range
		oldFreq := buildFreqMap(old, f.os, f.oe)
		newFreq := buildFreqMap(new, f.ns, f.ne)

		var oldUniq []lineIdx
		for i := f.os; i < f.oe; i++ {
			if oldFreq[old[i]] == 1 && newFreq[old[i]] == 1 {
				oldUniq = append(oldUniq, lineIdx{old[i], i})
			}
		}
		var newUniq []lineIdx
		for i := f.ns; i < f.ne; i++ {
			if newFreq[new[i]] == 1 && oldFreq[new[i]] == 1 {
				newUniq = append(newUniq, lineIdx{new[i], i})
			}
		}

		// Compute LCS of unique anchors
		anchors := lcsAnchors(oldUniq, newUniq)

		// No anchors: Myers fallback
		if len(anchors) == 0 {
			sub := m.Diff(old[f.os:f.oe], new[f.ns:f.ne])
			algo.ApplyOffset(sub, f.os, f.ns)
			edits = append(edits, sub...)
			edits = append(edits, tailEdits...)
			continue
		}

		// With anchors: push items onto stack in reverse order so they process left-to-right.
		// Order to emit: gap[0], anchor[0], gap[1], anchor[1], ..., anchor[n-1], gap[n], tail
		//
		// Since we pop from the end, push in reverse:
		// 1. Push tail emits (last to push = last to process, but we want them last → push first)
		// 2. Push gap[n] frame
		// 3. Push anchor[n-1] emit
		// 4. Push gap[n-1] frame
		// ...
		// n+2. Push gap[0] frame (last to push = first to pop)

		// Push tail in reverse (they'll be emitted last)
		for i := len(tailEdits) - 1; i >= 0; i-- {
			stack = append(stack, stackItem{isEmit: true, edit: tailEdits[i]})
		}

		// Push gaps and anchors in reverse order
		prevOld := f.os
		prevNew := f.ns

		type gapInfo struct {
			gos, goe, gns, gne int
		}
		gaps := make([]gapInfo, 0, len(anchors)+1)
		for _, a := range anchors {
			gaps = append(gaps, gapInfo{prevOld, a.oldIdx, prevNew, a.newIdx})
			prevOld = a.oldIdx + 1
			prevNew = a.newIdx + 1
		}
		gaps = append(gaps, gapInfo{prevOld, f.oe, prevNew, f.ne})

		// Push in reverse: gap[n], anchor[n-1], gap[n-1], ..., anchor[0], gap[0]
		for i := len(anchors); i >= 0; i-- {
			g := gaps[i]
			if g.gos < g.goe || g.gns < g.gne {
				stack = append(stack, stackItem{
					isEmit: false,
					os:     g.gos, oe: g.goe,
					ns: g.gns, ne: g.gne,
				})
			}
			if i > 0 {
				a := anchors[i-1]
				stack = append(stack, stackItem{
					isEmit: true,
					edit: edittype.Edit{
						Op:      edittype.Equal,
						OldLine: a.oldIdx + 1,
						NewLine: a.newIdx + 1,
					},
				})
			}
		}
	}

	return edits
}

// buildFreqMap returns a frequency map for lines[start:end].
func buildFreqMap(lines []string, start, end int) map[string]int {
	freq := make(map[string]int, end-start)
	for i := start; i < end; i++ {
		freq[lines[i]]++
	}
	return freq
}

// lineIdx pairs a line string with its original 0-indexed position in the file.
type lineIdx struct {
	line string
	idx  int
}

// lcsAnchors computes the LCS of two unique-line slices using O(N·M) DP.
// Returns matched anchor pairs with original file 0-indexed positions.
func lcsAnchors(oldUniq, newUniq []lineIdx) []anchor {
	N := len(oldUniq)
	M := len(newUniq)
	if N == 0 || M == 0 {
		return nil
	}

	// O(N·M) LCS DP table
	dp := make([][]int, N+1)
	for i := range dp {
		dp[i] = make([]int, M+1)
	}

	for i := N - 1; i >= 0; i-- {
		for j := M - 1; j >= 0; j-- {
			if oldUniq[i].line == newUniq[j].line {
				dp[i][j] = dp[i+1][j+1] + 1
			} else if dp[i+1][j] > dp[i][j+1] {
				dp[i][j] = dp[i+1][j]
			} else {
				dp[i][j] = dp[i][j+1]
			}
		}
	}

	// Backtrack to find matched pairs
	result := make([]anchor, 0, dp[0][0])
	i, j := 0, 0
	for i < N && j < M {
		if oldUniq[i].line == newUniq[j].line {
			result = append(result, anchor{oldUniq[i].idx, newUniq[j].idx})
			i++
			j++
		} else if dp[i+1][j] > dp[i][j+1] {
			i++
		} else {
			j++
		}
	}

	return result
}

// anchor holds matched old/new 0-indexed positions from the LCS of unique lines.
type anchor struct {
	oldIdx, newIdx int
}
