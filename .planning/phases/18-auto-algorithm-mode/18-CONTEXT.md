# Phase 18: Auto Algorithm Mode — Context

**Gathered:** 2026-04-01
**Status:** Ready for planning

<domain>
## Phase Boundary

This phase adds `Auto` as a fourth `Algorithm` constant and makes it the new
default for `drift.Diff()`. `Auto` selects between Myers and Histogram at
diff-time using an O(N) heuristic scan over old-side lines:

- If `len(oldLines) + len(newLines) > 2000`: use Myers (Histogram O(N²) risk
  at scale)
- Else if any old-side line appears > 32 times: use Myers (high-frequency
  lines would cause Histogram to fall back internally anyway)
- Otherwise: use Histogram (small, clean file — quality wins, cost acceptable)

Patience is explicitly excluded from Auto selection: it is never faster than
Myers and provides quality benefits only in very narrow niches.

Research findings are fully documented in `.planning/research/AUTO-ALGORITHM.md`.

This is a single-plan phase. All existing tests must pass after the plan.

</domain>

<decisions>
## Implementation Decisions

### Algorithm Constant

- **D-01:** `Auto` is added as a 4th constant after `Histogram` in `options.go`.
  The iota order is `Myers=0, Patience=1, Histogram=2, Auto=3`.

- **D-02:** `Auto` becomes the new default in `defaultConfig()`. Callers who
  explicitly pass `WithAlgorithm(Myers)` (or any other algorithm) are
  unaffected.

- **D-03:** The `Auto` constant value is 3 (next iota). Existing code that
  used the implicit default of 0 (Myers) is now using 3 (Auto) after this
  change — this is intentional and is a behavior change, not an API break.

### selectAuto() Heuristic

- **D-04:** `selectAuto(old, new []string) algoInterface` lives in `drift.go`
  alongside the `Diff()` dispatch switch. It is unexported (lowercase).

- **D-05:** Thresholds from the research doc:
  - `smallFileThreshold = 2000` (total lines: old+new)
  - `maxFreqThreshold   = 32`   (max occurrences of any single old-side line)

- **D-06:** The frequency scan is O(N) over old-side lines only (mirrors what
  Histogram does internally). For files > 2000 total lines, the scan is
  skipped entirely (zero overhead path).

- **D-07:** The switch case for `Auto` calls `selectAuto(oldLines, newLines)`.
  The existing `default` case remains Myers to protect against invalid
  Algorithm values.

### CLI Changes

- **D-08:** `parseAlgorithm()` in `cmd/drift/main.go` must accept `"auto"` and
  return `drift.Auto`.

- **D-09:** The `--algorithm` flag default value changes from `"myers"` to
  `"auto"`. The help string is updated to list `auto` as an option.

- **D-10:** The flag help string reads:
  `"diff algorithm: auto, myers, patience, histogram"` (auto listed first as
  it is the default).

### Test Coverage

- **D-11:** Integration tests in `drift_algorithm_integration_test.go` get:
  - `TestWithAlgorithm_Auto_RoundTrip` — verifies `apply(diff(a,b), a) == b`
    with `WithAlgorithm(drift.Auto)`.
  - `TestAuto_SelectsHistogram_SmallCleanFile` — verifies that on a small file
    with no high-frequency lines, Auto produces the same result as Histogram.
  - `TestAuto_SelectsMyers_LargeFile` — constructs a synthetic > 2000 line
    input and verifies Auto produces correct output (round-trip).
  - `TestAuto_SelectsMyers_HighFrequency` — verifies Auto falls back to Myers
    when a line exceeds the frequency threshold.

- **D-12:** The existing `TestWithAlgorithm_Myers_StillDefault` test must be
  **renamed** or updated: the default is no longer Myers, it is Auto. The test
  should be updated to assert that the default produces a correct round-trip
  (algorithm-agnostic) rather than asserting Myers-specific behavior.

- **D-13:** `TestAllAlgorithmsCorrect` in `drift_algorithm_integration_test.go`
  must be extended to include `drift.Auto` in the `algos` slice.

- **D-14:** The property-based tests in `drift_property_test.go` must include
  `drift.Auto` in their algorithm list.

### Documentation

- **D-15:** The `Auto` constant godoc in `options.go` reads:
  `// Auto selects Myers or Histogram based on file size and line-frequency analysis.`

- **D-16:** The `WithAlgorithm` godoc is updated to reference all four
  algorithm values: `// WithAlgorithm sets the diff algorithm (Myers, Patience, Histogram, or Auto).`

- **D-17:** `doc.go` — the algorithms section is updated to document `Auto` and
  note that it is the default.

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before implementing.**

### Research (primary specification for this phase)
- `.planning/research/AUTO-ALGORITHM.md` — Full analysis, benchmarks, proposed
  heuristic, threshold calibration, and implementation pseudocode.

