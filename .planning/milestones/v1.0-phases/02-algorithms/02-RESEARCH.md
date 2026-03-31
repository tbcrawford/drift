# Phase 2: Algorithms — Research

**Phase Goal:** Callers can select Patience or Histogram algorithms, with correct Myers fallback for edge cases
**Requirements:** ALGO-01, ALGO-02, ALGO-03, ALGO-04
**Research Date:** 2026-03-25

---

## 1. Algorithm Deep-Dives

### 1.1 Patience Diff

**Core idea:** Patience diff (Bram Cohen, 2007) only anchors on *unique* lines — lines that appear exactly once in each file. By finding the LCS of those unique anchors, it pins the diff to structurally meaningful points (function signatures, unique declarations) rather than syntactically cheap lines (braces, blank lines). Inter-anchor gaps fall back to Myers.

**Step-by-step algorithm:**

1. **Head/tail trim:** Consume any prefix of equal lines and any suffix of equal lines. These are trivially equal and need no further work.
2. **Find unique elements:** Build a frequency map for each side. Collect lines that appear exactly once in `old` and exactly once in `new`. These become the anchor candidates.
3. **LCS of unique anchors:** Run an O(N·M) LCS over the unique-element slices (N, M ≤ unique count, typically much smaller than file size). This gives a sequence of `(oldIdx, newIdx)` pairs — the "anchors".
4. **Recurse between anchors:** For each gap between consecutive anchors in old and new, recursively apply the same algorithm. When no unique lines exist in a sub-range, fall back to Myers for that range.
5. **Emit edits:** Walk the anchor list and recurse to produce the final `[]edittype.Edit` sequence.

**Key property:** The LCS is computed only over unique lines, so the algorithm degrades gracefully — if there are no unique lines in a sub-range, Myers handles it.

**From `peter-evans/patience` source (verified):**

```go
// Step 1: head trim
i := 0
for i < len(a) && i < len(b) && a[i] == b[i] { i++ }
// Step 2: tail trim  
j := 0
for j < len(a) && j < len(b) && a[len(a)-1-j] == b[len(b)-1-j] { j++ }
// Step 3: unique elements for each side
ua, idxa := uniqueElements(a)  // returns strings + original indices
ub, idxb := uniqueElements(b)
// Step 4: LCS of unique elements
lcs := LCS(ua, ub)  // O(N·M) over unique sets
// Step 5: translate back to original indices, recurse between anchors
```

**LCS for anchors** uses a standard O(N·M) DP table — acceptable because N/M are bounded by the number of unique lines (much smaller than full file size).

**Fallback trigger:** When `len(lcs) == 0` — i.e., there are no unique lines common to both sides — the current sub-range becomes a simple all-delete + all-insert block. This is the de facto Myers fallback (Myers produces the same result for all-different sub-ranges). For a proper recursive implementation, Myers is called explicitly on sub-ranges with no unique anchors.

**Implementation note for drift:** The existing `myers.New().Diff()` call already produces an `[]edittype.Edit` slice. Patience will call Myers on sub-ranges and splice the results together, re-numbering `OldLine`/`NewLine` to be relative to the full file (Myers produces 1-indexed lines relative to the sub-slice, so we need an offset correction).

---

### 1.2 Histogram Diff

**Core idea:** Git's Histogram diff (Shawn Pearce, jgit 2010) is a non-LCS algorithm that finds the "best matching region" — the longest contiguous block of matching lines with the *lowest occurrence count* in file A. It is iteratively applied, splitting the region into before/after pieces which are pushed onto a stack. The result tends to "clump" changes rather than spread them.

**Step-by-step algorithm:**

1. **Build occurrence histogram:** For each distinct line in the current A range, count how many times it appears. Store in a hash map: `line → count`.
2. **Scan B lines for candidate matches:** For each line in B (in order), find all positions in A where that line occurs. If a line's occurrence count exceeds `lowcount` (initially 65 per jgit), skip it.
3. **Widen each match:** For each candidate match at `(aIdx, bIdx)`, extend in both directions — check `a[aIdx-1] == b[bIdx-1]`, `a[aIdx+1] == b[bIdx+1]`, etc. — to get the longest contiguous matching region.
4. **Track the best match:** The "best" is the *longest* match, with ties broken by *lowest occurrence count* of any line in the A range of the match. Update `lowcount` to the lowest count seen so far.
5. **Recurse:** Push `after_match` then `before_match` onto the region stack (this ordering ensures the diff output comes out in file order).
6. **Fallback:** If no match is found (all A lines have count > `lowcount`), the entire current region is a change block (Delete all A lines, Insert all B lines). This is also when Myers fallback is triggered in the jgit implementation.

