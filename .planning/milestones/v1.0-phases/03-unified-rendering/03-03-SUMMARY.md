---
plan: "03-03"
status: "complete"
completed_at: "2026-03-25T21:10:00.000Z"
---

# Plan 03-03 Summary: Language auto-detection and WithLang/WithTheme option wiring

## What Was Built

### Task 03-03-01: TestDetectLexer tests (internal/highlight/highlight_test.go)

Added 6 `TestDetectLexer_*` test functions to `internal/highlight/highlight_test.go`:

- `TestDetectLexer_ExplicitLang` — verifies explicit lang="go" returns Go lexer and tokenizes correctly
- `TestDetectLexer_FilenameExtension` — verifies filename="main.go" triggers Go lexer via extension match
- `TestDetectLexer_PythonOverride` — verifies explicit lang="python" returns Python lexer
- `TestDetectLexer_UnknownExtensionFallback` — verifies unknown extension returns non-nil Fallback lexer
- `TestDetectLexer_ContentAnalysis` — verifies Go source content triggers Go lexer via content analysis
- `TestDetectLexer_ExplicitPriorityOverFilename` — verifies explicit lang overrides filename extension

All tests pass: `go test ./internal/highlight/... -run TestDetectLexer` exits 0.

**Key discovery**: The full `internal/highlight` package (highlight.go + test file) was being created in parallel by plan 03-02. This plan's test additions were already included in that commit. The `SelectTheme` fallback chain required a bug fix: `styles.Get()` never returns nil (always returns `styles.Fallback`), so the unknown-theme fallback check was changed to use `styles.Registry` direct map lookup.

### Task 03-03-02: options_test.go — WithLang/WithTheme/WithNoColor config wiring

Created `options_test.go` in the root `drift` package with 3 tests:

- `TestWithLang` — verifies `WithLang("go")` sets `config.lang = "go"`
- `TestWithTheme` — verifies `WithTheme("monokai")` sets `config.theme = "monokai"`  
- `TestWithNoColor_SetsFlag` — verifies `WithNoColor()` sets `config.noColor = true`

All tests pass: `go test . -run "TestWithLang|TestWithTheme|TestWithNoColor_SetsFlag"` exits 0.

## Requirements Addressed

- **REND-06**: `DetectLexer` priority chain fully tested — explicit lang > filename/extension > content analysis > Fallback
- **REND-07**: `WithLang` and `WithTheme` options verified to wire through `config` correctly

## Files Modified

| File | Change |
|------|--------|
| `internal/highlight/highlight.go` | Created (with parallel 03-02) — contains `DetectLexer`, `HighlightLine`, `FormatterForProfile`, `SelectTheme` |
| `internal/highlight/highlight_test.go` | Created (with parallel 03-02) — contains all highlight + DetectLexer tests |
| `options_test.go` | Created — WithLang, WithTheme, WithNoColor option wiring tests |

## Test Results

```
ok  github.com/tbcrawford/drift                     (options_test + drift_test)
ok  github.com/tbcrawford/drift/internal/highlight  (15 tests — all passing)
ok  github.com/tbcrawford/drift/...                 (all packages)
```

## Decisions

- `styles.Get()` returns `Fallback` (never nil) in Chroma v2 — use `styles.Registry` direct map lookup for explicit existence checking in `SelectTheme`
- `FormatterFunc` types are not comparable with `==` in Go — avoid identity comparison; use behavioral verification (ANSI output presence) instead
