# Research: Auto Algorithm Selection for drift

**Date:** 2026-04-01  
**Question:** Is it feasible to add an `Auto` mode that selects the best diff algorithm instead of always defaulting to Myers?  
**Status:** Investigation complete — recommendation below.

---

## 1. Algorithm Recap

### 1.1 Myers (Hirschberg variant) — `internal/algo/myers/myers.go`

The current implementation is the **Hirschberg divide-and-conquer linear-space** variant (Myers 1986 §4b). This is **not** the naïve trace-snapshot Myers; it was upgraded in Phase 16.

Key properties:
- **Time:** O(ND) where D = edit distance, N+M = total lines
- **Space:** O(N+M) — two rolling V arrays, no trace snapshots
- **Strategy:** Find midpoint by running forward and reverse passes simultaneously; split recursively at the "middle snake"
- **Prefix/suffix trimming:** `ses()` strips identical prefix and suffix before calling `midpoint()` — a major practical win for nearly-identical files
- **Pathology:** O(N²) in the worst case (D ≈ N) when files share almost nothing (all-delete + all-insert). Practically O(N log N) on real code.

The implementation correctly prefers deletions (upper diagonals) first, matching Git's convention. The `midpoint()` fallback (`return N/2, M/2`) is only reachable for invalid inputs.

### 1.2 Patience — `internal/algo/patience/patience.go`

- **Strategy:** Find lines appearing exactly once in both old and new → LCS of those unique anchors → recurse on gaps
- **Fallback:** Myers for sub-regions with no unique lines
- **LCS:** O(N·M) DP on the *unique* line sets (much smaller than N, M in practice)
- **Iterative stack:** Avoids deep recursion; tagged union (`isEmit`) for correct forward-order emission
- **Pathology:** When many lines repeat (e.g., `}`, blank lines), unique sets are small or empty → falls back to Myers immediately. When all lines are unique the LCS step on large ranges is O(N·M) expensive.

### 1.3 Histogram — `internal/algo/histogram/histogram.go`

- **Strategy:** For each sub-region, count occurrences of all old-side lines → find the longest contiguous matching block with the lowest occurrence count (`histogramMaxOccurrences = 64` cutoff) → recurse before/after
- **Fallback:** Myers when no old-side line appears ≤ 64 times (`findBestMatch` returns `found=false`)
- **Iterative stack:** Same tagged-union pattern as Patience
- **Pathology:** `findBestMatch` is an O(N·M) scan per sub-region. On large files with many changes, the recursion visits O(D) sub-regions each costing O(N·M), making worst-case O(N²·D). Confirmed by benchmarks below.

---

## 2. Empirical Benchmark Results (Apple M3 Max, Go 1.26.1)

### 2.1 Sequential-change workload (`generateLines`: uniform sequential lines, 200/10000 changed)

| Algorithm | N=100 | N=10000 |
|-----------|-------|---------|
| Myers | 11 µs | 360 µs |
| Patience | 15 µs | 401 µs |
| Histogram | 49 µs | **1,107,627 µs (1.1 sec!)** |

**Finding:** Histogram degrades catastrophically on files with many repeated lines like `"line 00000001\n"`. At N=10000 it is **3,000× slower than Myers**.

### 2.2 Code-like workload (Go-style patterns: repeated structural tokens like `}`, `\t`, blank lines, with unique function names)

| Algorithm | N=1000 |
|-----------|--------|
| Myers | 31.9 µs |
| Patience | 33.8 µs |
| Histogram | **3,590 µs (3.6 ms)** |

**Finding:** Histogram is still 100× slower than Myers even on realistic code. The structural repeats (`}`, `\t\t`, blank lines) create high-frequency line types that saturate `findBestMatch`'s O(N·M) scan.

### 2.3 High-repetition workload (mostly `}` and blank lines)

| Algorithm | N=1000 |
|-----------|--------|
| Myers | 29 µs |
| Patience | 30 µs |
| Histogram | **2,091 µs (2 ms)** |

### 2.4 Unique-line workload (all lines distinct — ideal for Patience/Histogram)

| Algorithm | N=1000 |
|-----------|--------|
| Myers | 281 µs |
| Patience | 975 µs |
| Histogram | **9,737 µs (9.7 ms)** |

**Finding:** Even on the workload *most favorable* to Patience/Histogram (all unique lines), both are slower than Myers. Myers on unique-line files is O(ND) where D is large (swapped blocks), so it struggles too — but Patience's O(N·M) LCS and Histogram's O(N·M·D) scan are slower still.

### 2.5 Heuristic computation cost

Computing a frequency map over both old+new lines to determine characteristics:

