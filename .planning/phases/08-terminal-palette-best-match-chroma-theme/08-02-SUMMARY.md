---
phase: 08-terminal-palette-best-match-chroma-theme
plan: "08-02"
subsystem: rendering
tags: [osc4, tty, cli]

requires:
  - phase: 08-01
    provides: BestMatchTheme, ParseOSC4Responses
provides:
  - Unix OSC 4 query via /dev/tty + raw mode
  - resolveChromaStyle in Render / RenderWithNames
  - WithThemeResolved CLI hook + hidden --show-theme
  - README terminal palette paragraph
affects: []

tech-stack:
  added: [golang.org/x/term]
  patterns:
    - "OSC 4 only when color profile is TTY-capable and writer is *os.File; piped stdout skips query"

key-files:
  created:
    - internal/terminal/palette_unix.go
    - internal/terminal/palette_other.go
    - internal/terminal/palette_other_test.go
  modified:
    - render.go
    - options.go
    - cmd/drift/main.go
    - README.md
    - go.mod
    - go.sum

key-decisions:
  - "WithThemeResolved callback — minimal API surface vs duplicating theme resolution in CLI"

patterns-established:
  - "Non-Unix QueryANSIPalette returns ErrNotSupported"

requirements-completed: [REND-04, REND-08]

duration: 25min
completed: 2026-03-26
---

# Phase 08: terminal palette — Plan 08-02 Summary

**OSC 4 auto-theme on Unix TTYs:** `internal/terminal` queries `/dev/tty` with 500ms timeout; `resolveChromaStyle` wires `BestMatchTheme` when explicit theme and no-color/NoTTY paths allow; CLI `--show-theme` prints resolved name via `WithThemeResolved`.

## Task Commits

1. **08-02 (combined tasks)** — `1759f3f` (feat: OSC4 TTY palette query + auto Chroma theme + --show-theme)

## Files Created/Modified

- `internal/terminal/palette_unix.go` — `QueryANSIPalette` batch OSC 4
- `internal/terminal/palette_other.go` — `ErrNotSupported` stub
- `internal/terminal/palette_other_test.go` — `!unix` builds assert `ErrNotSupported`
- `render.go` — `resolveChromaStyle`, `autoThemeName`
- `options.go` — `WithThemeResolved`
- `cmd/drift/main.go` — `--show-theme` flag
- `README.md` — syntax theme / OSC 4 paragraph

## Self-Check: PASSED

- `go test ./... -count=1` exits 0
