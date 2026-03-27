# Summary 10-01 — Theme-aware full-line diff styling

**Completed:** 2026-03-26

## Delivered

- `internal/highlight/diff_line.go` — theme backgrounds from Chroma gd/gi with blend/fallback; `ApplyDiffLineStyle` downgrades full ANSI resets between tokens
- Unified/split integration; `WithLineDiffStyle` / builder; default on via `defaultConfig` and `RenderConfig`

## Notes

Render tests use `noopConfig()` without `LineDiffStyle` set (zero false) so output stays plain; library `Render()` sets `LineDiffStyle: true` by default.