### Files to Modify
- `options.go` — Add `Auto` constant; update `defaultConfig()`; update godoc
- `drift.go` — Add `Auto` case in dispatch switch; add `selectAuto()` function
- `cmd/drift/main.go` — Add `"auto"` to `parseAlgorithm()`; update flag default + help
- `drift_algorithm_integration_test.go` — Add Auto tests; update default test
- `drift_property_test.go` — Add `drift.Auto` to algorithm list
- `doc.go` — Update algorithms documentation section

### Architecture constraint
- `drift.go`'s `algoInterface` is defined locally to break the import cycle
  (`internal/algo/myers` imports the root package). `selectAuto()` must return
  `algoInterface`, not a concrete type.

</canonical_refs>

<code_context>
## Existing Code Insights

### Current Algorithm dispatch (drift.go:45-52)
```go
var differ algoInterface
switch cfg.diff.algorithm {
case Patience:
    differ = patience.New()
case Histogram:
    differ = histogram.New()
default: // Myers
    differ = myers.New()
}
```

After this phase the switch becomes:
```go
var differ algoInterface
switch cfg.diff.algorithm {
case Patience:
    differ = patience.New()
case Histogram:
    differ = histogram.New()
case Auto:
    differ = selectAuto(oldLines, newLines)
default: // Myers (and invalid values)
    differ = myers.New()
}
```

### selectAuto() pseudocode (from research doc §7)
```go
func selectAuto(old, new []string) algoInterface {
    const (
        smallFileThreshold = 2000
        maxFreqThreshold   = 32
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

### defaultConfig() current (options.go:47-59)
```go
func defaultConfig() *config {
    return &config{
        diff: diffConfig{
            algorithm:    Myers,
            contextLines: 3,
        },
        ...
    }
}
```

After this phase: `algorithm: Auto`

### parseAlgorithm() current (cmd/drift/main.go:15-26)
```go
switch strings.ToLower(strings.TrimSpace(s)) {
case "myers":    return drift.Myers, nil
case "patience": return drift.Patience, nil
case "histogram": return drift.Histogram, nil
default: return 0, newExitCode(2, fmt.Sprintf("invalid algorithm: %q (use myers, patience, histogram)", s))
}
```

After this phase: add `case "auto": return drift.Auto, nil` and update the
error message to list `auto, myers, patience, histogram`.

### CLI flag definition (cmd/drift/main.go:52)
```go
cmd.Flags().StringVar(&flags.algorithm, "algorithm", "myers", "diff algorithm: myers, patience, histogram")
```

After this phase: default `"auto"`, help `"diff algorithm: auto, myers, patience, histogram"`.

### Test: Myers-as-default (drift_algorithm_integration_test.go:62-77)
The test `TestWithAlgorithm_Myers_StillDefault` calls `drift.Diff()` without
specifying an algorithm and asserts round-trip correctness. Since Auto
subsumes Myers correctness, this test remains valid but should be renamed to
`TestDefault_Algorithm_RoundTrip` and its comment updated to reflect that the
default is now Auto.

</code_context>

<specifics>
## Specific Ideas

- The heuristic cost is bounded: for files > 2000 lines it is O(1) (early
  return), for files ≤ 2000 lines it is one `make(map)` + O(N) scan — cheaper
  than the Myers diff that would follow.

- `selectAuto` does NOT need to be a method on any type. A package-level
  function returning `algoInterface` is the cleanest pattern matching the
  existing codebase style.

- The `Auto` constant being iota=3 means `Algorithm(0)` is still `Myers`. Any
  zero-value `diffConfig{}` would produce Myers (not Auto). This is correct:
  `defaultConfig()` explicitly sets `Auto`, so callers using the functional API
  get `Auto` by default; callers constructing `diffConfig{}` directly get Myers
  (which is acceptable since `diffConfig` is unexported).

</specifics>

<deferred>
## Deferred Ideas

- **Tunable thresholds:** Exposing `WithAutoThresholds(lines, freq int)` as an
  advanced option. Deferred: the hardcoded thresholds are calibrated from
  benchmarks and should be stable. Users who need fine-grained control can use
  `WithAlgorithm(drift.Histogram)` directly.

- **Fix Histogram O(N²) pathology:** A larger change to `findBestMatch` to
  detect high-frequency buckets early and avoid the full scan. Deferred to a
  future phase; this phase works around the pathology with the size+frequency
  gate.

- **Benchmark: Auto selection overhead:** A dedicated benchmark for
  `selectAuto()` call cost across file sizes. Deferred: the research doc
  measured this at 10–30 µs for 1000-line files, which is acceptable.

</deferred>

---

*Phase: 18-auto-algorithm-mode*
*Context gathered: 2026-04-01*
