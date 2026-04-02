---
phase: 23-performance-analysis-and-optimization
plan: "01"
subsystem: render
tags:
  - performance
  - benchmarks
  - gutter
  - lipgloss
dependency_graph:
  requires:
    - internal/render/gutter.go
    - internal/render/unified.go
    - internal/render/split.go
    - benchmark_test.go
  provides:
    - GutterStyleCache type for per-render style caching
    - Color benchmark suite for regression detection
  affects:
    - internal/render/unified.go
    - internal/render/split.go
tech_stack:
  added: []
  patterns:
    - "Pre-compute lipgloss styles once per render call via a keyed map cache (6 keys)"
    - "b.ReportAllocs() in all benchmarks for alloc-aware regression detection"
key_files:
  created: []
  modified:
    - benchmark_test.go
    - internal/render/gutter.go
    - internal/render/unified.go
    - internal/render/split.go
decisions:
  - "GutterStyleCache uses a map[gutterStyleKey]lipgloss.Style with exactly 6 pre-populated entries (2 sides × 3 ops) — exhaustive pre-population avoids map growth and nil checks at usage sites"
  - "GutterCache field added to RenderConfig (not passed as a separate arg) — avoids threading another parameter through all renderer call sites"
  - "Cache built lazily inside Unified/Split (nil check) — allows callers who set cfg.GutterCache externally to inject their own (testability)"
  - "styledGutterColumnSeparator pre-computed as local var before hunk loop in split.go — was called per pair; no interface changes"
metrics:
  duration: "~8 minutes"
  completed: "2026-04-02T19:43:40Z"
  tasks_completed: 2
  files_modified: 4
---

# Phase 23 Plan 01: Gutter Style Cache + Color Benchmarks Summary

**One-liner:** Pre-computed lipgloss.Style cache for gutter cells (6 variants) eliminates per-line lipgloss.NewStyle() allocations in the hot render loop.

## What Was Built

### Task 1: Color Benchmark Suite
Extended `benchmark_test.go` to cover the expensive color rendering path that was previously invisible to benchmarks (all existing benchmarks used `WithNoColor()`):

- `BenchmarkRenderUnified10kColor` — unified view with TrueColor, no no-color flag
- `BenchmarkRenderSplit10kColor` — split view with TrueColor
- `BenchmarkRenderSplitWithLineNumbers10kColor` — split view with line numbers + TrueColor (exercises `gutterStyleForCell` hot path)
- Added `b.ReportAllocs()` to all 6 benchmark functions

**Pre-optimization baseline recorded:**
```
BenchmarkRenderUnified10kColor-14     234    5700141 ns/op    1955308 B/op    52387 allocs/op
BenchmarkRenderSplit10kColor-14       211    5862243 ns/op    3363175 B/op    58522 allocs/op
BenchmarkRenderSplitWithLineNumbers10kColor-14   217    5528098 ns/op    3349645 B/op    58521 allocs/op
```

### Task 2: GutterStyleCache Implementation
Added `GutterStyleCache` to `internal/render/gutter.go`:

```go
type gutterStyleKey struct { oldColumn bool; op edittype.Op }
type GutterStyleCache struct { styles map[gutterStyleKey]lipgloss.Style }

func NewGutterStyleCache(style *chroma.Style, isDark, noColor bool) *GutterStyleCache
func (c *GutterStyleCache) Get(oldColumn bool, op edittype.Op) lipgloss.Style
```

- Pre-populates all 6 style combinations (2 sides × 3 ops: Equal/Delete/Insert)
- Added `GutterCache *GutterStyleCache` field to `RenderConfig`
- Both `unified.go` and `split.go` build the cache once before the hunk loop
- All 4 `gutterStyleForCell()` hot-path calls replaced with `cfg.GutterCache.Get()` lookups
- `styledGutterColumnSeparator` pre-computed as local `gutterSep` var in split.go (was called per pair)

**Post-optimization (Plan 23-01):**
```
BenchmarkRenderUnified10kColor-14     252    5268985 ns/op    1919218 B/op    49995 allocs/op  (-7.5% ns/op, -4.6% allocs/op)
BenchmarkRenderSplitWithLineNumbers10kColor-14   226    5211769 ns/op    3344672 B/op    54690 allocs/op
```

## Deviations from Plan

None — plan executed exactly as written.

## Commits

| Task | Commit | Description |
|------|--------|-------------|
| 1 | 8eadb2b | feat(23-01): add color benchmark functions and b.ReportAllocs() |
| 2 (RED) | (TDD test commit) | test(23-01): add failing tests for GutterStyleCache |
| 2 (GREEN) | 4d5a636 | feat(23-01): cache gutter styles per RenderConfig |

## Known Stubs

None — all implementation complete.

## Self-Check

- [x] benchmark_test.go contains 3 new color benchmark functions
- [x] `b.ReportAllocs()` present in all 6 benchmark functions
- [x] Baseline numbers recorded as comments in benchmark_test.go
- [x] `GutterStyleCache` type exists in `internal/render/gutter.go`
- [x] `RenderConfig` has `GutterCache *GutterStyleCache` field
- [x] Unified and split renderers build cache once per render call
- [x] `styledGutterColumnSeparator` computed once per render call in split.go
- [x] `go test ./...` passes with zero failures
- [x] Golden file tests pass unchanged
