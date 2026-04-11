---
id: 260411-o1x
type: quick
status: completed
date: "2026-04-11"
---

# Quick Task 260411-o1x — Summary

## Objective

Fix `--chrome delta` hunk header box rendering in non-pager (single-file) mode.
`DeltaTheme.RenderHunkHeader` returned `""` because `h.CodeFragment` was never
populated outside pager mode, causing a silent fallback to `@@ -N,n +N,n @@`.

## Root Cause

`edittype.Hunk.CodeFragment` is empty in single-file mode because drift's own
diff algorithm doesn't produce function-context lines — those come from parsing
git's `@@ ... @@ function_name` unified diff output. Only `runPagerMode` had
CodeFragment backfill (by positional index). All other code paths left
`CodeFragment` empty, so `DeltaTheme.RenderHunkHeader` always short-circuited.

## Changes

| File | Change |
|------|--------|
| `cmd/drift/gitworktree.go` | Added `gitHunkFragmentsForFile(path, contextLines)` — runs `git diff --unified=N HEAD -- <path>`, parses `@@` headers via `parseHunkHeader`, returns `newStart→codeFragment` map; nil on any error (best-effort) |
| `cmd/drift/flags.go` | Added `contextLines int` to `rootOptions`; populated from `flags.context` in `resolveRootOptions` |
| `cmd/drift/main.go` | Single-file path in `runRoot`: after `drift.Diff()`, before render, calls `gitHunkFragmentsForFile` and backfills `result.Hunks[i].CodeFragment` by matching `h.NewStart` (not positional index) |
| `cmd/drift/main_test.go` | Added `TestRunCLI_pagerMode_chromeDelta_codeFragment` — verifies DeltaTheme box chars (`┐`/`• `) appear and `@@ -` fallback is absent when `codeFragment != ""` |
| `internal/chrome/chrome.go` | DeltaTheme `RenderHunkHeader` fix + final phase-28 polish |
| `internal/chrome/chrome_test.go` | Updated tests for chrome theme |
| `internal/render/unified.go` | `HunkHeaderRenderer` field in `RenderConfig`; loop emits custom header when non-empty |
| `internal/render/split.go` | Same `HunkHeaderRenderer` integration |
| `options.go` | `WithHunkHeaderRenderer` public option |
| `render.go` | Wires `HunkHeaderRenderer` into `rcfg` |

## Verification

- 407 tests pass (`go test ./... -count=1`)
- `go vet ./...` clean
- Committed atomically: `fix(chrome): backfill CodeFragment in single-file mode so DeltaTheme renders hunk header box`

## Key Design Decisions

- **`NewStart` match over positional index**: `NewStart` is the source-of-truth
  line number from both git's unified diff and drift's hunk computation. Positional
  match is fragile when hunk counts differ (context mismatch).
- **`exec.Command` over go-git**: go-git does not expose function-context lines
  from unified diff headers — those require git's xdiff diff driver. Best-effort
  exec call matches the pager-mode pattern (nil return on any error).
- **Guard `len(opts.args) == 1`**: Only backfill for single git file path. The
  two-file diff path has no git context to query.
