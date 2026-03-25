---
phase: 04-split-rendering
plan: "01"
subsystem: rendering
tags: [lipgloss, split-view, chroma]

requires:
  - phase: 03-unified-rendering
    provides: Unified renderer, RenderConfig, HighlightLine
provides:
  - Split side-by-side renderer with JoinHorizontal
  - TermWidth on RenderConfig
  - pairHunkLines for delete/insert pairing
affects: [04-02, 04-03]

tech-stack:
  added: []
  patterns: [lipgloss two-panel layout with per-line Width styles]

key-files:
  created: [internal/render/split.go, internal/render/split_test.go]
  modified: [internal/render/unified.go]

key-decisions:
  - "Panel widths use (termWidth-3)/2 and remainder on right; min term width 40 in Split only"

patterns-established:
  - "highlightPanel fails open to plain content on HighlightLine error"

requirements-completed: [REND-02]

duration: 15min
completed: 2026-03-25
---

# Phase 4 Plan 04-01 Summary

**Split renderer:** two Lip Gloss panels per hunk with Unicode or ASCII separator, ANSI-aware width via `lipgloss.Width` in tests, and syntax highlighting through existing `HighlightLine`.

## Performance

- **Tasks:** 2
- **Files modified:** 3

## Task Commits

1. **04-01-01** — `5b3e4de` feat(render): TermWidth and Split
2. **04-01-02** — `5355948` test(render): Split tests

## Self-Check: PASSED

- `key-files.created` exist; `go test ./internal/render/... -run TestSplit` passes