**Pseudocode (from raygard's analysis of jgit):**

```
push region(all of file A and file B) on region stack

while region stack non-empty:
    current_region = pop()
    best_match = find_best_matching_region(current_region)
    if best_match is empty:
        emit current_region as diff (Delete A, Insert B)   ← Myers fallback point
    else:
        if after_match non-empty: push after_match
        if before_match non-empty: push before_match

find_best_matching_region(region):
    set lowcount = 65
    for each B line i in region:
        for each matching A line j (where count(j) <= lowcount):
            widen match from (j, i) to largest extent within region
            compute region_lowcount = min count of any A line in this match
            if match is longer than best OR region_lowcount < lowcount:
                update best_match
                update lowcount = region_lowcount
    return best_match
```

---

## 2. Histogram Frequency Cutoff Decision

**The threshold question: 65 or 512?**

From raygard's authoritative analysis (verified at raygard.net/2025/01/28/how-histogram-diff-works and raygard.net/2025/01/29/a-histogram-diff-implementation/):

- **jgit (and git) use 65** — the original threshold chosen by Shawn Pearce. This is the production threshold used in `git diff --histogram`.
- **Raygard's standalone implementation uses 512** — a more permissive threshold that reduces how often Myers fallback triggers on moderately repetitive files.
- **The cap's purpose:** Limits the hash chain search to O(len_A + len_B). Without a cap, pathological inputs (e.g., a,b,c,a,b,c,… vs c,b,a,c,b,a,…) approach O(len_A × len_B × len_A) cubic behavior.

**Decision for drift: Use 64 (matching jgit's effective behavior)**

Rationale:
- **Correctness invariant is unaffected:** The `apply(diff(a,b), a) == b` invariant holds regardless of threshold — both values produce correct output.
- **Compatibility with user expectations:** Users invoking `drift.WithAlgorithm(drift.Histogram)` have likely used `git diff --histogram` and expect similar behavior. Using jgit's 65 matches git's behavior most closely.
- **Conservative first:** The threshold can be raised later as an option (`WithHistogramMaxOccurrences(n)`) without breaking the public API. Starting low is safer for v1.
- **Exported as unexported constant:** `const histogramMaxOccurrences = 64` in `internal/algo/histogram/histogram.go`. Not exposed in the public API for v1.

**Myers fallback trigger for Histogram:** When `find_best_matching_region` returns an empty match (all A lines in the current region have occurrence count > `histogramMaxOccurrences`). This is a recursive fallback — Myers is called on just that sub-region, not the whole file.

---

## 3. Fallback Trigger Conditions

### Patience Fallback Triggers

| Condition | What Triggers | What Happens |
|-----------|---------------|--------------|
| No unique lines in a sub-range | `len(lcs) == 0` after `uniqueElements()` | Entire sub-range treated as Delete+Insert (functionally identical to Myers minimal diff on that range) |
| Recursion bottoms out at 0-line range | Either side is empty | Return all-Insert or all-Delete edits |
| Single-line sub-range | Head/tail trim consumes everything | Return Equal edit, no recursion |

**Note:** The peter-evans/patience reference returns `append(toDiffLines(a, Delete), toDiffLines(b, Insert)...)` when `len(lcs) == 0`. For drift, this should be replaced by `myers.New().Diff(aSubRange, bSubRange)` to get a minimal edit sequence for the fallback region rather than a simplistic delete-all/insert-all.

### Histogram Fallback Triggers

| Condition | What Triggers | What Happens |
|-----------|---------------|--------------|
| All A lines exceed `lowcount` (64) | No match found in `find_best_matching_region` | Myers called on the current region |
| Empty A range | Region has no A lines | Emit all B lines as Insert |
| Empty B range | Region has no B lines | Emit all A lines as Delete |
| Both ranges empty | Region is empty | No output, pop next region |

**Critical implementation note:** The Myers fallback in Histogram must be called with correct 0-indexed sub-slices and offsets applied to the resulting `OldLine`/`NewLine` values before emitting the edits.

---

## 4. Existing Codebase Analysis

### The `algo.Differ` Interface

```go
// internal/algo/algo.go
type Differ interface {
    Diff(oldLines, newLines []string) []edittype.Edit
}
```

**Contract:** Takes full `[]string` slices (already split by `drift.splitLines`). Returns `[]edittype.Edit` where:
- `Op`: `edittype.Equal`, `edittype.Insert`, or `edittype.Delete`
- `OldLine`: 1-indexed old file line number; 0 for Insert
- `NewLine`: 1-indexed new file line number; 0 for Delete
- Order: edits must be returned in forward file order

Both Patience and Histogram must satisfy this interface. The Myers struct already has a compile-time check:
```go
var _ algo.Differ = (*Myers)(nil)
```
Patience and Histogram will have the same compile-time assertion.

### The `drift.algoInterface` (local in `drift.go`)

The root package defines a local `algoInterface` (identical to `algo.Differ`) to avoid the import cycle:
```go
type algoInterface interface {
    Diff(oldLines, newLines []string) []Edit
}
```
The `switch cfg.algorithm` block in `drift.go` currently falls through Patience/Histogram to Myers. Phase 2 replaces this with real implementations — just add the imports and cases.

### The Edit Offset Problem

Both Patience and Histogram recursively call Myers on sub-ranges. Myers returns 1-indexed line numbers relative to the sub-slices it receives. After Myers returns, the calling code must add the sub-range offset:

```go
// Example: Myers called on oldLines[5:10], newLines[7:12]
// Myers returns OldLine=1 meaning oldLines[5] → must add offset 5
// Myers returns NewLine=2 meaning newLines[8] → must add offset 7
for i := range edits {
    if edits[i].OldLine > 0 { edits[i].OldLine += oldOffset }
    if edits[i].NewLine > 0 { edits[i].NewLine += newOffset }
}
```

This offset correction is a critical correctness requirement for both algorithms.

### DiffResult and Hunk Builder

The hunk builder (`internal/hunk/hunk.Build`) accepts `[]edittype.Edit` without knowing which algorithm produced them. It works on edit-sequence indices, not line numbers. Both Patience and Histogram output is consumed by the same hunk builder unchanged — no changes to `internal/hunk/` are needed.

### The `config` Struct and `WithAlgorithm()`

The `config.algorithm` field is already typed as `drift.Algorithm` (an `int` iota). `WithAlgorithm()` already exists:
```go
func WithAlgorithm(a Algorithm) Option {
    return func(c *config) { c.algorithm = a }
}
```
The `drift.go` dispatch switch just needs real implementations:
```go
case Patience:
    differ = patience.New()
case Histogram:
    differ = histogram.New()
```

---

## 5. Recommended Implementation Approach

### 5.1 Patience Implementation (`internal/algo/patience/`)

**Package structure:**
```
internal/algo/patience/
  patience.go        ← Patience struct, New(), Diff()
  patience_test.go   ← Table-driven + property tests
```

**Key design decisions:**

1. **Iterative recursion via explicit stack** — avoid actual Go recursion to prevent stack overflow on large files with many gaps. Use a `[]subRange` stack where each entry holds `(oldStart, oldEnd, newStart, newEnd)`.

2. **Sub-range dispatch:** Each sub-range is processed: find unique elements in `old[oldStart:oldEnd]` and `new[newStart:newEnd]`, compute LCS, split into anchors + gaps, push gaps back, emit anchors as Equal edits.

3. **Offset tracking:** Carry `oldOffset` and `newOffset` with each stack frame. All emitted edits have `OldLine = localIdx + oldOffset + 1` (converting from 0-indexed local to 1-indexed global).

4. **Myers fallback for no-unique sub-ranges:** Call `myers.New().Diff(oldLines[oldStart:oldEnd], newLines[newStart:newEnd])` and apply offsets before appending to output.

5. **Output ordering:** Process sub-ranges in file order. Since we push gaps in reverse (push right-gap first, then left-gap) and pop from the end, the stack gives left-to-right processing order.

**Simplified skeleton:**

```go
type Patience struct{}

func New() *Patience { return &Patience{} }

func (p *Patience) Diff(old, new []string) []edittype.Edit {
    // Handle edge cases (same as Myers)
    // ...
    
    type frame struct{ os, oe, ns, ne int }
    stack := []frame{{0, len(old), 0, len(new)}}
    edits := make([]edittype.Edit, 0, len(old)+len(new))
    m := myersDiffer  // reuse Myers for fallback
    
    for len(stack) > 0 {
        f := stack[len(stack)-1]
        stack = stack[:len(stack)-1]
        
        // Trim head equals
        // Trim tail equals (emit as Equal edits)
        // Find unique lines in old[f.os:f.oe] and new[f.ns:f.ne]
        // Compute LCS of unique lines
        // If no LCS: call Myers fallback with offsets
        // Else: push gaps, emit anchor Equal edits
    }
    // Sort edits by OldLine/NewLine and return
}
```

### 5.2 Histogram Implementation (`internal/algo/histogram/`)

**Package structure:**
```
internal/algo/histogram/
  histogram.go       ← Histogram struct, New(), Diff(), findBestMatch()
  histogram_test.go  ← Table-driven + property tests
```

**Key design decisions:**

1. **Iterative stack, not recursive:** Same rationale as Patience. Each stack frame is `(oldStart, oldEnd, newStart, newEnd)`.

2. **Occurrence map per frame:** For each frame, build `map[string]int` counting occurrences of each A-range line. This is `O(oe - os)` per frame.

3. **Best-match finding:** Scan B lines, find A matches, widen, track best. The widening loop must respect the frame boundaries (not wander outside `[os, oe)` and `[ns, ne)`).

4. **Myers fallback:** When no match is found (all A lines exceed `histogramMaxOccurrences = 64`), call Myers on the frame sub-range with offsets.

5. **Offset handling:** Same as Patience — `OldLine = localIdx + os + 1`, `NewLine = localIdx + ns + 1`.

**Simplified skeleton:**

```go
const histogramMaxOccurrences = 64

type Histogram struct{}

func New() *Histogram { return &Histogram{} }

func (h *Histogram) Diff(old, new []string) []edittype.Edit {
    type frame struct{ os, oe, ns, ne int }
    stack := []frame{{0, len(old), 0, len(new)}}
    edits := make([]edittype.Edit, 0, len(old)+len(new))
    m := myersDiffer
    
    for len(stack) > 0 {
        f := stack[len(stack)-1]
        stack = stack[:len(stack)-1]
        
        if f.oe == f.os && f.ne == f.ns { continue }
        if f.oe == f.os {
            // Emit all B lines as Insert
            continue
        }
        if f.ne == f.ns {
            // Emit all A lines as Delete
            continue
        }
        
        // Build occurrence histogram for A range
        counts := buildCounts(old[f.os:f.oe])
        
        // Find best matching region
        match, found := findBestMatch(old, new, f, counts)
        if !found {
            // Myers fallback
            sub := m.Diff(old[f.os:f.oe], new[f.ns:f.ne])
            appendWithOffset(sub, f.os, f.ns, &edits)
            continue
        }
        
        // Push after_match, before_match onto stack
        // Emit match.len Equal edits
    }
    return edits
}
```

### 5.3 Wiring `WithAlgorithm()` in `drift.go`

The switch statement in `drift.go` needs two new imports and two new cases. No changes to `DiffResult`, `hunk.Build`, or `options.go` — those are already correct.

```go
import (
    "github.com/tbcrawford/drift/internal/algo/histogram"
    "github.com/tbcrawford/drift/internal/algo/patience"
    "github.com/tbcrawford/drift/internal/algo/myers"
)

switch cfg.algorithm {
case Patience:
    differ = patience.New()
case Histogram:
    differ = histogram.New()
default: // Myers
    differ = myers.New()
}
```

---

## 6. Test Strategy

### 6.1 Unit Tests: Algorithm-Specific Inputs

These tests prove each algorithm's *distinguishing behavior* — what it does better than Myers.

#### Patience Superiority Test (function-move canonical example)

The canonical Patience vs Myers example from peter-evans/patience README:

**Old:**
```c
#include <stdio.h>

// Frobs foo heartily
int frobnitz(int foo) {
    int i;
    for(i = 0; i < 10; i++) {
        printf("Your answer is: ");
        printf("%d\n", foo);
    }
}

int fact(int n) {
    if(n > 1) {
        return fact(n-1) * n;
    }
    return 1;
}

int main(int argc, char **argv) {
    frobnitz(fact(10));
}
```

**New:** `fib` function added, `fact` removed, `printf("Your answer is: ")` removed.

**Expected Patience behavior:** Anchors on `// Frobs foo heartily`, `int frobnitz(int foo) {`, `int main(...)` — unique lines. Myers misidentifies the structure due to `{` and `}` repetition.

**Assert:** `apply(patience.Diff(old, new), old) == new` AND the diff has fewer cross-context hunk merges than Myers.

#### Histogram Superiority Test (repetitive-line file)

A file where every other line is `}` or `{`:

**Old:**
```go
func A() {
    doThing1()
}

func B() {
    doThing2()
}
```

**New:**
```go
func A() {
    doThing1()
    doThing3()
}

func B() {
    doThing2()
}
```

Histogram correctly identifies `func A() {` as the anchor (lowest occurrence count = 1), producing a clean hunk that only touches the `doThing3()` insertion. Myers may produce a messier diff due to the repeated `}` lines.

**Assert:** `apply(histogram.Diff(old, new), old) == new` AND the hunk boundaries correctly contain only the insertion.

#### Fallback Test: All-Identical Lines (maximum repetition)

Input where every A line appears > 64 times:

```go
old := make([]string, 200)
new := make([]string, 200)
for i := range old { old[i] = "}" }  // 200 identical lines
for i := range new { 
    if i == 100 { new[i] = "// inserted" } else { new[i] = "}" }
}
```

**Assert:**
- `histogram.Diff(old, new)` does not panic
- `apply(histogram.Diff(old, new), old) == new`
- The result is correct (Myers fallback fires)

#### Patience Fallback Test: No Unique Lines

```go
old := []string{"a", "a", "a", "b", "b"}
new := []string{"b", "b", "a", "a", "a"}
```
No line is unique in both files. Patience must fall back gracefully.

**Assert:** `apply(patience.Diff(old, new), old) == new`

### 6.2 Invariant Tests (same pattern as Myers)

Replicate the Phase 1 invariant structure for both algorithms:

| Test | What | How |
|------|------|-----|
| `TestBothEmpty` | `[]string{}, []string{}` returns `[]Edit{}` | Same as Myers test |
| `TestOldEmptyAllInserts` | All Insert edits, correct 1-indexed NewLine | Same as Myers test |
| `TestNewEmptyAllDeletes` | All Delete edits, correct 1-indexed OldLine | Same as Myers test |
| `TestLineInvariant` | `Equal+Delete == len(old)`, `Equal+Insert == len(new)` | Table-driven, ~8 cases |
| `TestRoundTrip_Patience` | `apply(diff(a,b),a) == b` for paper example | Manual test |
| `TestRoundTrip_Histogram` | Same for repetitive example | Manual test |

### 6.3 Property-Based Tests (`drift_property_test.go` extension)

The existing `TestProperty_RoundTrip` only tests Myers (via `drift.Diff` default). Add two more rapid properties:

```go
func TestProperty_RoundTrip_Patience(t *testing.T) {
    rapid.Check(t, func(t *rapid.T) {
        // same generator as TestProperty_RoundTrip
        result, _ := drift.Diff(oldText, newText, drift.WithAlgorithm(drift.Patience))
        // same round-trip assert
    })
}

func TestProperty_RoundTrip_Histogram(t *testing.T) {
    rapid.Check(t, func(t *rapid.T) {
        result, _ := drift.Diff(oldText, newText, drift.WithAlgorithm(drift.Histogram))
        // same round-trip assert
    })
}
```

These run 1000 random inputs through each algorithm with the full round-trip invariant. Any offset error or incorrect fallback will surface here.

### 6.4 Cross-Algorithm Consistency Test

For inputs where all three algorithms must produce equivalent `apply()` results (even if the diffs look different), verify this programmatically:

```go
func TestAllAlgorithmsCorrect(t *testing.T) {
    // For a fixed set of test inputs, all three algorithms satisfy apply()==new
    cases := []struct{ old, new string }{...}
    algos := []drift.Algorithm{drift.Myers, drift.Patience, drift.Histogram}
    for _, tc := range cases {
        for _, algo := range algos {
            result, _ := drift.Diff(tc.old, tc.new, drift.WithAlgorithm(algo))
            got := testdata.Apply(result, canonicalLines(tc.old))
            assert got == canonicalLines(tc.new)
        }
    }
}
```

---

## 7. Validation Architecture

### Correctness Validation Pyramid

```
Level 4: Property-based (rapid)
  ├── TestProperty_RoundTrip_Patience    (1000 random inputs × Patience)
  └── TestProperty_RoundTrip_Histogram  (1000 random inputs × Histogram)

Level 3: Behavioral correctness
  ├── TestPatienceSuperiority_FunctionMove   (canonical moved-block example)
  ├── TestHistogramSuperiority_RepetitiveLines  (repetitive-line example)
  └── TestAllAlgorithmsCorrect               (same inputs → same apply() result)

Level 2: Fallback correctness
  ├── TestHistogramFallback_AllIdenticalLines  (200 identical lines → Myers)
  ├── TestPatienceFallback_NoUniqueLines       (no unique common lines)
  └── TestFallback_NoPanic                     (no panics on pathological input)

Level 1: Unit / invariant
  ├── TestBothEmpty, TestOldEmpty, TestNewEmpty  (edge cases, both algos)
  ├── TestLineInvariant_Patience                (Equal+Delete==len(old), etc.)
  ├── TestLineInvariant_Histogram               (same)
  └── TestOffsetCorrectness                     (sub-range offsets are correct)
```

### Validation Approach per Requirement

| Requirement | Validation Method | Evidence |
|-------------|------------------|----------|
| ALGO-01: Patience in `internal/algo/patience/` | Package exists, `var _ algo.Differ = (*Patience)(nil)` compiles | Compile check |
| ALGO-02: Histogram in `internal/algo/histogram/` | Package exists, `var _ algo.Differ = (*Histogram)(nil)` compiles | Compile check |
| ALGO-03: `WithAlgorithm()` selection works | Integration test calling `drift.Diff(a, b, drift.WithAlgorithm(drift.Patience/Histogram))` | Unit test |
| ALGO-04: Both fall back to Myers correctly | `TestHistogramFallback_AllIdenticalLines`, `TestPatienceFallback_NoUniqueLines`, property tests | 1000-run rapid + targeted tests |

### Phase 2 Success Criteria Validation Map

| Success Criterion | Test(s) | How to Verify |
|-------------------|---------|---------------|
| 1. Patience identifies moved blocks that Myers misses | `TestPatienceSuperiority_FunctionMove` | Compare Patience diff hunk count vs Myers hunk count on canonical C example |
| 2. Histogram produces correct hunk boundaries on repetitive inputs | `TestHistogramSuperiority_RepetitiveLines` | Assert hunk boundaries contain only the actual change lines |
| 3. Both algorithms fall back without panic | `TestHistogramFallback_AllIdenticalLines`, `TestPatienceFallback_NoUniqueLines` | No panic + `apply()==new` |
| 4. All three algorithms satisfy `apply(diff(a,b), a) == b` | `TestProperty_RoundTrip_*` (rapid, 1000 runs each) | All 3000 property checks green |

### `go test ./...` Acceptance Bar

Phase 2 is complete when:
```
go test ./...          # all packages pass (including new patience/ and histogram/)
go test -race ./...    # race-clean (stack-based implementations are single-goroutine, no races expected)
go test -run TestProperty ./...  # all 5 property tests pass (3 existing + 2 new)
go vet ./...           # no vet issues
```

---

## 8. Key Risks and Mitigations

| Risk | Likelihood | Mitigation |
|------|-----------|------------|
| Offset arithmetic errors in sub-range recursion | HIGH | `TestOffsetCorrectness` with explicit index assertions; property tests catch all cases |
| Patience LCS O(N·M) too slow for large unique sets | MEDIUM | In practice, unique lines are a small fraction of file size; acceptable for v1; document as known behavior |
| Histogram widening walks outside frame bounds | MEDIUM | Add explicit `>= f.os && < f.oe` bounds check in widening loop |
| Stack overflow on deep recursion | LOW | Iterative stack implementation eliminates this risk entirely |
| All 3 algorithms agree on `apply()` but produce different hunk counts | EXPECTED | This is correct behavior — Patience/Histogram produce different (but valid) diffs than Myers |

---

## 9. Reference Materials

| Source | What It Provides | Verified |
|--------|-----------------|----------|
| `raygard.net/2025/01/28/how-histogram-diff-works/` | Authoritative jgit histogram pseudocode and algorithm explanation | ✓ Fetched |
| `raygard.net/2025/01/29/a-histogram-diff-implementation/` | Threshold decision (65 vs 512), worst-case behavior analysis | ✓ Fetched |
| `github.com/peter-evans/patience` source (`patience.go`, `lcs.go`) | Reference Patience implementation — head/tail trim, unique elements, LCS backtrack | ✓ Fetched |
| `bramcohen.livejournal.com/73318.html` | Bram Cohen's original description of Patience diff | Referenced |
| `jgit HistogramDiff.java` | Original Histogram implementation; threshold=65; chain cap rationale | Referenced |
| Phase 1 codebase (`internal/algo/myers/myers.go`) | Existing Myers implementation that Patience/Histogram will call as fallback | ✓ Read |

---

*Research completed: 2026-03-25*