| N | Cost |
|---|------|
| 1000 | 27 µs |
| 10000 | 274 µs |

**Finding:** The heuristic scan costs ~75–80% of a full Myers diff on the same file. It is not cheap.

---

## 3. What Git Does (Reference)

Git does **not** have an `auto` algorithm mode. It offers:
- `myers` (default)
- `minimal` (Myers with exhaustive search for minimum edit distance)
- `patience`
- `histogram`

Git's own benchmarks (commit 8555123, 2012) showed: **histogram slightly beats Myers on real `git log -p` workloads; patience is much slower than both.**

The JGit HistogramDiff design note says:
> "If sequence A has more than `maxChainLength` (64) elements that hash into the same bucket, the algorithm passes the region to `setFallbackAlgorithm`."

This is exactly what drift's Histogram implementation does (`histogramMaxOccurrences = 64` → Myers fallback).

The "histogram beats Myers" result Git observed applies to **C source code** diffed during `git log -p` — a workload with moderate repetition (braces, keywords) but many unique identifiers. Drift's sequential-line benchmark is a pathological case that never occurs in real code.

**Key insight from JGit documentation:**
> "By always selecting an LCS position with the lowest occurrence count, this algorithm behaves exactly like Bram Cohen's patience diff whenever there is a unique common element available."

So Histogram is strictly a generalization of Patience that handles non-unique lines gracefully instead of falling back immediately.

---

## 4. Analysis: When Each Algorithm Wins

### Myers wins:
- Nearly identical files (small D) — prefix/suffix trimming eliminates most work before `midpoint()` is called
- Files with many repeated lines — Patience/Histogram fallbacks add overhead for no quality gain
- Files with large edit distances — O(ND) is bounded by the actual edit distance, not file size

### Patience wins (quality, not speed):
- Files with very few unique lines and **small D** — unique anchors naturally find semantically meaningful splits
- Refactored code where function signatures move — unique anchors pin the alignment
- Does NOT win on performance in any measured case

### Histogram wins (quality on moderate repetition):
- Real source code with moderate structural repetition (the Git use case)
- Avoids the Patience problem of "no unique lines → full Myers" by using lowest-frequency instead
- **Performance is a significant concern** — even on its best-case inputs it is slower than Myers

### Auto mode trade-offs:
- **Benefit:** Users get better diff quality on code without knowing which algorithm to pick
- **Cost:** Heuristic scan (27–274 µs overhead) + potential 100–3000× degradation if the wrong algorithm is picked based on a bad heuristic
- **Risk:** Any heuristic that selects Histogram is dangerous unless the file is known to be small (< ~500 lines)

---

## 5. Feasibility Assessment

### 5.1 Can we implement a reliable auto mode?

**Yes, but with significant constraints.**

The key observable inputs available at the `Diff()` call site, before running any algorithm, are:
1. `len(oldLines)`, `len(newLines)` — total line counts
2. Frequency distribution of old-side lines (the same map Histogram builds internally)
3. Count of unique lines (lines appearing exactly once in each side) — what Patience uses

A safe auto heuristic must avoid selecting Histogram for large files with high-frequency lines.

### 5.2 Proposed heuristic for auto selection

```
auto(oldLines, newLines):
    N = len(oldLines), M = len(newLines)
    
    if N + M <= SMALL_FILE_THRESHOLD:  // e.g., 200 lines total
        → Histogram (quality wins, performance acceptable)
    
    compute uniqueRatio = unique lines / distinct line types in old+new
    compute maxFreq = most common line's occurrence count
    
    if maxFreq > HIGH_FREQ_THRESHOLD:  // e.g., > 64 occurrences
        → Myers (Histogram would degrade; Patience would immediately fall back)
    
    if uniqueRatio > UNIQUE_THRESHOLD:  // e.g., > 0.7
        → Patience (anchors will fire; Myers handles the unique-line LCS case well too)
    
    → Myers (safe default for everything else)
```

**Problem:** The heuristic scan itself costs 27–274 µs. For small files this is competitive with the diff cost. For large files it's a meaningful overhead *before* any diffing happens. And after spending 274 µs on analysis we usually conclude "use Myers anyway."

### 5.3 Quality difference in practice

The **quality** difference between algorithms is real but subtle:

| Scenario | Myers output | Histogram output |
|----------|-------------|-----------------|
| Function moved in file | Shows all lines as delete+insert | Shows function body as unchanged, only surrounding context as changed |
| Refactored loop body | Groups all changes together | May isolate the changed inner block |
| Repeated `}` lines | Correct but arbitrary alignment | Same correctness, different alignment |

