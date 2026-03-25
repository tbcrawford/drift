# Plan 03-02 Summary: Chroma v2 Highlight Pipeline

## Status: COMPLETED

## What Was Built

Implemented `internal/highlight/` — the full Chroma v2 per-line syntax highlighting pipeline.

### Files Modified
- `go.mod` / `go.sum` — added `github.com/alecthomas/chroma/v2 v2.23.1` (and `github.com/dlclark/regexp2 v1.11.5` as transitive dep)
- `internal/highlight/highlight.go` — new file, 108 lines
- `internal/highlight/highlight_test.go` — new file, 135 lines

### Functions Exported
| Function | Description |
|---|---|
| `HighlightLine(line, lexer, style, formatter)` | Per-line tokenization and ANSI rendering; fail-open on error |
| `FormatterForProfile(p)` | Maps colorprofile.Profile → correct Chroma formatter (16m/256/16/noop) |
| `SelectTheme(requested, isDark)` | Resolves theme by name or auto-detects monokai (dark) / github (light) |
| `DetectLexer(lang, filename, content)` | Priority-ranked lexer selection with Coalesce wrapping |

## Key Decisions

- **`styles.Get()` returns Fallback, not nil** — Chroma's `styles.Get()` always returns the fallback style ("swapoff") when a name is unknown. `SelectTheme` uses direct `Registry[name]` map lookup to distinguish "found" vs "not found", enabling proper fallback to auto-detected theme.
- **`FormatterFunc` not comparable** — Chroma's `FormatterFunc` type panics on `==` comparison. Tests use behavioral output comparison (same probe text produces same output) rather than pointer identity.
- **colorprofile imported as indirect dep** — `github.com/charmbracelet/colorprofile v0.4.2` was already in go.sum as indirect dep from lipgloss; now used directly by this package.

## Test Results
- `go test ./internal/highlight/...` → PASS
- `go test -race ./internal/highlight/...` → PASS (no data races)

## Commits
1. `feat(deps): add github.com/alecthomas/chroma/v2 v2.23.1 dependency`
2. `feat(highlight): implement Chroma v2 highlight pipeline (HighlightLine, FormatterForProfile, SelectTheme, DetectLexer)`
3. `feat(highlight): add unit tests for Chroma v2 highlight pipeline`
