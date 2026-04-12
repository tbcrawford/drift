# Quick Task: Fix DeltaTheme Gutter and Chrome Accent Colors

**Date:** 2026-04-11
**Task ID:** 260411-r7k
**Commit:** 3f42406

## What Was Done

Fixed two color issues in the DeltaTheme that were using the hardcoded ANSI-256 index `"63"` (renders as purple on many terminals).

### Issue 1: Gutter line number foreground color

**Problem:** `gutterStyleForCell` in `internal/render/gutter.go` used `lipgloss.Color("63")` for all changed-line gutter cells (delete and insert rows), making numbers appear purplish rather than matching the line's diff color.

**Fix:** Changed the foreground color to use `WordSpanBackgroundColour(style, isDark, isDelete)` as the gutter number foreground. This gives a semantically-colored, brighter red (for delete rows) or green (for insert rows) that visually matches the line highlight color family. Context (Equal) rows keep the existing dim gray foreground.

Files changed:
- `internal/render/gutter.go` — `gutterStyleForCell` function (replaced hardcoded accent with `WordSpanBackgroundColour`-derived hex color)

### Issue 2: Chrome accent color derived from Chroma style

**Problem:** DeltaTheme chrome elements (Δ glyph, filename, rules, hunk header box) used hardcoded `lipgloss.Color("63")` which renders as purple on many terminals instead of blue.

**Fix:**
1. Added `ChromeAccentColor(style *chroma.Style, isDark bool) string` helper to `internal/highlight/diffcolors.go` — queries `chroma.Keyword`, `chroma.NameFunction`, `chroma.LiteralString` token colors (in order) to find a blue-ish color from the active theme. Falls back to `#5f87ff` (bright blue, dark) or `#0050d0` (deep blue, light) when no usable color found.
2. Changed `DeltaTheme` struct in `internal/chrome/chrome.go` to hold an `AccentColor string` field. Added `deltaAccent()` method with zero-value fallback (`#5f87ff`). All `RenderFileHeader` and `RenderHunkHeader` color uses now go through `deltaAccent()`.
3. In `cmd/drift/flags.go`, after resolving isDark and the OSC4 palette theme name, we call `highlight.SelectTheme(resolvedThemeName, resolvedIsDark)` + `highlight.ChromeAccentColor(style, isDark)` and inject the result into `DeltaTheme.AccentColor`. This flows through the `RenderFileHeader`/`RenderHunkHeader` calls at render time.

Files changed:
- `internal/highlight/diffcolors.go` — added `ChromeAccentColor` helper + `"fmt"` import
- `internal/chrome/chrome.go` — `DeltaTheme` struct with `AccentColor` field + `deltaAccent()` method
- `cmd/drift/flags.go` — track `resolvedIsDark`/`resolvedThemeName`, inject accent into DeltaTheme

## Test Results

- `go test ./...`: **414 passed** (all existing tests pass; no new tests added — chrome tests use `DeltaTheme{}` zero value which uses the `#5f87ff` fallback, which is already a correct blue)
- `go vet ./...`: no issues
- `just install`: success
