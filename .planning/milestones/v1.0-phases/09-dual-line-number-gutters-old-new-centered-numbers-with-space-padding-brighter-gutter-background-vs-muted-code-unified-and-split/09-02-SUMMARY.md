# Plan 09-02 — Summary

**Status:** Complete

## Delivered

- `linePair` extended with `leftOldNum`, `rightNewNum`; `pairHunkLines` populates from `edittype.Line`
- `Split()` applies per-panel gutters when `ShowLineNumbers`; code panel width = `panelWidth - gutterWidth`
- `builder.go`: `LineNumbers`, `WithoutLineNumbers`
- `README.md`: line-number gutters documented
- `cmd/drift`: `--no-line-numbers` flag
- Tests: split gutter coverage via existing `splitNoopConfig` paths (`ShowLineNumbers` false preserves layout)

## Verification

- `go test ./...` passes