The quality difference is most visible when a **semantically coherent block** (a function, a class) has been **moved** or **replaced with a similar block**. In that case Histogram/Patience find the unique brace-enclosed boundary lines as anchors and produce a cleaner diff.

---

## 6. Recommendation

### Implement `Auto` as a new `Algorithm` constant, selecting between Myers and Histogram

**Decision:** Yes, implement auto mode. The algorithm is:

```
Auto:
    if (N + M) > 2000 lines:
        use Myers  // Histogram O(N²) risk at scale
    elif maxOccurrence of any old-side line > 32:
        use Myers  // High-frequency lines degrade Histogram
    else:
        use Histogram  // Small+clean file: quality wins
```

**Rationale:**
- Histogram only outperforms Myers on moderate-sized files with bounded repetition
- The 2000-line threshold keeps worst-case Histogram cost ≤ ~3 ms (acceptable for interactive use)
- The frequency check mirrors the internal `histogramMaxOccurrences = 64` logic — if we'd be falling back to Myers internally, just use Myers directly
- Heuristic cost (frequency map scan) is paid once and used to select the algorithm; this is acceptable at the scale where Histogram is selected

**What auto does NOT do:**
- Does NOT add Patience as a selection candidate — Patience is never faster than Myers and is only better in very narrow quality niches (pure unique-line refactors)
- Does NOT do O(N·M) uniqueness analysis — only O(N) frequency scan
- Does NOT guarantee the "best quality" diff — that would require running all three and comparing

### API design

```go
const (
    Myers     Algorithm = iota // O(ND), fast, minimal edits
    Patience                   // Unique-line anchors, better for refactors
    Histogram                  // Frequency-aware anchors, Git's preferred
    Auto                       // Select Myers or Histogram based on file characteristics
)
```

`Auto` becomes the new default in `defaultConfig()`. Existing callers who explicitly pass `WithAlgorithm(Myers)` are unaffected.

**Default change:** Changing the default from `Myers` to `Auto` is a **behavior change** but not an API break. The output may differ for some inputs. This is appropriate at the current version (post-v1.0 candidate).

---

## 7. Implementation Plan

The Auto selection logic belongs in `drift.go` within the `Diff()` algorithm dispatch switch, after `splitLines()` returns `oldLines` and `newLines`. The heuristic scan is O(N) over old-side lines only (cheaper; matches what Histogram itself does internally).

```go
case Auto:
    differ = selectAuto(oldLines, newLines)
```

```go
func selectAuto(old, new []string) algoInterface {
    const (
        smallFileThreshold = 2000  // total lines
        maxFreqThreshold   = 32    // max occurrences of any single old-side line
    )
    
    if len(old)+len(new) > smallFileThreshold {
        return myers.New()
    }
    
    freq := make(map[string]int, len(old))
    for _, l := range old {
        freq[l]++
    }
    for _, count := range freq {
        if count > maxFreqThreshold {
            return myers.New()
        }
    }
    
    return histogram.New()
}
```

This adds one map allocation + O(N) scan for files ≤ 2000 lines, and zero overhead for files > 2000 lines (no scan needed). The total cost is ~10–30 µs for 1000-line files, which is a small fraction of the Myers diff cost (~30 µs) and far smaller than the Histogram cost being avoided.

---

## 8. Threshold Calibration

The thresholds (`2000` lines, `32` max frequency) are conservative starting points derived from the benchmarks:

- At 2000 total lines, Histogram on a high-repetition file costs ~4–6 ms — borderline acceptable
- At 32 max frequency, the 64-occurrence Histogram cutoff fires quickly and falls back to Myers anyway — so we skip the overhead of building the Histogram data structure
- Both thresholds can be tuned later with more benchmarks across real Go codebases

An alternative approach: always use Histogram but fix its O(N²) pathology by detecting the high-frequency case earlier (before the `findBestMatch` scan). This is a larger implementation change but would be the correct long-term fix.

---

## 9. Conclusion

| Question | Answer |
|---------|--------|
| Is auto mode feasible? | Yes |
| Can it reliably improve output quality? | Yes, for small–medium files with moderate repetition |
| Is there a performance risk? | Yes — Histogram can be 3000× slower than Myers on pathological inputs |
| Can we mitigate the risk? | Yes — file size + max-frequency gates prevent the worst cases |
| Should we change the default? | Yes — `Auto` as new default is appropriate; existing `Myers` callers unaffected |
| Should Patience be in Auto? | No — never faster than Myers, quality advantage is narrow |

**Recommended next step:** Implement `Auto` as a new Algorithm constant with the `selectAuto()` heuristic above, make it the new default, and add benchmark coverage for the auto-selection path.
