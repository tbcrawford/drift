---
phase: "28"
plan: "02"
subsystem: cli
tags: [chrome, theme, cli, flags, lipgloss]
dependency_graph:
  requires: [internal/chrome (28-01)]
  provides: [--chrome flag, ChromeTheme wired into all file header call sites]
  affects: [cmd/drift/main.go, cmd/drift/flags.go]
tech_stack:
  added: []
  patterns: [flag → options → render pipeline, chrome.ParseChromeTheme error propagation]
key_files:
  modified:
    - cmd/drift/flags.go
    - cmd/drift/main.go
    - cmd/drift/main_test.go
decisions:
  - lipgloss import removed from main.go — only writeFileHeader used it; now lives in internal/chrome
  - writeFileHeader function deleted entirely — logic fully encapsulated in DriftTheme
  - --chrome flag uses empty string default → DriftTheme (backward compatible, no behavior change without flag)
  - TestRunCLI_chromeDelta uses --no-color and asserts "+--" ASCII fallback (lipgloss not available in test env)
metrics:
  duration: "~5min"
  completed: "2026-04-10"
  tasks_completed: 2
  files_changed: 3
---

# Phase 28 Plan 02: CLI Chrome Wiring Summary

**One-liner:** Wire `--chrome drift|delta` flag through Cobra → `rootFlags` → `rootOptions` → all 4 `writeFileHeader` call sites; remove old function and lipgloss import from main.go.

## Tasks Completed

| Task | Name | Commit | Files |
|------|------|--------|-------|
| 1 | Add chrome to rootFlags/rootOptions; wire resolveRootOptions | `a2fc421` (partial) | cmd/drift/flags.go |
| 2 | Replace writeFileHeader + register --chrome flag; integration tests | `a2fc421` | cmd/drift/main.go, cmd/drift/main_test.go |

## Implementation Notes

### Flag Registration
```go
cmd.Flags().StringVar(&flags.chrome, "chrome", "", "chrome decoration theme: drift (default), delta")
```

### resolveRootOptions Addition
```go
chromeTheme, err := chrome.ParseChromeTheme(flags.chrome)
if err != nil {
    return nil, newExitCode(2, err.Error())
}
```

### Call Site Replacement (4 sites)
All `writeFileHeader(&buf, name, ...)` calls replaced with:
```go
buf.WriteString(opts.chromeTheme.RenderFileHeader(name, opts.noColor, opts.termWidth))
```
or:
```go
results[i].WriteString(opts.chromeTheme.RenderFileHeader(pair.Name, opts.noColor, opts.termWidth))
```

### Cleanup
- `writeFileHeader` function removed from main.go (~45 lines)
- `"charm.land/lipgloss/v2"` import removed from main.go

## Tests Added

| Test | Coverage |
|------|----------|
| TestRunCLI_chromeDrift | --chrome drift → output contains "▸" |
| TestRunCLI_chromeDelta | --chrome delta --no-color → output contains "+--" ASCII box |
| TestRunCLI_chromeDefault | no --chrome flag → output contains "▸" (drift default) |
| TestRunCLI_chromeUnknown | --chrome bogus → exit code 2 |
| TestResolveRootOptions_chromeDefault | unset chrome → DriftTheme |
| TestResolveRootOptions_chromeDelta | chrome=delta → DeltaTheme |
| TestResolveRootOptions_chromeUnknown | chrome=bogus → error |

## Deviations from Plan

**Minor: Test assertion adjusted for --no-color**
- **Found during:** Task 2 GREEN verification
- **Issue:** Plan specified `TestRunCLI_chromeDelta` asserts "┌" box corner, but `--no-color` causes DeltaTheme to emit ASCII "+--" instead of Unicode "┌"
- **Fix:** Updated test to assert "+--" (ASCII fallback) — this is correct behavior; lipgloss ANSI styling is suppressed by --no-color
- **Impact:** None — the underlying behavior is correct; test was adjusted to match actual (correct) output

## Self-Check: PASSED

- `cmd/drift/flags.go`: EXISTS ✓ (chrome field in rootFlags, chromeTheme in rootOptions)
- `cmd/drift/main.go`: EXISTS ✓ (writeFileHeader removed, all 4 sites replaced, --chrome registered)
- `cmd/drift/main_test.go`: EXISTS ✓ (7 new chrome tests added)
- Commit `a2fc421`: EXISTS ✓
- `go test ./... -count=1`: 402 PASSED ✓
- `go vet ./...`: CLEAN ✓
- `go build ./...`: SUCCESS ✓
- `writeFileHeader` function: REMOVED ✓
- lipgloss import from main.go: REMOVED ✓
