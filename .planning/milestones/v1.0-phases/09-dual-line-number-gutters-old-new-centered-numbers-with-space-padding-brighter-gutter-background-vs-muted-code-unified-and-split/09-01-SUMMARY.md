# Plan 09-01 — Summary

**Status:** Complete

## Delivered

- `internal/render/gutter.go`: `gutterWidths`, `centerLineNumber`, `gutterStyle`, `gutterPairWidths`
- `RenderConfig`: `ShowLineNumbers`, `IsDark`
- `Unified()` emits old/new gutters before diff prefix when `ShowLineNumbers`
- `options.go`: `WithLineNumbers`, `WithoutLineNumbers`; default `lineNumbers: true`
- `render.go`: plumbs config into `RenderConfig`
- Tests: `TestUnified_ShowLineNumbersGutter` (and existing tests use `ShowLineNumbers: false` via `noopConfig`)

## Verification

- `go test ./...` passes
