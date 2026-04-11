---
id: 260411-q5c
type: quick
title: Fix DeltaTheme gutter and hunk header box
completed: "2026-04-11T23:03:56Z"
duration: ~15min
tasks_completed: 3
files_modified: 7
commits:
  - f77ed4a
  - 93fed1e
  - 814f99a
key_decisions:
  - DeltaTheme.RenderHunkHeader always renders box — "• N:" when no fragment, "• N: fragment" when present
  - DeltaTheme.GutterSeparators uses literal Unicode ⋮ and │ for both color/noColor paths (no ASCII fallback — characters render fine without ANSI)
  - DriftTheme.GutterSeparators returns (" │", "") preserving existing behavior exactly
  - WithGutterSeparators exported as public API (needed by cmd/drift/flags.go which is in package main importing drift)
  - GutterRightBorder not applied in split mode — panel separator serves same visual role
  - driftGutterSep constant added to chrome package to avoid cross-package import of render/gutter.go constant
tags:
  - chrome
  - delta-theme
  - gutter
  - hunk-header
  - bug-fix
---

# Quick Task 260411-q5c Summary

**One-liner:** Fixed two DeltaTheme visual bugs — hunk header box now always renders (using "• N:" for empty fragments) and gutter now uses delta-style " ⋮ " / " │" separators.

## Tasks Completed

| # | Name | Commit | Files |
|---|------|--------|-------|
| 1 | Fix DeltaTheme.RenderHunkHeader to always render box | f77ed4a | internal/chrome/chrome.go, internal/chrome/chrome_test.go |
| 2 | Add GutterSeparators to ChromeTheme interface and wire through renderers | 93fed1e | internal/chrome/chrome.go, internal/render/unified.go, internal/render/split.go, internal/render/gutter.go, render.go, options.go, cmd/drift/flags.go |
| 3 | Update tests and verify full suite | 814f99a | internal/chrome/chrome_test.go, cmd/drift/main_test.go |

## Changes Made

### Bug 1: Hunk header box never rendered when `codeFragment == ""`

**Root cause:** `DeltaTheme.RenderHunkHeader` had an early-return `if codeFragment == "" { return "" }` that caused the renderer to fall back to the `@@ -N,n +N,n @@` format — even when `--chrome delta` was active.

**Fix:** Removed the early return. When no fragment: content = `• N:`. When fragment present: content = `• N: fragment`. Box rendering is identical for both cases.

### Bug 2: Gutter separator does not match delta's format

**Root cause:** Delta uses `{old} ⋮ {new} │{content}` but drift was emitting `{old} │{new}{content}`.

**Fix:** Added `GutterSeparators(noColor bool) (middleSep, rightBorder string)` to the `ChromeTheme` interface:
- `DriftTheme.GutterSeparators` → `(" │", "")` — no behavioral change
- `DeltaTheme.GutterSeparators` → `(" ⋮ ", " │")`

Wired through: `RenderConfig.GutterMiddleSep` / `GutterRightBorder` → `styledGutterColumnSeparator` → 4 `fmt.Fprintf` calls in `Unified` renderer → `WithGutterSeparators` public option → `flags.go` injection after chrome theme resolution.

`gutterSepWidth` in the unified renderer is now computed dynamically from `cfg.GutterMiddleSep + cfg.GutterRightBorder` widths instead of the hardcoded `const gutterSepWidth = 2`.

## Tests Added

| Test | Package | What It Verifies |
|------|---------|-----------------|
| `TestDeltaTheme_RenderHunkHeader_noFragment` (updated) | internal/chrome | Box renders with "• 10:" when no fragment; colored path has ┐ corner |
| `TestDeltaTheme_RenderHunkHeader_noFragment_colored` (new) | internal/chrome | Box renders in color mode with ┐/┘ corners when no fragment |
| `TestDeltaTheme_GutterSeparators` | internal/chrome | Returns (" ⋮ ", " │") for both color/noColor |
| `TestDriftTheme_GutterSeparators` | internal/chrome | Returns (" │", "") — no regression |
| `TestRunCLI_pagerMode_chromeDelta_noFragment` | cmd/drift | CLI: box renders even without code fragment; no @@ fallback |
| `TestRunCLI_chromeDelta_gutterFormat` | cmd/drift | CLI: ⋮ appears in gutter with --chrome delta |

## Verification

```
go test ./... -count=1   → 412 passed (was 408 before, 4 new tests)
go vet ./...             → no issues
just install             → success
```

## Deviations from Plan

**1. [Rule 1 - Bug] driftGutterSep constant added to chrome package**
- **Found during:** Task 2 Step A
- **Issue:** `DriftTheme.GutterSeparators` needed to return `gutterColumnSeparator` but that constant is in `internal/render/gutter.go` — chrome package cannot import render (circular dependency risk and wrong layering).
- **Fix:** Added `driftGutterSep = " │"` constant in `chrome.go` with a comment noting it matches `gutterColumnSeparator` in gutter.go.
- **Files modified:** internal/chrome/chrome.go
- **Commit:** 93fed1e

**2. [Rule 2 - Missing] DriftTheme GutterSeparators noColor path verified**
- The plan noted DriftTheme only needed one path, but added both `false` and `true` coverage in `TestDriftTheme_GutterSeparators` for completeness.

## Self-Check: PASSED

- [x] internal/chrome/chrome.go — exists, DeltaTheme.RenderHunkHeader no early-return, GutterSeparators on both themes
- [x] internal/render/unified.go — GutterMiddleSep/GutterRightBorder fields, dynamic gutterSepWidth, 4 fmt.Fprintf updated
- [x] internal/render/gutter.go — styledGutterColumnSeparator uses cfg.GutterMiddleSep when non-empty
- [x] internal/render/split.go — comment added for GutterRightBorder
- [x] render.go — GutterMiddleSep/GutterRightBorder wired in both rcfg blocks
- [x] options.go — gutterMiddleSep/gutterRightBorder fields, WithGutterSeparators option
- [x] cmd/drift/flags.go — GutterSeparators called and WithGutterSeparators injected
- [x] Commits f77ed4a, 93fed1e, 814f99a all exist in git log
- [x] 412 tests pass, go vet clean, just install success
